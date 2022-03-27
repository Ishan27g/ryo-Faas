package docker

//
//import (
//	"fmt"
//	"log"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/docker/docker/client"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestDocker_Start(t *testing.T) {
//	assert.True(t, New().Start())
//}
//func TestDocker_StopRfa(t *testing.T) {
//	d := New()
//	assert.NotNil(t, d)
//	d.Stop()
//	fmt.Println(d.StatusAll())
//}
//
//func TestDocker_Stop_Step(t *testing.T) {
//	d := New()
//	assert.NotNil(t, d.Stop())
//}
//
//func TestDocker_List(t *testing.T) {
//	New().StatusAll()
//}
//
//func TestDocker_RunFunction(t *testing.T) {
//	d := New()
//	d.RunFunction("method-async")
//	<-time.After(5 * time.Second)
//	isRunning := d.CheckFunction("method-async")
//	fmt.Println("isRunning", isRunning)
//}
//func TestDocker_StopFunction(t *testing.T) {
//	New().StopFunction("method-async")
//}
//func TestDocker_CheckLabel(t *testing.T) {
//	d := New()
//	d.CheckLabel()
//}
//func TestDocker_StartProxy(t *testing.T) {
//	cmd, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
//	if err != nil {
//		return
//	}
//	d := docker{false, cmd, log.New(os.Stdout, "docker", log.LstdFlags)}
//	d.startProxy()
//}
//func TestDocker_StopProxy(t *testing.T) {
//	cmd, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
//	if err != nil {
//		return
//	}
//	d := docker{false, cmd, log.New(os.Stdout, "docker", log.LstdFlags)}
//	d.stopProxy()
//}
