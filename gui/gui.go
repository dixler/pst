package gui

import (
	"log"
	"time"

	"github.com/dixler/pst/gui/proc"
	"github.com/rivo/tview"
)

const (
	InputPanel int = iota + 1
	ProcessesPanel
	ProcessInfoPanel
	ProcessEnvPanel
	ProcessTreePanel
	ProcessFilePanel
)

type Gui struct {
	FilterInput     *tview.InputField
	ProcessManager  *ProcessManager
	ProcessInfoView *ProcessInfoView
	ProcessTreeView *ProcessTreeView
	ProcessEnvView  *EnvView
	ProcessFileView *ProcessFileView
	NaviView        *NaviView
	App             *tview.Application
	Pages           *tview.Pages
	updateChannel   chan proc.PID
	Panels
}

type Panels struct {
	Current int
	Panels  []tview.Primitive
	Kinds   []int
}

func New() *Gui {
	filterInput := tview.NewInputField().SetLabel("cmd name:")
	processManager := NewProcessManager()
	processInfoView := NewProcessInfoView()
	processTreeView := NewProcessTreeView(processManager.GetProcesses)
	processEnvView := NewEnvView()
	processFileView := NewProcessFileView()
	naviView := NewNaviView()
	updateChannel := make(chan proc.PID, 50)

	g := &Gui{
		FilterInput:     filterInput,
		ProcessManager:  processManager,
		App:             tview.NewApplication(),
		ProcessInfoView: processInfoView,
		ProcessTreeView: processTreeView,
		ProcessEnvView:  processEnvView,
		ProcessFileView: processFileView,
		NaviView:        naviView,
		updateChannel:   updateChannel,
	}

	redraw := func(pid proc.PID) {
		g.ProcessInfoView.UpdateInfoWithPid(g, pid)
		g.ProcessTreeView.UpdateTree(g, pid)
		g.ProcessEnvView.UpdateViewWithPid(g, pid)
		g.ProcessFileView.UpdateViewWithPid(g, pid)
		g.NaviView.UpdateView(g)
	}

	go func() {
		duration := 250 * time.Millisecond
		t := time.NewTicker(duration)
		var curPid *proc.PID = nil
		var newPid *proc.PID = nil
		for {
			select {
			case <-t.C:
				if newPid == nil {
					continue
				}
				if newPid == curPid {
					continue
				}
				curPid = newPid
				redraw(*newPid)
			case pid := <-g.updateChannel:
				newPid = &pid
				t.Reset(duration)
			}
		}
	}()

	g.Panels = Panels{
		Panels: []tview.Primitive{
			filterInput,
			processManager,
			processInfoView,
			processFileView,
			processEnvView,
			processTreeView,
		},
		Kinds: []int{
			InputPanel,
			ProcessesPanel,
			ProcessInfoPanel,
			ProcessFilePanel,
			ProcessEnvPanel,
			ProcessTreePanel,
		},
	}

	return g
}

func (g *Gui) Confirm(message, doneLabel string, primitive tview.Primitive, doneFunc func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel, "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.CloseAndSwitchPanel("modal", primitive)
			if buttonLabel == doneLabel {
				g.App.QueueUpdateDraw(func() {
					doneFunc()
				})
			}
		})

	g.Pages.AddAndSwitchToPage("modal", g.Modal(modal, 50, 29), true).ShowPage("main")
}

func (g *Gui) CloseAndSwitchPanel(removePrimitive string, primitive tview.Primitive) {
	g.Pages.RemovePage(removePrimitive).ShowPage("main")
	g.SwitchPanel(primitive)
}

func (g *Gui) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func (g *Gui) SwitchPanel(p tview.Primitive) *tview.Application {
	//g.UpdateViews()
	return g.App.SetFocus(p)
}

func (g *Gui) UpdateViews(pid proc.PID) {
	g.updateChannel <- pid

}

func (g *Gui) CurrentPanelKind() int {
	return g.Panels.Kinds[g.Panels.Current]
}

func (g *Gui) Run() error {
	g.SetKeybinds()
	if err := g.ProcessManager.UpdateView(); err != nil {
		return err
	}
	// when start app, set select index 0
	g.ProcessManager.Select(1, 0)
	p := g.ProcessManager.Selected()
	g.UpdateViews(p.Pid)

	infoGrid := tview.NewGrid().SetRows(0, 0, 0, 0).
		SetColumns(30, 0).
		AddItem(g.ProcessManager, 0, 0, 4, 1, 0, 0, true).
		AddItem(g.ProcessInfoView, 0, 1, 1, 1, 0, 0, true).
		AddItem(g.ProcessFileView, 1, 1, 1, 1, 0, 0, true).
		AddItem(g.ProcessEnvView, 2, 1, 1, 1, 0, 0, true).
		AddItem(g.ProcessTreeView, 3, 1, 1, 1, 0, 0, true)

	grid := tview.NewGrid().SetRows(1, 0, 2).
		SetColumns(30).
		AddItem(g.FilterInput, 0, 0, 1, 1, 0, 0, true).
		AddItem(infoGrid, 1, 0, 1, 2, 0, 0, true).
		AddItem(g.NaviView, 2, 0, 1, 2, 0, 0, true)

	g.Pages = tview.NewPages().
		AddAndSwitchToPage("main", grid, true)

	if err := g.App.SetRoot(g.Pages, true).Run(); err != nil {
		g.App.Stop()
		log.Println(err)
		return err
	}

	return nil
}
