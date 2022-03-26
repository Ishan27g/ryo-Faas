package docker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Docker interface {
	Setup() bool
	CheckImages() bool

	SetForcePull()
	SetSilent()
	SetLocalProxy()

	StatusAll() bool
	StatusAny() bool
	CheckLabel() bool

	Start() bool

	Stop() bool

	RunFunction(serviceName string) error
	CheckFunction(serviceName string) bool
	StopFunction(serviceName string) error
}

func (d *docker) Setup() bool {
	var wg sync.WaitGroup
	var done = make(chan bool, 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		done <- d.ensureNetwork()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		done <- d.ensureImages()
	}()
	wg.Wait()
	close(done)
	for b := range done {
		if !b {
			return false
		}
	}
	return true
}
func (d *docker) Start() bool {
	if !d.Setup() {
		return false
	}

	var errs = make(chan error, 4)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- d.startZipkin()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- d.startDatabase()
	}()
	if !d.isProxyLocal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- d.startProxy()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- d.startNats()
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return false
		}
	}
	return true
}

func (d *docker) Stop() bool {
	containers := d.checkAllRfa()
	for _, t := range containers {
		err := d.stop(t.Names[0])
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	//force
	d.stopDatabase()
	d.stopNats()
	d.stopProxy()
	d.stopZipkin()
	return true
}

func (d *docker) StopFunction(serviceName string) error {
	name := serviceContainerName(serviceName)
	return d.stop(name)
}

func (d *docker) StatusAll() bool {
	var allRunning = true
	for s, s2 := range d.Check() {
		fmt.Printf("\n%s:\t\t%s", s2, s)
		if s2 != "running" {
			allRunning = false
		}
	}
	fmt.Printf("\n")
	return allRunning
}

func (d *docker) StatusAny() bool {
	var anyRunning = false
	for _, s2 := range d.Check() {
		if s2 == "running" || s2 == "created" {
			return true
		}
	}
	return anyRunning
}

func (d *docker) Check() map[string]string {
	var status = make(map[string]string)
	status[natsContainerName] = d.check(asFilter("name", natsContainerName))
	status[databaseContainerName()] = d.check(asFilter("name", databaseContainerName()))
	status[proxyContainerName()] = d.check(asFilter("name", proxyContainerName()))
	status[zipkinContainerName()] = d.check(asFilter("name", zipkinContainerName()))
	return status
}
func (d *docker) CheckLabel() bool {
	containers := d.checkAllRfa()
	return containers[0].State == "running"
}
func (d *docker) CheckFunction(serviceName string) bool {
	ctx := context.Background()
	containers, err := d.ContainerList(ctx, types.ContainerListOptions{Filters: asFilter("name", serviceContainerName(serviceName))})
	if err != nil {
		panic(err)
	}
	return containers[0].State == "running"
}

func (d *docker) RunFunction(serviceName string) error {
	dir := os.Getenv("RYA_FAAS")
	if dir == "" {
		return errors.New("cannot find ryo-faas directory")
	}
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	name := serviceContainerName(serviceName)

	err = d.imageBuild(d.Client, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	ctx := context.Background()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	ports := map[nat.Port]struct{}{
		deployedFnNetworkPort + "/tcp": {},
	}

	config = &container.Config{Image: name, Hostname: name, ExposedPorts: ports, Env: defaultEnv, Labels: labels}

	// if proxy is running outside docker, expose fn-container port to host
	if d.isProxyLocal {
		// bind container port to host port
		hostBinding := nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: deployedFnNetworkPort,
		}
		containerPort, err := nat.NewPort("tcp", deployedFnNetworkPort)
		if err != nil {
			d.Println("Unable to get the port", err.Error())
			return err
		}
		hostConfig.PortBindings = nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	}

	// attach container to network
	networkingConfig.EndpointsConfig[networkName] = &network.EndpointSettings{}
	hostConfig.LogConfig = container.LogConfig{
		Type:   "json-file",
		Config: map[string]string{},
	}
	resp, err := d.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := d.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil

}

func databaseNwHost() string {
	return trimVersion(databaseImage) + ":" + databaseHostRpcPort
}
func (d *docker) SetLocalProxy() {
	d.isProxyLocal = true
}
func (d *docker) SetForcePull() {
	d.forcePull = true
}

func (d *docker) SetSilent() {
	d.silent = true
}
func New() Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}
	return &docker{false, false, false, cli, log.New(os.Stdout, "docker", log.LstdFlags)}
}
