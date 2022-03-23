package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	cp "github.com/otiai10/copy"
	"github.com/pkg/errors"
)

const (
	dockerRemote = "ishan27g/ryo-faas:"
	versionStr   = ".v0.1"
	//agentImage    = dockerRemote + "rfa-agent" + versionStr
	databaseImage   = dockerRemote + "rfa-database" + versionStr
	proxyImage      = dockerRemote + "rfa-proxy" + versionStr
	deployBaseImage = dockerRemote + "rfa-deploy-base" + versionStr
	natsImage       = "nats"
	natsVersion     = ":alpine"

	actionTimeout = 150 * time.Second
	networkName   = "rfa_nw"

	databaseHostRpcPort = "5000"
	agentHostPost       = "9000"
	proxyRpcHostPort    = "9998"
	proxyHttpHostPort   = "9999"

	natsHostPort1 = "4222"
	natsHostPort2 = "8222"
)

var agentPorts = map[nat.Port]struct{}{
	"9000/tcp": {},
	"6000/tcp": {},
	"6001/tcp": {},
	"6002/tcp": {},
	"6003/tcp": {},
	"6004/tcp": {},
}

var trimVersion = func(from string) string {
	return strings.TrimPrefix(strings.TrimSuffix(from, versionStr), dockerRemote)
}

type Docker interface {
	Pull() error
	StatusAll() bool
	StatusAny() bool
	Check() map[string]string
	StartNats() error
	StopNats() error

	StartDatabase() error
	StopDatabase() error

	StartProxy() error
	StopProxy() error

	RunFunction(serviceName string) error
	StopFunction(serviceName string) error
}
type docker struct {
	*client.Client
	*log.Logger
}

func (d *docker) StopFunction(serviceName string) error {
	name := serviceContainerName(serviceName)
	return d.stop(name)
}

func imageBuild(dockerClient *client.Client, serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	tar, err := archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "deploy.dockerfile",
		Tags:       []string{serviceName},
		Remove:     true,
	}
	res, err := dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		fmt.Println("build error")
		return err
	}

	defer res.Body.Close()

	err = print(res.Body)
	if err != nil {
		return err
	}

	return nil
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func print(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
func tmpDockerFile(entrypointFile string) bool {

	err := cp.Copy("deploy.dockerfile", "deploy.tmp.dockerfile")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	tmp, err := os.OpenFile("deploy.tmp.dockerfile",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return false
	}
	defer tmp.Close()
	var entrypoint = "ENTRYPOINT [\"./deployments" + "\"," + "\"--port\", \"6000\"]"
	fmt.Println(entrypoint)

	if _, err := tmp.WriteString("\n" + entrypoint + "\n"); err != nil {
		log.Println(err)
		return false
	}
	tmp.Sync()
	return true
}
func (d *docker) RunFunction(serviceName string) error {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	name := serviceContainerName(serviceName)
	//
	err = imageBuild(d.Client, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	ctx := context.Background()

	dbAddress := trimVersion(databaseImage) + ":" + databaseHostRpcPort

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	ports := map[nat.Port]struct{}{
		"6000/tcp": {},
	}
	config = &container.Config{Image: name, Hostname: name, ExposedPorts: ports, Env: []string{"DATABASE=" + dbAddress}}

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

	fmt.Println(resp.ID)
	return nil

}

func serviceContainerName(serviceName string) string {
	return "rfa-deploy-" + serviceName
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
		if s2 == "running" {
			return true
		}
	}
	return anyRunning
}

func (d *docker) StartNats() error {
	ctx := context.Background()

	name := "rfa-" + natsImage

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: natsImage + natsVersion, Hostname: name}

	// bind container port to host port
	hostBinding1 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: natsHostPort1,
	}
	hostBinding2 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: natsHostPort2,
	}
	containerPort1, err := nat.NewPort("tcp", natsHostPort1)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	containerPort2, err := nat.NewPort("tcp", natsHostPort2)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	hostConfig.PortBindings = nat.PortMap{containerPort1: []nat.PortBinding{hostBinding1}, containerPort2: []nat.PortBinding{hostBinding2}}

	// attach container to network
	networkingConfig.EndpointsConfig[networkName] = &network.EndpointSettings{}

	resp, err := d.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := d.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(resp.ID)
	return nil
}

