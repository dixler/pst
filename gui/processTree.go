package gui

import (
	"fmt"

	"github.com/dixler/pst/gui/proc"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type ProcessTreeView struct {
	*tview.TreeView
	getProcesses func() (map[proc.PID]proc.Process, error)
}

func NewProcessTreeView(
	getProcesses func() (map[proc.PID]proc.Process, error)) *ProcessTreeView {

	p := &ProcessTreeView{
		TreeView:     tview.NewTreeView(),
		getProcesses: getProcesses,
	}

	p.SetBorder(true).SetTitle("process tree").SetTitleAlign(tview.AlignLeft)
	return p
}

func (p *ProcessTreeView) ExpandToggle(node *tview.TreeNode, isExpand bool) {
	reference := node.GetReference()
	if reference == nil {
		return // Selecting the root node does nothing.
	}
	children := node.GetChildren()
	if len(children) == 0 {
		pid := reference.(proc.PID)
		p.addNode(node, pid)
	} else {
		// Collapse if visible, expand if collapsed.
		node.SetExpanded(isExpand)
	}
}

func (p *ProcessTreeView) UpdateTree(g *Gui, pid proc.PID) {

	root := tview.NewTreeNode(pid.String()).
		SetColor(tcell.ColorYellow)

	p.SetRoot(root).
		SetCurrentNode(root)

	p.addNode(root, pid)
}

func (p *ProcessTreeView) addNode(target *tview.TreeNode, pid proc.PID) {
	processes, err := p.getProcesses()
	if err != nil {
		return
	}

	pro, ok := processes[pid]
	if !ok {
		return
	}

	for _, child := range pro.Child {
		node := tview.NewTreeNode(fmt.Sprintf("[%s] %s", child, proc.GetCommand(child))).
			SetReference(child)

		childProcess, ok := processes[child]
		node.SetSelectable(ok)
		if len(childProcess.Child) > 0 {
			node.SetColor(tcell.ColorGreen)
		}
		target.AddChild(node)
	}
}
