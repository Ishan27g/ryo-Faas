package shell

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/stretchr/testify/assert"
)

func TestShell_Run(t *testing.T) {

	err := os.Chdir("../agent")
	assert.NoError(t, err)

	pr, pw := io.Pipe()
	defer pw.Close()

	shOpts := []Option{WithOutput(pw), WithCmd(run.CMD("go", "run", "main.go"))}
	sh := New(shOpts...)

	sh.Run()
	go func() {
		io.Copy(os.Stdout, pr)
	}()
	<-time.After(3 * time.Second)
	fmt.Println("killing")
	sh.Kill()

	<-time.After(2 * time.Second)

}
func Test_TaskShell(t *testing.T) {
	err := os.Chdir("../")
	assert.NoError(t, err)

	pr, pw := io.Pipe()
	defer pw.Close()

	shOpts := []Option{WithOutput(pw), WithCmd(run.CMD("task", "imgDb"))}
	sh := New(shOpts...)

	sh.Run()
	go func() {
		io.Copy(os.Stdout, pr)
	}()
	sh.WaitTillDone()
}
