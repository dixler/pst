package gui

import (
	"fmt"

	"github.com/dixler/pst/gui/proc"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type ProcessNode struct {
	node     *tview.TreeNode
	children []proc.PID
}

type ProcessTreeView struct {
	*tview.TreeView
	getProcess func(proc.PID) *proc.Process
	pidMap     map[proc.PID]*ProcessNode
}

func NewProcessTreeView(
	getProcess func(proc.PID) *proc.Process) *ProcessTreeView {

	p := &ProcessTreeView{
		TreeView:   tview.NewTreeView(),
		getProcess: getProcess,
		pidMap:     make(map[proc.PID]*ProcessNode),
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

func (p *ProcessTreeView) UpdateTree(pid proc.PID) {
	ps := p.getProcess(pid)
	if ps == nil {
		return
	}
	curRoot := p.GetRoot()
	if curRoot != nil {
		rootPid := curRoot.GetReference().(proc.PID)
		if rootPid == pid {
			return
		}
	}

	root, ok := p.pidMap[pid]
	if ok {
		p.SetRoot(root.node).
			SetCurrentNode(root.node)
		p.addNode(root.node, pid)
		return
	}
	root = &ProcessNode{
		children: ps.Child,
		node: tview.
			NewTreeNode(pid.String()).
			SetReference(pid),
	}
	p.SetRoot(root.node).
		SetCurrentNode(root.node)
	p.addNode(root.node, pid)
	p.pidMap[pid] = root
}

func (p *ProcessTreeView) addNode(target *tview.TreeNode, pid proc.PID) {
	pro := p.getProcess(pid)
	if pro == nil {
		return
	}

	for _, child := range pro.Child {
		node, ok := p.pidMap[child]
		if ok {
			continue

		}
		node = &ProcessNode{
			node: tview.
				NewTreeNode(fmt.Sprintf("[%s] %s", child, proc.GetCommand(child))).
				SetReference(child),
		}

		childProcess := p.getProcess(child)
		if childProcess == nil {
			continue
		}
		node.node.SetSelectable(true)
		if len(childProcess.Child) > 0 {
			node.node.SetColor(tcell.ColorGreen)
		}
		target.AddChild(node.node)
		p.pidMap[child] = node
	}
}
