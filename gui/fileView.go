package gui

import (
	"github.com/dixler/pst/gui/proc"
	"github.com/rivo/tview"
)

type ProcessFileView struct {
	*tview.TextView
}

func NewProcessFileView() *ProcessFileView {
	p := &ProcessFileView{
		TextView: tview.NewTextView().SetDynamicColors(true),
	}

	p.SetTitleAlign(tview.AlignLeft).SetTitle("process open files").SetBorder(true)
	p.SetWrap(false)
	return p
}

func (p *ProcessFileView) UpdateViewWithPid(g *Gui, pid proc.PID) {
	text := ""
	info, err := proc.OpenFiles(pid)
	if err != nil {
		text = err.Error()
	} else {
		text = info
	}

	g.App.QueueUpdateDraw(func() {
		p.SetText(text)
		p.ScrollToBeginning()
	})
}
