package shell

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DavidGamba/dgtools/run"
)

type Option func(*shell)

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
	s.done = make(chan bool, 1)
	return s
}

type Shell interface {
	Kill()
	Run() bool
	WaitTillDone()
}

func (s *shell) Run() bool {

	ctx, cancel := context.WithCancel(context.Background())
	s.kill = func() {
		cancel()
		s.done <- true
	}
	var err error
	go func(s *shell, ctx context.Context) {
		err = s.c.Ctx(ctx).Run(s.logWriter)
		if err != nil {
			fmt.Println(err.Error())
		}
		s.kill()
	}(s, ctx)
	<-time.After(2 * time.Second)
	if err == nil {
	}
	return err == nil
}

type kill func()
type shell struct {
	c         *run.RunInfo
	logWriter io.WriteCloser
	kill
	done chan bool
}

func (s *shell) WaitTillDone() {
	<-s.done
	return
}

func (s *shell) Kill() {
	s.kill()
}
