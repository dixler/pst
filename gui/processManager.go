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
	processes  []Process
	FilterWord string
	procDs     procDataSource
}

func NewProcessManager() *ProcessManager {
	procDs, err := NewProcDataSource()
	if err != nil {
		panic("")
	}

	p := &ProcessManager{
		Table:  tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false),
		procDs: procDs,
	}
	p.SetBorder(true).SetTitle("processes").SetTitleAlign(tview.AlignLeft)
	return p
}

func (p *ProcessManager) GetProcesses() (map[PID]Process, error) {
	pids := p.procDs.GetAllProcesses()

	procs := make(map[PID]Process)
	for _, pid := range pids {
		// skip pid 0
		if pid == "0" {
			continue
		}

		procs[pid] = Process{
			Pid:   pid,
			Cmd:   p.procDs.GetCommand(pid),
			Child: p.procDs.GetChildren(pid),
		}
	}

	for _, proc := range procs {
		if strings.Contains(proc.Cmd, p.FilterWord) {
			continue
		}

		p.processes = append(p.processes, proc)
	}

	sort.Slice(p.processes, func(i, j int) bool {
		return p.processes[i].Pid < p.processes[j].Pid
	})

	return procs, nil
}

var headers = []string{
	"Pid",
	"PPid",
	"Cmd",
}

func (p *ProcessManager) UpdateView() error {
	// get processes
	if _, err := p.GetProcesses(); err != nil {
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
	for _, proc := range p.processes {
		pid := string(proc.Pid)
		ppid := string(proc.PPid)
		table.SetCell(i+1, 0, tview.NewTableCell(pid))
		table.SetCell(i+1, 1, tview.NewTableCell(ppid))
		table.SetCell(i+1, 2, tview.NewTableCell(proc.Cmd))
		i++
	}

	return nil
}

func (p *ProcessManager) Selected() *Process {
	if len(p.processes) == 0 {
		return nil
	}
	row, _ := p.GetSelection()
	if row < 0 {
		return nil
	}
	if len(p.processes) < row {
		return nil
	}
	return &p.processes[row-1]
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
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	if pid == "0" {
		return "", nil
	}

	buf := bytes.Buffer{}
	cmd := exec.Command("ps", "-o", psArgs, "-p", pid.String())
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", errors.New(buf.String())
	}

	return buf.String(), nil
}

func (p *ProcessManager) Env(pid PID) (string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	if pid == "0" {
		return "", nil
	}

	buf := bytes.Buffer{}
	cmd := exec.Command("ps", "eww", "-o", "command", "-p", pid.String())
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.Split(buf.String(), "\n")

	var (
		envStr []string
		envs   []string
	)

	if len(result) > 1 {
		envStr = strings.Split(result[1], " ")[1:]
	} else {
		return buf.String(), nil
	}

	for _, e := range envStr {
		kv := strings.Split(e, "=")
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
