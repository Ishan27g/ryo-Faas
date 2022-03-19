package shell

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DavidGamba/dgtools/run"
)

type Option func(*shell)

func WithFunctionName(name string) Option {
	return func(s *shell) {
		s.fnName = name
	}
}
func WithCmd(c *run.RunInfo) Option {
	return func(s *shell) {
		s.c = c
	}
}
func WithOutput(w io.WriteCloser) Option {
	return func(s *shell) {
		s.logWriter = w
	}
}
func New(o ...Option) Shell {
	s := &shell{}
	for _, i2 := range o {
		i2(s)
	}
	s.logs = []string{}
	return s
}

type Shell interface {
	Kill()
	Run() bool
}

func (s *shell) Run() bool {
	ctx, cancel := context.WithCancel(context.Background())
	s.kill = func() {
		cancel()
	}
	var err error
	go func() {
		fmt.Println(".....................................starting")
		err = s.c.Ctx(ctx).Run(s.logWriter)
		if err != nil {
			fmt.Println(err.Error())
		}
		s.kill()
		fmt.Println(".....................................killed")
		//runtime.Goexit()
	}()
	<-time.After(1 * time.Second)
	if err == nil {
		fmt.Println(".....................................started")
	}
	return err == nil
}

type kill func()
type shell struct {
	c         *run.RunInfo
	fnName    string
	logs      []string
	logWriter io.WriteCloser
	kill
}

func (s *shell) Kill() {
	s.kill()
}
func buildCommand(filePath string, entrypoint string, port string, ctx context.Context) *run.RunInfo {
	cmd := run.CMD("go", "run", filePath).
		Env("PORT=" + port).SaveErr().Ctx(ctx)
	// Env("PORT="+port, "URL="+strings.ToLower(entrypoint)).SaveErr().Ctx(ctx)
	return cmd
}
