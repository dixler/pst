package gui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type PID string

func (p PID) String() string {
	return string(p)
}

func (p PID) Int() int {
	num, err := strconv.Atoi(string(p))
	if err != nil {
		panic("invalid PID '" + string(p) + "'")
	}
	return num
}

type Datasource[T any] struct {
	Get       func(pid PID) (T, error)
	GetStream func() (chan PID, chan T, error)
}

/*
TODO
==================
WARNING: DATA RACE
Write at 0x00c000078120 by goroutine 10:
  runtime.mapassign_faststr()
      /usr/lib/go/src/runtime/map_faststr.go:203 +0x0
  main.NewSource[...].func2()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/datasource.go:61 +0x14e

Previous read at 0x00c000078120 by goroutine 15:
  runtime.mapaccess2_faststr()
      /usr/lib/go/src/runtime/map_faststr.go:108 +0x0
  main.NewSource[...].func3()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/datasource.go:69 +0x65
  main.main.func1()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/main.go:61 +0x24d

Goroutine 10 (running) created at:
  main.NewSource[...]()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/datasource.go:42 +0x7d9
  main.NewExecDataSource()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/exec.go:24 +0x3c
  main.main()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/main.go:21 +0xec

Goroutine 15 (running) created at:
  main.main()
      /home/kdixler/go/src/github.com/dixler/edb/pkg/cmd/main.go:49 +0x4d1
==================
*/

func NewSource[T any](program string,
	process func(line string) (PID, T, error),
	get ...func(pid PID) (T, error),
) (Datasource[T], error) {

	cmd := exec.Command("bpftrace", "-e", program)
	cmd.Env = append(cmd.Env, "BPFTRACE_STRLEN=200")
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err.Error())
	}
	rd := bufio.NewReader(out)

	pidCh := make(chan PID, 500)
	dataCh := make(chan T, 500)
	done := make(chan bool, 1)

	lock := sync.RWMutex{}
	cache := make(map[PID]T)

	go func() {
		err = cmd.Run()
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		<-done
		close(done)
		cmd.Process.Kill()
	}()

	go func() {
		first := true
		for {
			str, err := rd.ReadString('\n')
			if err != nil {
				fmt.Println("Read Error:", err)
				done <- true
				return
			}
			if first {
				first = false
				continue
			}

			func(str string) {
				pid, d, err := process(str)
				if err != nil {
					return
				}

				lock.Lock()
				cache[pid] = d
				lock.Unlock()
				pidCh <- pid
				dataCh <- d
			}(str)
		}
	}()

	return Datasource[T]{
		Get: func(pid PID) (T, error) {
			lock.RLock()
			t, ok := cache[pid]
			lock.RUnlock()
			if !ok && len(get) > 0 {
				return get[0](pid)
			}
			if !ok {
				return t, fmt.Errorf("PID[%s] does not exist", pid)
			}
			return t, err
		},
		GetStream: func() (chan PID, chan T, error) {
			return pidCh, dataCh, err
		},
	}, nil
}
