package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	agentImage    = "rfa-agent:v0.1"
	databaseImage = "rfa-database:v0.1"
	proxyImage    = "rfa-proxy:v0.1"
	natsImage     = "nats"
	natsVersion   = ":alpine"

	actionTimeout = 30 * time.Second
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

var trimVersion = func(versionStr string) string {
	return strings.TrimSuffix(versionStr, ":v0.1")
}

type Docker interface {
	Pull() error

	StartNats() error
	StopNats() error

	StartDatabase() error
	StopDatabase() error

	StartAgent(instance string) error
	StopAgent(instance string) error

	StartProxy() error
	StopProxy() error
}
type docker struct {
	*client.Client
	*log.Logger
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

	name := trimVersion(proxyImage)

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

func (d *docker) StopProxy() error {
	name := trimVersion(proxyImage)
	return d.stop(name)
}

func (d *docker) StartAgent(instance string) error {
	ctx := context.Background()

	name := agentInstanceName(instance)

	dbAddress := trimVersion(databaseImage) + ":" + databaseHostRpcPort

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: agentImage, Hostname: name, ExposedPorts: agentPorts, Env: []string{"DATABASE=" + dbAddress}}

	// bind container port to host port
	//hostBinding := nat.PortBinding{
	//	HostIP:   "0.0.0.0",
	//	HostPort: agentHostPost,
	//}
	//containerPort, err := nat.NewPort("tcp", agentHostPost)
	//if err != nil {
	//	d.Println("Unable to get the port", err.Error())
	//	return err
	//}
	//hostConfig.PortBindings = nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

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

func agentInstanceName(instance string) string {
	return trimVersion(agentImage) + "-" + instance
}

func (d *docker) StopAgent(instance string) error {
	name := agentInstanceName(instance)
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

	resp, err := d.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, trimVersion(databaseImage))
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

func (d *docker) Pull() error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(actionTimeout))
	defer cancel()
	if err := d.pull(ctx, agentImage); err != nil {
		d.Println("error-pull", agentImage, err.Error())
		return err
	}
	if err := d.pull(ctx, databaseImage); err != nil {
		d.Println("error-pull", databaseImage, err.Error())
		return err
	}
	if err := d.pull(ctx, natsImage+natsVersion); err != nil {
		d.Println("error-pull", natsImage+natsVersion, err.Error())
		return err
	}
	if err := d.pull(ctx, proxyImage); err != nil {
		d.Println("error-pull", proxyImage, err.Error())
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

func New() Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}
	return &docker{cli, log.New(os.Stdout, "docker", log.LstdFlags)}
}
