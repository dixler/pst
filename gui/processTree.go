package gui

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type ProcessTreeView struct {
	*tview.TreeView
}

func NewProcessTreeView(pm *ProcessManager) *ProcessTreeView {
	p := &ProcessTreeView{
		TreeView: tview.NewTreeView(),
	}

	p.SetBorder(true).SetTitle("process tree").SetTitleAlign(tview.AlignLeft)
	return p
}

func (p *ProcessTreeView) ExpandToggle(pm *ProcessManager, node *tview.TreeNode, isExpand bool) {
	reference := node.GetReference()
	if reference == nil {
		return // Selecting the root node does nothing.
	}
	children := node.GetChildren()
	if len(children) == 0 {
		pid := reference.(PID)
		p.addNode(pm, node, pid)
	} else {
		// Collapse if visible, expand if collapsed.
		node.SetExpanded(isExpand)
	}
}

func (p *ProcessTreeView) UpdateTree(g *Gui, pid PID) {

	root := tview.NewTreeNode(pid.String()).
		SetColor(tcell.ColorYellow)

	p.SetRoot(root).
		SetCurrentNode(root)

	p.addNode(g.ProcessManager, root, pid)
}

func (p *ProcessTreeView) addNode(pm *ProcessManager, target *tview.TreeNode, pid PID) {
	processes, err := pm.GetProcesses()
	if err != nil {
		return
	}

	proc, ok := processes[pid]
	if !ok {
		return
	}

	for _, p := range proc.Child {
		node := tview.NewTreeNode(fmt.Sprintf("PID: %s CMD: %s", p, pm.procDs.GetCommand(p))).
			SetReference(p)

		p, ok := processes[p]
		node.SetSelectable(ok)
		if len(p.Child) > 0 {
			node.SetColor(tcell.ColorGreen)
		}
		target.AddChild(node)
	}
}
