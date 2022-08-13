package gui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type ProcDataSource interface {
	GetAllProcesses() []PID

	// procfs based
	GetChildren(pid PID) []PID
	GetCommand(pid PID) string
	GetEnviron(pid PID) []string

	// ebpf based
	GetExecTrace(pid PID) []ExecData
	GetOpenTrace(pid PID) []OpenData
	GetChdirTrace(pid PID) []ChdirData
}

type procDataSource struct {
	execDs  Datasource[ExecData]
	openDs  Datasource[OpenData]
	openLog map[PID][]OpenData
	chdirDs Datasource[ChdirData]
}

func NewProcDataSource() (procDataSource, error) {
	execDs, err := NewExecDataSource()
	if err != nil {
		return procDataSource{}, err
	}
	openDs, err := NewOpenDataSource()
	if err != nil {
		return procDataSource{}, err
	}
	chdirDs, err := NewChdirDataSource()
	if err != nil {
		return procDataSource{}, err
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

	return procDataSource{
		execDs:  execDs,
		openDs:  openDs,
		openLog: opens,
		chdirDs: chdirDs,
	}, err
}

func (pds *procDataSource) GetAllProcesses() []PID {
	files, err := filepath.Glob("/proc/*")
	if err != nil {
		panic("glob panicked")
	}
	pids := make([]PID, 0, 500)
	for _, f := range files {
		fmt.Println(f)
		candidate := path.Base(f)
		if _, err := strconv.Atoi(candidate); err != nil {
			continue
		}
		pids = append(pids, PID(candidate))
	}
	return pids
}

func (pds *procDataSource) GetChildren(pid PID) []PID {
	files, err := filepath.Glob(path.Join("/proc", pid.String(), "task", "*", "children"))
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
			if _, err := strconv.Atoi(f); err != nil {
				panic("proc/children file format has diverged")
			}
			pids = append(pids, PID(p))
		}
	}
	return pids
}

func readProcPathBytes(pid PID, p string) ([]byte, error) {
	return ioutil.ReadFile(path.Join("/proc", pid.String(), p))
}

func readProcPath(pid PID, p string) (string, error) {
	b, err := ioutil.ReadFile(path.Join("/proc", pid.String(), p))
	if err != nil {
		return "", err
	}
	return string(b), err
}

func (pds *procDataSource) GetCommand(pid PID) string {
	str, _ := readProcPath(pid, "cmdline")
	return str
}

func (pds *procDataSource) GetEnviron(pid PID) []string {
	panic("unimplemented")
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
