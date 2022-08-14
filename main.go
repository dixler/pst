package main

import (
	"flag"
	"io/ioutil"
	"runtime"
	"runtime/debug"

	"github.com/dixler/pst/gui"
)

var (
	enableLog  = flag.Bool("log", false, "enable output log")
	filterWord = flag.String("proc", "", "use word to filtering process name when starting")
)

func run() int {
	flag.Parse()

	if err := gui.New().Run(); err != nil {
		return 1
	}

	return 0
}

func main() {
	// TODO implements windows
	if runtime.GOOS == "windows" {
		panic("no windows")
	}

	defer func() {
		if r := recover(); r != nil {
			ioutil.WriteFile("crashdump.txt", []byte(debug.Stack()), 0666)
		}
	}()
	run()
}
