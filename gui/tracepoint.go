package gui

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type ExecData struct {
	PPID    PID
	Command string
}

func NewExecDataSource() (Datasource[ExecData], error) {
	const execTrace = `
tracepoint:syscalls:sys_enter_exec*
{
    printf("%d %d %s\n", curtask->real_parent->tgid, pid, str(args->argv[0]));
}
`
	return NewSource(execTrace, func(line string) (PID, ExecData, error) {
		s := strings.SplitN(strings.TrimSpace(line), " ", 3)
		if len(s) != 3 {
			return PID(""), ExecData{}, fmt.Errorf("unable to parse '%s'\n", line)
		}
		pid, ppid, cmd := PID(s[0]), PID(s[1]), s[2]
		return pid, ExecData{
			PPID:    ppid,
			Command: cmd,
		}, nil
	})
}

type ChdirData struct {
	Cwd string
}

func NewChdirDataSource() (Datasource[ChdirData], error) {
	const chdirTrace = `
tracepoint:syscalls:sys_enter_chdir*
{
    printf("%d %s\n", pid, str(args->filename));
}
`
	return NewSource(chdirTrace, func(line string) (PID, ChdirData, error) {
		s := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(s) != 2 {
			return PID(""), ChdirData{}, fmt.Errorf("unable to parse '%s'\n", line)
		}
		pid, cwd := PID(s[0]), s[1]
		return pid, ChdirData{
			Cwd: cwd,
		}, nil
	}, func(pid PID) (ChdirData, error) {
		path, err := filepath.EvalSymlinks(path.Join("/proc", pid.String(), "cwd"))
		if err != nil {
			return ChdirData{}, err
		}
		return ChdirData{
			Cwd: path,
		}, nil
	},
	)
}

type OpenData struct {
	Filepath string
}

func NewOpenDataSource() (Datasource[OpenData], error) {
	const openTrace = `
tracepoint:syscalls:sys_enter_open,
tracepoint:syscalls:sys_enter_openat
{
	@filename[tid] = args->filename;
}

tracepoint:syscalls:sys_exit_open,
tracepoint:syscalls:sys_exit_openat
/@filename[tid]/
{
	$ret = args->ret;
	$fd = $ret > 0 ? $ret : -1;

	printf("%d %d %s\t\n", pid, $fd, str(@filename[tid]));
	delete(@filename[tid]);
}

END
{
	clear(@filename);
}
`
	return NewSource(openTrace, func(line string) (PID, OpenData, error) {
		s := strings.SplitN(strings.TrimSpace(line), " ", 3)
		if len(s) != 3 {
			return PID(""), OpenData{}, fmt.Errorf("unable to parse '%s'\n", line)
		}

		pid, retval, filepath := PID(s[0]), s[1], s[2]

		if retval == "-1" {
			return PID(""), OpenData{}, fmt.Errorf("open retval -1 '%s'\n", line)
		}

		if len(filepath) == 200-1 {
			filepath = filepath + "<...>"
		}

		return pid, OpenData{
			Filepath: filepath,
		}, nil
	})
}
