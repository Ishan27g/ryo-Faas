package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
)

type registry struct {
	address   string
	portStart int
	portLimit int
	ports     map[string]bool
	functions map[string]struct {
		port string
		*deploy.Function
	}
	system *system
	*log.Logger
}

func prettyJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}
func setup(atAgent string) registry {
	//	SetBuildCommand(nil)

	var err error
	var reg registry
	reg.portStart, err = strconv.Atoi(os.Getenv("PORT_START"))
	if err != nil {
		reg.portStart = 5000
	}
	reg.portLimit, err = strconv.Atoi(os.Getenv("NUM_PORTS"))
	if err != nil {
		reg.portLimit = 20
	}
	reg.Logger = log.New(os.Stdout, "[REGISTRY]", log.Ltime)
	reg.address = atAgent
	reg.ports = make(map[string]bool)
	for i := reg.portStart; i < reg.portStart+reg.portLimit; i++ {
		reg.ports[strconv.Itoa(i)] = true // free ports
	}
	reg.functions = make(map[string]struct {
		port string
		*deploy.Function
	})
	reg.system = newSystem()
	reg.Logger.Println("initialised at AgentHandler-", reg)
	return reg

}

func (r *registry) nextPort() string {
	var port string
	for p, free := range r.ports {
		if free {
			r.ports[p] = false
			port = p
			return port
		}
	}
	return ""
}
func (r *registry) deployed(fnName string) {
	fn := r.functions[fnName]
	fn.Status = "DEPLOYED"
	r.functions[fnName] = fn
}
func (r *registry) stopped(fnName string) *deploy.Function {

	r.system.stop(fnName)
	port := r.functions[fnName].port
	r.ports[port] = true

	os.RemoveAll(r.functions[fnName].Dir)
	os.Remove(r.functions[fnName].FilePath)

	fn := &deploy.Function{Entrypoint: fnName, Status: "STOPPED", AtAgent: r.address}

	delete(r.functions, fnName)
	r.functions[fnName] = struct {
		port string
		*deploy.Function
	}{"", fn}

	return fn
}

func (r *registry) list(rFn *deploy.Empty) *deploy.DeployResponse {
	var rsp = new(deploy.DeployResponse)
	if rFn.GetEntrypoint() == "" {
		for _, fn := range r.functions {
			rsp.Functions = append(rsp.Functions, fn.Function)
		}
	} else {
		fn := r.functions[rFn.GetEntrypoint()]
		rsp.Functions = append(rsp.Functions, fn.Function)
	}
	return rsp
}
func (r *registry) deploy(rFn *deploy.Function) *deploy.Function {
	entryPoint := rFn.Entrypoint

	hn := "localhost"
	port := r.nextPort()
	uFn := r.functions[entryPoint]
	if uFn.Status == "" {
		r.Println(entryPoint, "not uploaded")
		return nil
	}
	_, file := filepath.Split(rFn.GetFilePath())
	r.Println("File name is ", file)
	uFn.FilePath = uFn.Dir + file

	r.Println(prettyJson(uFn.Function))
	valid, genFile := astLocalCopy(uFn.Function)
	if !valid {
		r.Println("invalid file ", uFn.Function)
		return nil
	}
	registered := &deploy.Function{
		Entrypoint:       entryPoint,
		Dir:              uFn.Dir,
		Zip:              uFn.Zip,
		AtAgent:          r.address,
		FilePath:         genFile,
		ProxyServiceAddr: "http://" + hn + ":" + port,
		Url:              "http://" + hn + ":" + port + "/" + strings.ToLower(entryPoint),
		Status:           "DEPLOYING",
	}

	r.functions[entryPoint] = struct {
		port string
		*deploy.Function
	}{port, registered}

	go func() {
		if r.system.run(registered, port) {
			r.deployed(entryPoint)
		} else {
			r.stopped(entryPoint)
		}
	}()
	r.Println("DEPLOYED", prettyJson(registered))
	return registered
}

func (r *registry) upload(entrypoint string, dir string) {
	registered := &deploy.Function{
		Entrypoint: entrypoint,
		Dir:        dir,
		Status:     "UPLOADED",
	}

	r.functions[entrypoint] = struct {
		port string
		*deploy.Function
	}{"", registered}

	r.Println("uploaded", entrypoint, "to", dir)
}
