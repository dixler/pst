package proc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

func Info(pid PID) (string, error) {
	if pid == "0" {
		return "", nil
	}

	cmd := exec.Command("ps", "-o", "pid,ppid,%cpu,%mem,lstart,user,command", "-p", pid.String())
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	if err := cmd.Run(); err != nil {
		return "", errors.New(string(buf))
	}

	return string(buf), nil
}

func GetEnv(pid PID) ([]string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return []string{}, nil
	}

	if pid == "0" {
		return []string{}, nil
	}

	env, err := readProcPath(pid, "environ")
	if err != nil {
		return []string{}, nil
	}

	result := strings.Split(env, "\x00")

	return result, nil
}

func OpenFiles(pid PID) (string, error) {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		return "", nil
	}

	if pid == "0" {
		return "", nil
	}

	buf := bytes.Buffer{}
	cmd := exec.Command("lsof", "-p", pid.String())
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return "", errors.New(buf.String())
	}

	result := strings.SplitN(buf.String(), "\n", 2)
	if len(result) > 1 {
		result[0] = fmt.Sprintf("[yellow]%s[white]", result[0])
	} else {
		return buf.String(), nil
	}

	return strings.Join(result, "\n"), nil
}

func GetCommand(pid PID) string {
	str, _ := readProcPath(pid, "comm")
	return str
}

func readProcPathBytes(pid PID, p string) ([]byte, error) {
	return ioutil.ReadFile(path.Join("/proc", pid.String(), p))
}

func readProcPath(pid PID, p string) (string, error) {
	b, err := ioutil.ReadFile(path.Join("/proc", pid.String(), p))
	if err != nil {
		return "", err
	}
	return string(b), err
}
