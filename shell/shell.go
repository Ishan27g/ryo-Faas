package shell

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
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

type Cmd struct {
	out  io.WriteCloser
	port string
	ctx  context.Context
	Cmd  *exec.Cmd
}

func (c Cmd) Run() error {
	c.Cmd.Stdout = io.MultiWriter(os.Stdout, c.out)
	c.Cmd.Stderr = io.MultiWriter(os.Stderr, c.out)
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	waitDone := make(chan struct{})
	defer close(waitDone)
	go func() {
		select {
		case <-c.ctx.Done():
			if err := c.Cmd.Process.Kill(); err != nil {
				fmt.Println(err.Error())
			}
			if err := c.Cmd.Process.Signal(os.Interrupt); err != nil {
				fmt.Println(err.Error())
				_ = c.Cmd.Process.Kill()
			} else {
				defer time.AfterFunc(time.Second*10, func() {
					if err := c.Cmd.Process.Kill(); err != nil {
						fmt.Println(err.Error())
					}
				}).Stop()
				<-waitDone
			}
		case <-waitDone:
		}
	}()
	err := c.Cmd.Wait()
	log.Printf("kill %q", c.Cmd.Args)
	return err
}
func CommandContext(ctx context.Context, filePath string, port string, w io.WriteCloser) Cmd {
	c := exec.Command("go", "run", filePath, "--port", port)
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	return Cmd{ctx: ctx, Cmd: c, port: port, out: w}
}
func (s *shell) Run() bool {

	ctx, cancel := context.WithCancel(context.Background())
	s.kill = func() {
		cancel()
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
	fnName    string
	logs      []string
	logWriter io.WriteCloser
	kill
}

func (s *shell) Kill() {
	s.kill()
}
func buildCommand(filePath string, entrypoint string, port string, ctx context.Context) *run.RunInfo {
	cmd := run.CMD("go", "run", filePath, "--port", port).
		Env("PORT=" + port).SaveErr().Ctx(ctx)
	// Env("PORT="+port, "URL="+strings.ToLower(entrypoint)).SaveErr().Ctx(ctx)
	return cmd
}
