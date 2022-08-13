package gui

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var psArgs = GetEnv("PS_ARGS", "pid,ppid,%cpu,%mem,lstart,user,command")

type ProcessManager struct {
	*tview.Table
	pids       *[]PID
	FilterWord string
	procDs     procDataSource
}

func NewProcessManager() *ProcessManager {
	procDs, err := NewProcDataSource()
	if err != nil {
		panic("")
	}

	p := &ProcessManager{
		pids:   nil,
		Table:  tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false),
		procDs: procDs,
	}
	p.SetBorder(true).SetTitle("processes").SetTitleAlign(tview.AlignLeft)
	return p
}

func (p *ProcessManager) GetProcesses() (map[PID]Process, error) {
	procs := p.procDs.GetProcesses(p.FilterWord)

	for _, proc := range procs {
		// skip pid 0
		if proc.Pid == "0" {
			continue
		}
	}

	return procs, nil
}

var headers = []string{
	"Pid",
	"PPid",
	"Cmd",
}

func (p *ProcessManager) UpdateView() error {
	// get processes
	procs, err := p.GetProcesses()
	if err != nil {
		return err
	}

	table := p.Clear()

	// set headers
	for i, h := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            h,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
		})
	}

	// set process info to cell
	var i int
	pids := make([]PID, 0, len(procs))
	for pid := range procs {
		pids = append(pids, pid)
	}

	sort.Slice(pids, func(i, j int) bool {
		a, b := pids[i], pids[j]

		if len(a) == len(b) {
			return a < b
		}
		return len(a) < len(b)
	})

	for _, pid := range pids {
		proc := procs[pid]
		pid := string(proc.Pid)
		ppid := string(proc.PPid)
		table.SetCell(i+1, 0, tview.NewTableCell(pid))
		table.SetCell(i+1, 1, tview.NewTableCell(ppid))
		table.SetCell(i+1, 2, tview.NewTableCell(proc.Cmd))
		i++
	}

	p.pids = &pids

	return nil
}

func (p *ProcessManager) Selected() *Process {
	if p.pids == nil {
		return nil
	}
	row, _ := p.GetSelection()
	if row < 0 {
		return nil
	}
	if len(*p.pids) < row {
		return nil
	}
	focusedPid := (*p.pids)[row-1]
	return p.procDs.GetProcess(focusedPid)
}

func (p *ProcessManager) Kill() error {
	pid := p.Selected().Pid
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

func (p *ProcessManager) KillWithPid(pid PID) error {
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

func (p *ProcessManager) Info(pid PID) (string, error) {
	if pid == "0" {
		return "", nil
	}

	cmd := exec.Command("ps", "-o", psArgs, "-p", pid.String())
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if err := cmd.Run(); err != nil {
		return "", errors.New(string(buf))
	}

	return string(buf), nil
}

func (p *ProcessManager) Env(pid PID) (string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	if pid == "0" {
		return "", nil
	}

	env, err := readProcPath(pid, "environ")
	if err != nil {
		return "", err
	}

	result := strings.Split(env, "\x00")

	var (
		envs []string
	)

	for _, e := range result {
		kv := strings.SplitN(e, "=", 1)
		if len(kv) != 2 {
			continue
		}
		envs = append(envs, fmt.Sprintf("[yellow]%s[white]\t%s", kv[0], kv[1]))
	}

	return strings.Join(envs, "\n"), nil
}

func (p *ProcessManager) OpenFiles(pid PID) (string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	if pid == "0" {
		return "", nil
	}

	buf := bytes.Buffer{}
	cmd := exec.Command("lsof", "-p", pid.String())
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return "", errors.New(buf.String())
	}

	result := strings.SplitN(buf.String(), "\n", 2)
	if len(result) > 1 {
		result[0] = fmt.Sprintf("[yellow]%s[white]", result[0])
	} else {
		return buf.String(), nil
	}

	return strings.Join(result, "\n"), nil
}
