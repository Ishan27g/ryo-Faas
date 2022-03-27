package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/Ishan27g/ryo-Faas/pkg/shell"
)

var bI = flag.Bool("buildImages", false, "build Images ? ")
var pI = flag.Bool("pushImages", false, "push Images ? ")
var silent = flag.Bool("silent", false, "show logs ? ")

var (
	buildImages = builds
	pushImages  = push
)

var builds = [][]string{
	{"task", "imgDeployBase"},
	{"task", "imgDb"},
	{"task", "imgProxy"},
}
var push = [][]string{
	{"task", "pushImages"},
}

func runTask(wg *sync.WaitGroup, command ...string) {
	defer wg.Done()
	pr, pw := io.Pipe()
	defer pw.Close()
	shOpts := []shell.Option{shell.WithOutput(pw), shell.WithCmd(run.CMD(command...))}
	sh := shell.New(shOpts...)
	sh.Run()
	if !*silent {
		go func() { io.Copy(os.Stdout, pr) }()
	}
	sh.WaitTillDone()
}
func timeIt(since time.Time) {
	fmt.Println("\nTook : ", time.Since(since).String())
}

// run from root
func main() {
	defer timeIt(time.Now())
	flag.Parse()
	var wg sync.WaitGroup
	if *bI {
		for _, image := range buildImages {
			wg.Add(1)
			go runTask(&wg, image...)
		}
		wg.Wait()
	}
	if *pI {
		for _, image := range pushImages {
			wg.Add(1)
			go runTask(&wg, image...)
		}
	}
	wg.Wait()
}
