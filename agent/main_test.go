package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/Ishan27g/ryo-Faas/agent/registry"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/stretchr/testify/assert"
)

func setup(ctx context.Context, registryPort string) *registry.AgentHandler {
	registry.SetBuildCommand(func(fn *deploy.Function, port string) *run.RunInfo {
		return run.CMD("sleep", "10").Ctx(ctx)
	})
	agent := registry.Init(registryPort)
	transport.Init(ctx, agent, DefaultPort, nil, "").Start()
	return agent
}

func TestSetup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-time.After(3 * time.Second)
	}()
	assert.NotNil(t, setup(ctx, DefaultPort))
}
func TestList(t *testing.T) {

	fmt.Println(os.Getwd())

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-time.After(3 * time.Second)
	}()
	agentHandler := setup(ctx, DefaultPort)
	assert.NotNil(t, agentHandler)

	dir := "/Users/ishan/Desktop/multi/method1"
	filePath := dir + "/method1.go"
	entrypoint := "Method1"
	c := transport.ProxyGrpcClient(DefaultPort)

	uploaded := transport.UploadDir(c, ctx, dir, entrypoint)
	assert.True(t, uploaded)

	deployRsp, err := c.Deploy(ctx, &deploy.DeployRequest{Functions: &deploy.Function{
		Entrypoint:       entrypoint,
		FilePath:         filePath,
		Dir:              dir,
		Zip:              "",
		AtAgent:          "",
		ProxyServiceAddr: "",
		Url:              "",
		Status:           "",
	},
	})
	assert.NoError(t, err)

	agentHandler.Println(deployRsp)

	fmt.Println()
	c.GetMetrics()

	<-time.After(2 * time.Second)

	logs, err := c.Logs(ctx, &deploy.Function{Entrypoint: entrypoint})
	if err != nil {
		return
	}
	assert.NoError(t, err)

	agentHandler.Println(logs)
	assert.Equal(t, "DEPLOYED", logs.Fn.Status)

	fmt.Println()
	c.GetMetrics()

	<-time.After(3 * time.Second)
	logs, err = c.Logs(ctx, &deploy.Function{Entrypoint: entrypoint})
	if err != nil {
		return
	}
	assert.NoError(t, err)

	agentHandler.Println(logs)
	assert.Equal(t, "DEPLOYED", logs.Fn.Status)

	fmt.Println()
	c.GetMetrics()

}