func (d *docker) StopNats() error {
	name := "rfa-" + natsImage
	return d.stop(name)
}

func (d *docker) StartProxy() error {
	ctx := context.Background()

	name := proxyContainerName()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: proxyImage, Hostname: name, ExposedPorts: map[nat.Port]struct{}{
		"9999/tcp": {},
		"9998/tcp": {},
	}}

	// bind container port to host port
	hostHttpBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: proxyHttpHostPort,
	}
	hostRpcBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: proxyRpcHostPort,
	}
	containerHttpPort, err := nat.NewPort("tcp", proxyHttpHostPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	containerRpcPort, err := nat.NewPort("tcp", proxyRpcHostPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	hostConfig.PortBindings = nat.PortMap{
		containerHttpPort: []nat.PortBinding{hostHttpBinding},
		containerRpcPort:  []nat.PortBinding{hostRpcBinding},
	}
	hostConfig.Binds = []string{"/var/run/docker.sock:/var/run/docker.sock"}

	// attach container to network
	networkingConfig.EndpointsConfig[networkName] = &network.EndpointSettings{}

	resp, err := d.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := d.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(resp.ID)
	return nil
}

func proxyContainerName() string {
	return trimVersion(proxyImage)
}

func (d *docker) StopProxy() error {
	name := proxyContainerName()
	return d.stop(name)
}

func (d *docker) StopDatabase() error {
	name := trimVersion(databaseImage)
	return d.stop(name)
}

func (d *docker) stop(name string) error {
	ctx := context.Background()
	if err := d.ContainerStop(ctx, name, nil); err != nil {
		log.Printf("Unable to stop container %s: %s", name, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := d.ContainerRemove(ctx, name, removeOptions); err != nil {
		log.Printf("Unable to remove container: %s", err)
		return err
	}
	return nil
}
func (d *docker) StartDatabase() error {
	ctx := context.Background()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: databaseImage, Hostname: trimVersion(databaseImage)}

	// bind container port to host port
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: databaseHostRpcPort,
	}
	containerPort, err := nat.NewPort("tcp", databaseHostRpcPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	hostConfig.PortBindings = nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

	// attach container to network
	networkingConfig.EndpointsConfig[networkName] = &network.EndpointSettings{}

	resp, err := d.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, databaseContainerName())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := d.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(resp.ID)
	return nil
}

func databaseContainerName() string {
	return trimVersion(databaseImage)

}

func (d *docker) Pull() error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(actionTimeout))
	defer cancel()
	if err := d.pull(ctx, databaseImage); err != nil {
		d.Println("error-pull", databaseImage, err.Error())
		return err
	}
	if err := d.pull(ctx, proxyImage); err != nil {
		d.Println("error-pull", proxyImage, err.Error())
		return err
	}
	if err := d.pull(ctx, natsImage+natsVersion); err != nil {
		d.Println("error-pull", natsImage+natsVersion, err.Error())
		return err
	}
	if err := d.pull(ctx, deployBaseImage); err != nil {
		d.Println("error-pull", natsImage+natsVersion, err.Error())
		return err
	}
	return nil
}
func (d *docker) pull(ctx context.Context, refStr string) error {
	out, err := d.ImagePull(ctx, refStr, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}

func (d *docker) Check() map[string]string {
	var status = make(map[string]string)
	status["rfa-"+natsImage] = d.check("rfa-" + natsImage)
	status[databaseContainerName()] = d.check(databaseContainerName())
	status[proxyContainerName()] = d.check(proxyContainerName())
	return status
}
func (d *docker) check(containerName string) string {
	ctx := context.Background()

	filter := filters.NewArgs()
	filter.Add("name", containerName)
	containers, err := d.ContainerList(ctx, types.ContainerListOptions{Filters: filter})
	if err != nil {
		panic(err)
	}
	if len(containers) > 0 {
		return containers[0].State
	}
	return ""
}

func New() Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}
	return &docker{cli, log.New(os.Stdout, "docker", log.LstdFlags)}
}
