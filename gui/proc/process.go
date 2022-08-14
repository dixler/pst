package proc

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Process struct {
	Pid   PID
	PPid  PID
	Cmd   string
	Child []PID
}

type ProcDataSource interface {
	GetProcesses(filters ...string) map[PID]Process
	GetProcess(pid PID) *Process

	// ebpf based
	GetExecTrace(pid PID) []ExecData
	GetOpenTrace(pid PID) []OpenData
	GetChdirTrace(pid PID) []ChdirData
}

type procDataSource struct {
	execDs  Datasource[ExecData]
	openDs  Datasource[OpenData]
	chdirDs Datasource[ChdirData]

	openLog       map[PID][]OpenData
	procCacheLock *sync.RWMutex
	procCache     map[PID]ExecData
}

func NewProcDataSource() (*procDataSource, error) {
	execDs, err := NewExecDataSource()
	if err != nil {
		return &procDataSource{}, err
	}
	openDs, err := NewOpenDataSource()
	if err != nil {
		return &procDataSource{}, err
	}
	chdirDs, err := NewChdirDataSource()
	if err != nil {
		return &procDataSource{}, err
	}

	openDsPID, openDsData, err := openDs.GetStream()
	if err != nil {
		os.Exit(1)
	}

	opens := make(map[PID][]OpenData)
	go func() {
		for pid := range openDsPID {
			e := <-openDsData

			d, ok := opens[pid]
			if !ok {
				d = make([]OpenData, 0, 4)
			}
			opens[pid] = append(d, e)
		}
	}()

	pds := procDataSource{
		execDs:        execDs,
		openDs:        openDs,
		openLog:       opens,
		chdirDs:       chdirDs,
		procCache:     make(map[PID]ExecData),
		procCacheLock: &sync.RWMutex{},
	}

	pds.bootstrapProcCache()

	return &pds, nil
}

func (pds *procDataSource) GetProcess(pid PID) *Process {
	pds.procCacheLock.RLock()
	p, ok := pds.procCache[pid]
	pds.procCacheLock.RUnlock()
	if !ok {
		p = ExecData{
			Command: GetCommand(pid),
		}
		pds.procCacheLock.Lock()
		pds.procCache[pid] = p
		pds.procCacheLock.Unlock()
	}
	return &Process{
		Pid:   pid,
		Cmd:   p.Command,
		Child: GetChildren(pid),
	}
}

func (pds *procDataSource) bootstrapProcCache() {
	files, err := filepath.Glob("/proc/*")
	if err != nil {
		panic("glob panicked")
	}
	for _, f := range files {
		candidate := path.Base(f)
		if _, err := strconv.Atoi(candidate); err != nil {
			continue
		}
		pid := PID(candidate)
		pds.procCacheLock.Lock()
		pds.procCache[pid] = ExecData{
			Command: GetCommand(pid),
		}
		pds.procCacheLock.Unlock()
	}

	execDsPID, execDsData, err := pds.execDs.GetStream()
	if err != nil {
		os.Exit(1)
	}

	go func() {
		for pid := range execDsPID {
			e := <-execDsData

			pds.procCacheLock.Lock()
			pds.procCache[pid] = e
			pds.procCacheLock.Unlock()
		}
	}()
}

func (pds *procDataSource) GetProcesses(filters ...string) map[PID]Process {
	pds.procCacheLock.RLock()
	defer pds.procCacheLock.RUnlock()

	results := make(map[PID]Process)
	for pid, p := range pds.procCache {
		if !strings.Contains(p.Command, filters[0]) {
			continue
		}
		results[pid] = Process{
			Pid:   pid,
			Cmd:   p.Command,
			Child: GetChildren(pid),
		}
	}
	return results
}

func GetChildren(pid PID) []PID {
	prefix := path.Join("/proc", pid.String(), "task")
	files, err := filepath.Glob(path.Join(prefix, "*", "children"))

	if err != nil {
		panic("glob panicked")
	}
	pids := make([]PID, 0, 500)
	for _, f := range files {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			continue
		}
		str := strings.TrimSpace(string(b))
		for _, p := range strings.Split(str, " ") {
			if p == "" {
				continue
			}

			if _, err := strconv.Atoi(p); err != nil {
				//fmt.Printf("proc/children file format has diverged got(%s) from '%s'\n", str, f)
				continue
			}
			pids = append(pids, PID(p))
		}
	}
	return pids
}

func (pds *procDataSource) GetExecTrace(pid PID) []ExecData {
	panic("unimplemented")
}
func (pds *procDataSource) GetOpenTrace(pid PID) []OpenData {
	return pds.openLog[pid]
}
func (pds *procDataSource) GetChdirTrace(pid PID) []ChdirData {
	panic("unimplemented")
}

func Kill(pid PID) error {
	proc, err := os.FindProcess(pid.Int())
	if err != nil {
		log.Println(err)
		return err
	}

	if err := proc.Kill(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
