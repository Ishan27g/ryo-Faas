package registry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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
var registryDir = "/registry"

const FnFw = "/deploy/"

var pathToDeployment = path() + registryDir + FnFw
var PathToFns = pathToDeployment + "functions/"

var getGenFilePath = func(fileName string) string {

	return PathToFns + strings.ToLower(fileName) + "_generated" + strconv.Itoa(rand.Intn(10000)) + ".go"
}
var modFile = func() string {
	return path() + registryDir + "/template/template.go"
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
	rsp.Functions = r
	return rsp, nil
}

func (a *AgentHandler) List(ctx context.Context, empty *deploy.Empty) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	rsp := a.registry.list(empty)
	return rsp, nil
}

func (a *AgentHandler) Stop(ctx context.Context, request *deploy.Empty) (*deploy.DeployResponse, error) {
	defer timeIt(time.Now())
	var rsp = new(deploy.DeployResponse)
	rsp.Functions = a.registry.stopped(&deploy.Function{Entrypoint: request.GetEntrypoint()})
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

	dir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		fmt.Println("cannot mkdir temp", err.Error())
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
		if strings.Contains(err.Error(), "file already exists") {
			err := os.RemoveAll(PathToFns + fname)
			if err != nil {
				fmt.Println("remove error ", err.Error())
				return err
			}
			err = archiver.Unarchive(tmpZip, PathToFns)
			if err != nil {
				fmt.Println("Twice Un-archive error ", err.Error())
				return err
			}
		}
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

	agent := new(AgentHandler)
	agent.registry = new(registry)
	agent.address = rpcAddress
	agent.Logger = log.New(os.Stdout, "[AGENT-HANDLER]", log.Ltime)
	*agent.registry = setup(rpcAddress)
	os.RemoveAll(PathToFns)
	os.MkdirAll(PathToFns, os.ModePerm)

	return agent
}
