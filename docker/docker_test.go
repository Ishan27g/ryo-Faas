package docker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDocker_Start(t *testing.T) {
	assert.True(t, New().Start())
}
func TestDocker_StopRfa(t *testing.T) {
	d := New()
	assert.NotNil(t, d)
	d.Stop()
	fmt.Println(d.StatusAll())
}

func TestDocker_Stop_Step(t *testing.T) {
	d := New()
	assert.NotNil(t, d.Stop())
}

func TestDocker_List(t *testing.T) {
	New().StatusAll()
}

func TestDocker_RunFunction(t *testing.T) {
	d := New()
	d.RunFunction("method-async")
	<-time.After(5 * time.Second)
	isRunning := d.CheckFunction("method-async")
	fmt.Println("isRunning", isRunning)
}
func TestDocker_StopFunction(t *testing.T) {
	New().StopFunction("method-async")
}
func TestDocker_CheckLabel(t *testing.T) {
	d := New()
	d.CheckLabel()
}
