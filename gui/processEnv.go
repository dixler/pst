package gui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/dixler/pst/gui/proc"
	"github.com/rivo/tview"
)

type EnvView struct {
	*tview.TextView
}

func NewEnvView() *EnvView {
	p := &EnvView{
		TextView: tview.NewTextView().SetDynamicColors(true),
	}

	p.SetTitleAlign(tview.AlignLeft).SetTitle("process environments").SetBorder(true)
	p.SetWrap(false)
	return p
}

func (p *EnvView) UpdateViewWithPid(g *Gui, pid proc.PID) {
	text := ""
	info, err := renderEnv(pid)
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

func renderEnv(pid proc.PID) (string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	env, err := proc.GetEnv(pid)
	if err != nil {
		return "", err
	}

	var (
		envs []string
	)

	for _, e := range env {
		kv := strings.SplitN(e, "=", 1)
		if len(kv) != 2 {
			envs = append(envs, fmt.Sprintf("[magenta]%s", e))
			continue
		}
		envs = append(envs, fmt.Sprintf("[yellow]%s[white]\t%s", kv[0], kv[1]))
	}

	return strings.Join(envs, "\n"), nil
}
