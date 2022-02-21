package registry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/mholt/archiver/v3"
)

var importPath = "github.com/Ishan27g/ryo-Faas/agent/registry/deploy/functions/"

var path = func() string {
	cwd, _ := os.Getwd()
	// if cwd == "/"{
	// 	return "/remote"
	// }
	return cwd //+ "/remote"
}
var agentDir = "/registry"

const FnFw = "/deploy/"

var pathToDeployment = path() + agentDir + FnFw
var PathToFns = pathToDeployment + "functions/"

// func defaultPath() string {
// 	unzipDir = pathToDeployment
// 	return path() + agentDir + FnFw
// }

var getGenFilePath = func(fileName string) string {
	return PathToFns + strings.ToLower(fileName) + "_generated.go"
}
var modFile = func() string {
	return pathToDeployment + "template.go"
}

type AgentHandler struct {
	*registry
	*log.Logger
}

func timeIt(since time.Time) {
	fmt.Println("\n----- took : ", time.Since(since).String())
}
func (a *AgentHandler) Deploy(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	var rsp = new(deploy.DeployResponse)
	r := a.registry.deploy(request.Functions)
	rsp.Functions = append(rsp.Functions, r)
	return rsp, nil
}

func (a *AgentHandler) List(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	return a.registry.list(empty), nil
}

func (a *AgentHandler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	var rsp = new(deploy.DeployResponse)
	rsp.Functions = append(rsp.Functions, a.registry.stopped(request.GetEntrypoint()))
	return rsp, nil
}

func (a *AgentHandler) Details(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	return a.registry.list(empty), nil
}

func (a *AgentHandler) Upload(stream deploy.Deploy_UploadServer) error {
	defer timeIt(time.Now())
	var fileName string
	var entrypoint string
	imageData := bytes.Buffer{}

	for {
		ch, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				goto END
			}
			return err
		}
		chunk := ch.GetContent()
		_, err = imageData.Write(chunk)
		if err != nil {
			fmt.Println(err.Error())
		}
		fileName = ch.FileName
		entrypoint = ch.Entrypoint
	}
END:
	err := stream.SendAndClose(&deploy.Empty{Rsp: nil})
	fmt.Println("END: ", err)

	dir, err := os.MkdirTemp("/app/agent", "")
	if err != nil {
		fmt.Println("cannot mkdir temp")
		return err
	}

	tmpZip := dir + "tmp.zip"
	file, err := os.OpenFile(tmpZip, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("cannot create image file: %w", err)
		return err
	}
	defer file.Close()
	defer func() {
		os.Remove(file.Name())
		os.RemoveAll(dir)
	}()

	_, err = imageData.WriteTo(file)
	if err != nil {
		fmt.Println("cannot write image to file ", fileName, err.Error())
		return err
	}
	_, fname := filepath.Split(fileName)
	unzipTo := PathToFns + strings.TrimSuffix(fname, ".zip") + "/"

	fmt.Println("unzipping ", fileName, " to ", PathToFns)
	err = archiver.Unarchive(tmpZip, PathToFns)
	if err != nil {
		fmt.Println("Un-archive error ", err.Error())
		return err
	}

	a.registry.upload(entrypoint, unzipTo)
	return nil

}

func (a *AgentHandler) Logs(ctx context.Context, function *deploy.Function) (*deploy.Logs, error) {
	defer timeIt(time.Now())
	logs := a.registry.system.logs(function.Entrypoint)
	return &deploy.Logs{
		Data: logs,
	}, nil
}

func Init(rpcAddress string) *AgentHandler {
	// pathToFnFw = defaultPath()

	fmt.Println(rpcAddress)
	agent := new(AgentHandler)
	agent.registry = new(registry)
	agent.address = rpcAddress
	agent.Logger = log.New(os.Stdout, "[AGENT-HANDLER]", log.Ltime)
	*agent.registry = setup(rpcAddress)
	agent.Println("AgentInterface configured at ", agent.address)

	os.Mkdir(PathToFns, os.ModePerm)

	return agent
}
