package gui

import (
	"sort"

	"github.com/dixler/pst/gui/proc"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type ProcessManager struct {
	*tview.Table
	pids       *[]proc.PID
	FilterWord string
	procDs     proc.ProcDataSource
}

func NewProcessManager() *ProcessManager {
	procDs, err := proc.NewProcDataSource()
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

func (p *ProcessManager) GetProcesses() (map[proc.PID]proc.Process, error) {
	procs := p.procDs.GetProcesses(p.FilterWord)

	procmap := make(map[proc.PID]proc.Process)
	for _, p := range procs {
		// skip pid 0
		if p.Pid == "0" {
			continue
		}
		procmap[p.Pid] = p
	}

	return procmap, nil
}

var headers = []string{
	"Pid",
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
	pids := make([]proc.PID, 0, len(procs))
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

	for i, pid := range pids {
		proc := procs[pid]
		pid := string(proc.Pid)
		table.SetCell(i+1, 0, tview.NewTableCell(pid))
		table.SetCell(i+1, 1, tview.NewTableCell(proc.Cmd))
	}

	p.pids = &pids

	return nil
}

func (p *ProcessManager) Selected() *proc.Process {
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
