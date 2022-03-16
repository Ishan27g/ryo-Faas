package registry

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/DavidGamba/dgtools/run"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
)

// system provides multi to control functions that are run as os processes
//type system1 interface {
//	// Run a function as its own process
//	run(fn *deploy.Function, port string) bool
//	// Stop a running process by its function name
//	stop(fnName string)
//	// Logs returns a channel that returns latest logs (todo tail logs instead?)
//	logs(fnName string) []byte
//}

func newSystem() *system {
	return &system{
		process: make(map[string]*shell),
	}
}

type kill func()
type system struct {
	process map[string]*shell
}

func (s *system) run(fns []*deploy.Function, port string) bool {
	var runs []string
	fn := fns[0]
	// for _, fn := range fns {
	s.process[fn.Entrypoint] = newShell(fn.GetEntrypoint())
	if !s.process[fn.Entrypoint].run(fn.FilePath, fn.Entrypoint, port) {
		for _, name := range runs {
			s.stop(name)
		}
	} else {
		runs = append(runs, fn.Entrypoint)

	}
	// }
	return len(runs) != 0
}

func (s *system) stop(fnName string) {
	if s.process[fnName] == nil {
		return
	}
	s.process[fnName].kill()
}

func (s *system) logs(fnName string) []string {
	if s.process[fnName] == nil {
		return nil
	}
	return s.process[fnName].logs
}

type shell struct {
	fnName string
	logs   []string
	kill
}

func newShell(fnName string) *shell {
	return &shell{
		fnName: fnName,
		logs:   []string{},
		kill:   nil,
	}
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

//
//func newProcess() *process {
//	var pr = new(process)
//	pr.logs = []string{}
//	return pr
//}
//func (s *shell) run(fn *deploy.Function, port string) bool {
//
//	pr := newProcess()
//	cmd := bc(fn, port)
//	s.processes[fn.Entrypoint] = &command{
//		Fn:   fn,
//		p:    pr,
//		done: make(chan bool),
//	}
//	ok := pr.execCmd(cmd, s.processes[fn.Entrypoint].done)
//	if !ok {
//		delete(s.processes, fn.Entrypoint)
//	}
//	s.Println("RUNNING - ", fn.Entrypoint, s.processes[fn.Entrypoint].Fn)
//	return ok
//}
//
//func (s *shell) stop(fnName string) {
//	if s.processes[fnName] == nil {
//		fmt.Println("shell process not found", fnName)
//		return
//	}
//	s.processes[fnName].done <- true
//	fmt.Println("Stopped ", fnName)
//	delete(s.processes, fnName)
//}
//
//func (s *shell) logs(fnName string) []string {
//	if s.processes[fnName] == nil {
//		fmt.Println("Process not found", fnName)
//		return nil
//	}
//	return s.processes[fnName].p.logs()
//}
//
//func (p *process) execCmd(cmd *run.RunInfo, end chan bool) bool {
//	var err error
//	r, w := io.Pipe()
//	p.ReadCloser = r
//	p.logs = []string{}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	cmd.Ctx(ctx)
//	defer func() {
//		go func() {
//			<-end
//			cancel()
//		}()
//	}()
//	// run command
//	go func() {
//		err = cmd.Run(w)
//		if err != nil {
//			fmt.Println("p-run-", err.Error())
//		}
//	}()
//	// close pipes on cancel
//	go func() {
//		<-ctx.Done()
//		w.Close()
//		p.ReadCloser.Close()
//		fmt.Println("Context done")
//	}()
//
//	go func() {
//		scanner := bufio.NewScanner(p.ReadCloser)
//		for scanner.Scan() {
//			log := scanner.Text()
//			fmt.Println(log)
//			p.logs = append(p.logs, log)
//		}
//		fmt.Println("EXITTTTTT")
//	}()
//
//	return err == nil
//}
//
//var bc func(fn *deploy.Function, port string) *run.RunInfo
//
//func SetBuildCommand(f func(fn *deploy.Function, port string) *run.RunInfo) {
//	if f == nil {
//		bc = buildCommand
//		return
//	}
//	bc = f
//}
//
//func buildCommand(fn *deploy.Function, port string) *run.RunInfo {
//	cmd := run.CMD("go", "run", fn.FilePath).
//		Env("PORT="+port, "URL="+strings.ToLower(fn.Entrypoint))
//	//Ctx(ctx)
//	return cmd
//}
var cmd *run.RunInfo

func SetBuildCommand(c *run.RunInfo) {
	cmd = c
}
func (s *shell) run(filePath, entrypoint, port string) bool {
	fmt.Println("fn.GenFilePath - ", filePath)
	fmt.Println("starting on - ", port)

	ctx, cancel := context.WithCancel(context.Background())
	cmd = buildCommand(filePath, entrypoint, port, ctx)

	r, w := io.Pipe()

	// start the service
	var err error
	go func() {
		err = cmd.Run(w)
		if err != nil {
			fmt.Println(err.Error())
			runtime.Goexit()
		}
	}()
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			log := scanner.Text()
			fmt.Println(log)
			s.logs = append(s.logs, log)
		}
		fmt.Println("exitttt")
		s.kill()
	}()
	s.kill = func() {
		cancel()
		w.Close()
		fmt.Println("killed")
	}
	return err == nil
}

func buildCommand(filePath string, entrypoint string, port string, ctx context.Context) *run.RunInfo {
	cmd := run.CMD("go", "run", filePath).
		Env("PORT=" + port).SaveErr().Ctx(ctx)
	// Env("PORT="+port, "URL="+strings.ToLower(entrypoint)).SaveErr().Ctx(ctx)
	return cmd
}
func (p *process) Logs() []string {
	return p.logs
}
