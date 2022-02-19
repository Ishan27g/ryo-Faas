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
	Fn *deploy.Function
	p  *process
}
type process struct {
	context.CancelFunc
	io.ReadCloser
	logs []string
}

func newProcess() (*process, context.Context) {
	var pr = new(process)
	ctx, cancel := context.WithCancel(context.Background())
	pr.CancelFunc = cancel
	pr.logs = []string{}
	return pr, ctx
}
func (s *shell) run(fn *deploy.Function, port string) bool {

	pr, ctx := newProcess()
	cmd := bc(fn, port, ctx)
	ok := pr.execCmd(ctx, cmd)
	if ok {
		s.processes[fn.Entrypoint] = &command{
			Fn: fn,
			p:  pr,
		}
	}
	s.Println("RUNNING - ", s.processes[fn.Entrypoint])
	return ok
}

func (s *shell) Stop(fnName string) {
	if s.processes[fnName] == nil {
		fmt.Println("shell process not found", fnName)
		return
	}
	s.processes[fnName].p.CancelFunc()
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

func (p *process) execCmd(ctx context.Context, cmd *run.RunInfo) bool {
	var err error

	r, w := io.Pipe()
	p.ReadCloser = r
	p.logs = []string{}

	fmt.Println("EXECUTING - ")

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

	if err != nil {
		return false
	}

	return true
}

var bc func(fn *deploy.Function, port string, ctx context.Context) *run.RunInfo

func SetBuildCommand(f func(fn *deploy.Function, port string, ctx context.Context) *run.RunInfo) {
	if f == nil {
		bc = buildCommand
		return
	}
	bc = f
}

func buildCommand(fn *deploy.Function, port string, ctx context.Context) *run.RunInfo {
	cmd := run.CMD("go", "run", fn.FilePath, deployFile()).
		Env("PORT="+port, "URL="+strings.ToLower(fn.Entrypoint)).
		Ctx(ctx)
	return cmd
}
func (p *process) Logs() []string {
	return p.logs
}
