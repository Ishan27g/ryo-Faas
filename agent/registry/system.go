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

func newSystem() *system {
	return &system{
		process: make(map[string]*shell),
	}
}

type kill func()
type system struct {
	process map[string]*shell
}

func (s *system) run(fn *deploy.Function, port string) bool {
	var runs []string
	s.process[fn.Entrypoint] = newShell(fn.GetEntrypoint())
	if !s.process[fn.Entrypoint].run(fn.FilePath, fn.Entrypoint, port) {
		for _, name := range runs {
			s.stop(name)
		}
	} else {
		runs = append(runs, fn.Entrypoint)
	}
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
