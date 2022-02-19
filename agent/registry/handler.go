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

var importPath = "github.com/Ishan27g/ryo-Faas/agent/registry/funcFrameworkWrapper/"

var path = func() string {
	cwd, _ := os.Getwd()
	// if cwd == "/"{
	// 	return "/remote"
	// }
	return cwd //+ "/remote"
}
var agentDir = "/registry"

const FnFw = "/funcFrameworkWrapper/"

var pathToFnFw = path() + agentDir + FnFw
var unzipDir = pathToFnFw

func defaultPath() string {
	unzipDir = pathToFnFw
	return path() + agentDir + FnFw
}

var getGenFilePath = func(fileName string) string {
	return pathToFnFw + strings.ToLower(fileName) + "_generated.go"
}
var modFile = func() string {
	return pathToFnFw + "template.go"
}
var deployFile = func() string {
	return pathToFnFw + "deploy.go"
}

type AgentHandler struct {
	ctx    context.Context
	cancel context.CancelFunc

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

func (a *AgentHandler) Stop(ctx context.Context, request *deploy.DeployRequest) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	var rsp = new(deploy.DeployResponse)
	function := request.Functions
	rsp.Functions = append(rsp.Functions, a.registry.stopped(function.Entrypoint))
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

	dir, err := os.MkdirTemp("/app", "")
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
	fmt.Println("unzipping ", fileName, " to ", unzipDir)
	err = archiver.Unarchive(tmpZip, unzipDir)
	if err != nil {
		fmt.Println("Un-archive error ", err.Error())
		return err
	}
	_, fname := filepath.Split(fileName)
	unzipTo := unzipDir + strings.TrimSuffix(fname, ".zip") + "/"
	a.registry.upload(entrypoint, unzipTo)
	return nil

}

func (a *AgentHandler) Logs(ctx context.Context, function *deploy.Function) (*deploy.Logs, error) {
	defer timeIt(time.Now())
	fn := &deploy.Empty_Entrypoint{Entrypoint: function.Entrypoint}
	list := a.registry.list(&deploy.Empty{Rsp: fn})
	logs := a.registry.Logs(function.Entrypoint)
	return &deploy.Logs{
		Fn:   list.Functions[0],
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

	return agent
}
