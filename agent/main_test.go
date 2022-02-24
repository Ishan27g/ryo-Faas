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
	registry.SetBuildCommand(run.CMD("sleep", "10").Ctx(ctx))
	agent := registry.Init(registryPort)
	transport.Init(ctx, agent, DefaultPort, nil, "").Start()
	return agent
}

func TestSetup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		<-time.After(3 * time.Second)
		cancel()
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

	list, err := c.List(ctx, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: entrypoint}})
	assert.NoError(t, err)
	assert.Equal(t, "DEPLOYED", list.Functions[0].Status)

	//<-time.After(5 * time.Second)
	//list, err = c.List(ctx, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: entrypoint}})
	//assert.NoError(t, err)
	//assert.Equal(t, "DEPLOYED", list.Functions[0].Status)
	//
	//logs, err := c.Logs(ctx, &deploy.Function{Entrypoint: entrypoint})
	//assert.NoError(t, err)
	//agentHandler.Println(logs)
	//assert.NotNil(t, logs)

	fmt.Println(list)
	os.RemoveAll(list.Functions[0].Dir)
	os.Remove(list.Functions[0].FilePath)

}
