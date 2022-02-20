package registry

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/DavidGamba/dgtools/run"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
)

// system provides multi to control functions that are run as os processes
type system interface {
	// Run a function as its own process
	run(fn *deploy.Function, port string) bool
	// Stop a running process by its function name
	Stop(fnName string)
	// Logs returns a channel that returns latest logs (todo tail logs instead?)
	Logs(fnName string) []string
}

func newSystem() system {
	return &shell{
		make(map[string]*command),
		log.New(os.Stdout, "[SHELL]", log.Ltime),
	}
}

type shell struct {
	processes map[string]*command
	*log.Logger
}

type command struct {
	done chan bool
	Fn   *deploy.Function
	p    *process
}
type process struct {
	io.ReadCloser
	logs []string
}

func newProcess() *process {
	var pr = new(process)
	pr.logs = []string{}
	return pr
}
func (s *shell) run(fn *deploy.Function, port string) bool {

	pr := newProcess()
	cmd := bc(fn, port)
	s.processes[fn.Entrypoint] = &command{
		Fn:   fn,
		p:    pr,
		done: make(chan bool),
	}
	ok := pr.execCmd(cmd, s.processes[fn.Entrypoint].done)
	if !ok {
		delete(s.processes, fn.Entrypoint)
	}
	s.Println("RUNNING - ", fn.Entrypoint, s.processes[fn.Entrypoint].Fn)
	return ok
}

func (s *shell) Stop(fnName string) {
	if s.processes[fnName] == nil {
		fmt.Println("shell process not found", fnName)
		return
	}
	s.processes[fnName].done <- true
	fmt.Println("Stopped ", fnName)
	delete(s.processes, fnName)
}

func (s *shell) Logs(fnName string) []string {
	if s.processes[fnName] == nil {
		fmt.Println("Process not found", fnName)
		return nil
	}
	return s.processes[fnName].p.Logs()
}

func (p *process) execCmd(cmd *run.RunInfo, end chan bool) bool {
	var err error
	r, w := io.Pipe()
	p.ReadCloser = r
	p.logs = []string{}

	ctx, cancel := context.WithCancel(context.Background())
	cmd.Ctx(ctx)
	defer func() {
		go func() {
			<-end
			cancel()
		}()
	}()
	// run command
	go func() {
		err = cmd.Run(w)
		if err != nil {
			fmt.Println("p-run-", err.Error())
		}
	}()

	// close pipes on cancel
	go func() {
		<-ctx.Done()
		w.Close()
		p.ReadCloser.Close()
		fmt.Println("Context done")
	}()

	go func() {
		scanner := bufio.NewScanner(p.ReadCloser)
		for scanner.Scan() {
			log := scanner.Text()
			fmt.Println(log)
			p.logs = append(p.logs, log)
		}
		fmt.Println("EXITTTTTT")
	}()

	return err == nil
}

var bc func(fn *deploy.Function, port string) *run.RunInfo

func SetBuildCommand(f func(fn *deploy.Function, port string) *run.RunInfo) {
	if f == nil {
		bc = buildCommand
		return
	}
	bc = f
}

func buildCommand(fn *deploy.Function, port string) *run.RunInfo {
	cmd := run.CMD("go", "run", fn.FilePath).
		Env("PORT="+port, "URL="+strings.ToLower(fn.Entrypoint))
		//Ctx(ctx)
	return cmd
}
func (p *process) Logs() []string {
	return p.logs
}
