package main

import (
	"context"
	"encoding/json"
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

func printJson(js interface{}) {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(data))
}
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
	<-time.After(1 * time.Second)

	type export = struct {
		dir        string
		filepath   string
		entrypoint string
	}
	var exports = []export{
		{
			dir:        "../examples/method1",
			filepath:   "../examples/method1/method1.go",
			entrypoint: "Method1",
		}, {
			dir:        "../examples/method2",
			filepath:   "../examples/method2/method2.go",
			entrypoint: "Method2",
		},
	}
	c := transport.ProxyGrpcClient(DefaultPort)
	for _, e := range exports {
		uploaded := transport.UploadDir(c, ctx, e.dir, e.entrypoint)
		assert.True(t, uploaded)
	}
	var fns []*deploy.Function
	for _, e := range exports {
		fns = append(fns, &deploy.Function{
			Entrypoint:       e.entrypoint,
			FilePath:         e.filepath,
			Dir:              e.dir,
			Zip:              "",
			AtAgent:          "",
			ProxyServiceAddr: "",
			Url:              "",
			Status:           "",
		})
	}
	deployRsp, err := c.Deploy(ctx, &deploy.DeployRequest{Functions: fns})
	assert.NoError(t, err)
	<-time.After(3 * time.Second)
	agentHandler.Println(deployRsp)
	for _, e := range exports {
		list, err := c.List(ctx, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: e.entrypoint}})
		assert.NoError(t, err)
		assert.Equal(t, "DEPLOYED", list.Functions[0].Status)
		printJson(list)
		for _, function := range list.Functions {
			if e.entrypoint == function.Entrypoint {
				os.RemoveAll(function.Dir)
			}
		}
	}
}
