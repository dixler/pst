package gui

import (
	"time"

	"github.com/dixler/pst/gui/proc"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (g *Gui) nextPanel() {
	idx := (g.Panels.Current + 1) % len(g.Panels.Panels)
	g.Panels.Current = idx
	g.SwitchPanel(g.Panels.Panels[g.Panels.Current])
}

func (g *Gui) prePanel() {
	g.Panels.Current--

	if g.Panels.Current < 0 {
		g.Current = len(g.Panels.Panels) - 1
	} else {
		idx := (g.Panels.Current) % len(g.Panels.Panels)
		g.Panels.Current = idx
	}
	g.SwitchPanel(g.Panels.Panels[g.Panels.Current])
}

func (g *Gui) GlobalKeybind(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyTab:
		g.nextPanel()
	case tcell.KeyBacktab:
		g.prePanel()
	}

	g.NaviView.UpdateView(g)
}

func (g *Gui) ProcessManagerKeybinds() {
	g.ProcessManager.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			g.App.Stop()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'K':
			if p := g.ProcessManager.Selected(); p != nil {
				g.Confirm("Do you want to kill this process?", "kill", g.ProcessManager, func() {
					proc.Kill(p.Pid)
					//g.ProcessManager.UpdateView()
				})
			}
		}

		g.GlobalKeybind(event)
		return event
	})

	g.ProcessManager.SetSelectionChangedFunc(func(row, col int) {
		if row < 1 {
			return
		}

		p := g.ProcessManager.Selected()

		g.UpdateViews(p.Pid)
	})
}

func (g *Gui) FilterInputKeybinds() {
	g.FilterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			g.App.Stop()
		case tcell.KeyEnter:
			g.nextPanel()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.GlobalKeybind(event)
		return event
	})

	g.FilterInput.SetChangedFunc(func(text string) {
		g.ProcessManager.FilterWord = text
		g.ProcessManager.UpdateView()
	})
}

func (g *Gui) ProcessTreeViewKeybinds() {
	g.ProcessTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		g.ProcessTreeView.ExpandToggle(g.ProcessManager, node, !node.IsExpanded())
	})

	g.ProcessTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		node := g.ProcessTreeView.GetCurrentNode()
		switch event.Rune() {
		case 'K':
			if ref := node.GetReference(); ref != nil {
				g.Confirm("Do you want to kill this process?", "kill", g.ProcessTreeView, func() {
					proc.Kill(ref.(proc.PID))
					// wait a little to finish process killing
					time.Sleep(1 * time.Millisecond)
					p := g.ProcessManager.Selected()
					g.ProcessTreeView.UpdateTree(g, p.Pid)
				})
			}
		case 'l':
			g.ProcessTreeView.ExpandToggle(g.ProcessManager, node, true)
		case 'h':
			g.ProcessTreeView.ExpandToggle(g.ProcessManager, node, false)
		}
		g.GlobalKeybind(event)
		return event
	})

	g.ProcessTreeView.SetChangedFunc(func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref := node.GetReference()
		if ref == nil {
			return
		}
		pid := ref.(proc.PID)

		g.UpdateViews(pid)
	})
}

func (g *Gui) ProcessEnvViewKeybinds() {
	g.ProcessEnvView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.GlobalKeybind(event)
		return event
	})
}

func (g *Gui) ProcessInfoViewKeybinds() {
	g.ProcessInfoView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.GlobalKeybind(event)
		return event
	})
}

func (g *Gui) ProcessFileViewKeybinds() {
	g.ProcessFileView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.GlobalKeybind(event)
		return event
	})
}

func (g *Gui) SetKeybinds() {
	g.FilterInputKeybinds()
	g.ProcessManagerKeybinds()
	g.ProcessTreeViewKeybinds()
	g.ProcessInfoViewKeybinds()
	g.ProcessEnvViewKeybinds()
	g.ProcessFileViewKeybinds()
}
