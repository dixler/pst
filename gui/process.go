package gui

type Process struct {
	Pid   PID
	PPid  PID
	Cmd   string
	Child []PID
}
