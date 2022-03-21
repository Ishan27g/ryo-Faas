package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/shell"
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
	//system *system

	// systems   map[string]shell.Shell
	systemCmd map[string]context.CancelFunc
	logs      map[string]io.ReadCloser

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
	//reg.system = newSystem()

	reg.systemCmd = make(map[string]context.CancelFunc)
	reg.logs = make(map[string]io.ReadCloser)
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
func (r *registry) deployed(fns ...*deploy.Function) {
	for _, f := range fns {
		fn := r.functions[f.Entrypoint]
		fn.Status = "DEPLOYED"
		r.functions[fn.Entrypoint] = fn
	}
}
func (r *registry) stopped(fns ...*deploy.Function) []*deploy.Function {
	var rsp []*deploy.Function
	for _, f := range fns {
		if r.systemCmd[f.Entrypoint] != nil {
			r.systemCmd[f.Entrypoint]()
		}
		port := r.functions[f.Entrypoint].port
		r.ports[port] = true

		sendStop(r.functions[f.Entrypoint].ProxyServiceAddr)

		os.RemoveAll(r.functions[f.Entrypoint].Dir)
		os.Remove(r.functions[f.Entrypoint].FilePath)

		fn := &deploy.Function{Entrypoint: f.Entrypoint, Status: "STOPPED", AtAgent: r.address, Async: f.GetAsync()}

		delete(r.functions, f.Entrypoint)
		r.functions[f.Entrypoint] = struct {
			port string
			*deploy.Function
		}{"", fn}
		rsp = append(rsp, fn)

	}

	return rsp
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
func (r *registry) deploy(fns []*deploy.Function) []*deploy.Function {
	var uploadedFns []*deploy.Function
	for _, rFn := range fns {
		entryPoint := rFn.Entrypoint
		uFn := r.functions[entryPoint]
		if uFn.Status == "" {
			r.Println(entryPoint, "not uploaded")
			return nil
		}
		rFn.FilePath = uFn.Dir + filepath.Base(rFn.FilePath)
		uploadedFns = append(uploadedFns, rFn)
	}
	valid, genFile := astLocalCopy(uploadedFns)
	if !valid {
		r.Println("invalid file ")
		return nil
	}
	var registered []*deploy.Function
	hn := "localhost"
	port := r.nextPort()
	var entryPoint string
	for i, rFn := range fns {
		entryPoint = rFn.Entrypoint
		uFn := r.functions[entryPoint]
		if uFn.Status == "" {
			r.Println(entryPoint, "not uploaded")
			return nil
		}
		_, file := filepath.Split(rFn.GetFilePath())
		uFn.FilePath = uFn.Dir + file
		registered = append(registered, &deploy.Function{
			Entrypoint:       entryPoint,
			Dir:              uFn.Dir,
			Zip:              uFn.Zip,
			AtAgent:          r.address,
			FilePath:         genFile,
			ProxyServiceAddr: "http://" + hn + ":" + port,
			Url:              "http://" + hn + ":" + port + "/" + strings.ToLower(entryPoint),
			Status:           "DEPLOYING",
			Async:            uFn.GetAsync(),
		})
		r.functions[entryPoint] = struct {
			port string
			*deploy.Function
		}{port, registered[i]}

	}
	go func() {
		// run functions as one process
		pr, pw := io.Pipe()
		r.logs[registered[0].Entrypoint] = pr
		ctx, can := context.WithCancel(context.Background())
		r.systemCmd[registered[0].Entrypoint] = can
		if shell.CommandContext(ctx, registered[0].FilePath, port, pw).Run() != nil { // blocking
			// todo
		}
	}()
	<-time.After(3 * time.Second)
	for _, fn := range registered {
		if checkHealth(fn.ProxyServiceAddr) {
			r.deployed(fn)
		} else {
			r.stopped(fn)
		}
	}
	registered = nil
	for _, fn := range r.functions {
		registered = append(registered, fn.Function)
		// if fn.Async{
		// 	FuncFw.NewNatsAsync(fn.Entrypoint, fn.Url, )
		// }
	}
	r.Println("Deploy response", prettyJson(registered))
	return registered
}
func sendStop(addr string) bool {
	resp, err := http.Get(addr + "/stop")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}
func checkHealth(addr string) bool {
	resp, err := http.Get(addr + "/healthcheck")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
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
