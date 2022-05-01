package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
)

const (
	dockerRemote      = "ishan27g/ryo-faas:"
	versionStr        = ".v0.1"
	databaseImage     = dockerRemote + "rfa-database" + versionStr
	proxyImage        = dockerRemote + "rfa-proxy" + versionStr
	deployBaseImage   = dockerRemote + "rfa-deploy-base" + versionStr
	natsImage         = "nats"
	natsVersion       = ":alpine3.15"
	natsContainerName = "rfa-" + natsImage

	zipKinRep      = "openzipkin/"
	zipkinImage    = "zipkin"
	zipkinVersion  = ":2.23.15"
	zipkinHostPort = "9411"

	jaegerImage     = "jaegertracing/all-in-one"
	jaegerVersion   = ":1.31"
	jaegerHostPort1 = "14268"
	jaegerHostPort2 = "16686"

	networkName = "rfa_nw"
	natsNwHost  = "nats://rfa-nats:4222"

	databaseHostRpcPort   = "5000"
	databaseHostHttpPort  = "5001"
	proxyRpcHostPort      = "9998"
	proxyHttpHostPort     = "9999"
	deployedFnNetworkPort = "6000"

	natsHostPort1 = "4222"
	natsHostPort2 = "8222"

	localTimeout  = 30 * time.Second
	remoteTimeout = 100 * time.Second

	defaultProvider = "JAEGER="
)

var defaultEnv = []string{"DATABASE=" + databaseNwHost(), "NATS=" + natsNwHost, defaultProvider + defaultProviderHost()}

func defaultProviderHost() string {
	if defaultProvider == "ZIPKIN=" {
		return zipkinContainerName()
	}
	if defaultProvider == "JAEGER=" {
		return jaegerContainerName()
	}
	return ""
}

func jaegerContainerName() string {
	return "rfa-jaeger"
}

var labels = map[string]string{
	"rfa": "faas",
}

var trimVersion = func(from string) string {
	return strings.TrimPrefix(strings.TrimSuffix(from, versionStr), dockerRemote)
}

func databaseContainerName() string {
	return trimVersion(databaseImage)
}

func proxyContainerName() string {
	return trimVersion(proxyImage)
}

func serviceImageName(serviceName string) string {
	return "rfa-deploy-" + serviceName
}
func zipkinContainerName() string {
	return "rfa-" + zipkinImage
}
func asFilter(key, val string) filters.Args {
	filter := filters.NewArgs()
	filter.Add(key, val)
	return filter
}

type docker struct {
	forcePull    bool
	silent       bool
	isProxyLocal bool
	*client.Client
	*log.Logger
}

func (d *docker) PruneImages() bool {
	return d.pruneImages()
}

func (d *docker) CheckImages() bool {
	return d.checkImages()
}

func (d *docker) ensureNetwork() bool {
	ctx := context.Background()
	_, err := d.NetworkCreate(ctx, networkName, types.NetworkCreate{
		CheckDuplicate: true,
		Labels:         labels,
	})
	if err != nil {
		if strings.Contains(err.Error(), networkName+" already exists") {
			return true
		}
		fmt.Printf("\nUnable to create network %s: \n", networkName)
		return false
	}
	return true
}
func (d *docker) pruneImages() bool {
	ctx := context.Background()
	list, err := d.ImageList(ctx, types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return false
	}
	for _, summary := range list {
		if summary.Labels["label"] == "rfa" {
			d.ImageRemove(ctx, summary.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
		}
	}
	return true
}
func (d *docker) pruneFunctionImages(name string) bool {
	ctx := context.Background()
	list, err := d.ImageList(ctx, types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return false
	}
	for _, summary := range list {
		if summary.Labels["rfa"] == name {
			d.ImageRemove(ctx, summary.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
		}
	}
	return true
}
func (d *docker) stop(name string) error {
	ctx := context.Background()
	if err := d.ContainerStop(ctx, name, nil); err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			fmt.Printf("\nUnable to stop container %s: %s\n", name, err)
			return err
		}
	}
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := d.ContainerRemove(ctx, name, removeOptions); err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			fmt.Printf("\nUnable to remove container %s: \n", networkName)
			return err
		}
	}
	return nil
}
func (d *docker) stopZipkin() error {
	name := zipkinContainerName()
	return d.stop(name)
}
func (d *docker) stopJaeger() error {
	name := jaegerContainerName()
	return d.stop(name)
}
func (d *docker) stopProxy() error {
	name := proxyContainerName()
	return d.stop(name)
}

func (d *docker) stopDatabase() error {
	name := trimVersion(databaseImage)
	return d.stop(name)
}

func (d *docker) checkAllRfa() []types.Container {
	ctx := context.Background()
	containers, err := d.ContainerList(ctx, types.ContainerListOptions{Filters: asFilter("label", "rfa")})
	if err != nil {
		panic(err)
	}
	return containers
}

func (d *docker) check(args filters.Args) string {
	ctx := context.Background()
	containers, err := d.ContainerList(ctx, types.ContainerListOptions{Filters: args})
	if err != nil {
		panic(err)
	}
	if len(containers) > 0 {
		return containers[0].State
	}
	return ""
}

func (d *docker) stopNats() error {
	name := "rfa-" + natsImage
	return d.stop(name)
}

func (d *docker) imageBuild(dockerClient *client.Client, serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	// todo
	// remove
	//if d.checkImage(serviceName) {
	//	return nil
	//}

	tar, err := archive.TarWithOptions("./", &archive.TarOptions{})
	if err != nil {
		fmt.Println("tar err ", err.Error())
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile:     "deploy.dockerfile",
		Tags:           []string{serviceName},
		Labels:         map[string]string{"label": "rfa", "rfa": serviceName},
		Remove:         true,
		ForceRemove:    true,
		SuppressOutput: d.silent,
	}
	res, err := dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		fmt.Println("build error")
		return err
	}
	defer res.Body.Close()
	if !d.silent {
		err = d.print(res.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *docker) print(rd io.Reader) error {
	var lastLine string

	type ErrorDetail struct {
		Message string `json:"message"`
	}
	type ErrorLine struct {
		Error       string      `json:"error"`
		ErrorDetail ErrorDetail `json:"errorDetail"`
	}

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

func (d *docker) pull(ctx context.Context, refStr string) error {
	fmt.Println("Pulling image", refStr)
	out, err := d.ImagePull(ctx, refStr, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	//if !d.silent {
	io.Copy(os.Stdout, out)
	//}
	return nil
}

func (d *docker) checkImage(imageName string) bool {
	if d.forcePull {
		return false
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(localTimeout))
	defer cancel()
	list, err := d.ImageList(ctx, types.ImageListOptions{Filters: asFilter("reference", imageName)})
	if err != nil {
		d.Println(err.Error())
		return false
	}
	//for _, summary := range list {
	//	d.Println("checkImage - ", summary.ID)
	//}
	return len(list) == 1
}
func (d *docker) checkImages() bool {
	if !d.checkImage(deployBaseImage) {
		return false
	}
	//if !d.checkImage(zipKinRep + zipkinImage + zipkinVersion) {
	//	return false
	//}
	if !d.checkImage(databaseImage) {
		return false
	}
	if !d.checkImage(proxyImage) {
		return false
	}
	if !d.checkImage(natsImage + natsVersion) {
		return false
	}

	return true
}
func (d *docker) ensureImages() bool {
	var wg sync.WaitGroup
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(remoteTimeout))
	defer cancel()

	var errs = make(chan error, 5)
	// var errs = make(chan error, 5)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !d.checkImage(deployBaseImage) {
			if err := d.pull(ctx, deployBaseImage); err != nil {
				d.Println("error-pull", deployBaseImage, err.Error())
				errs <- err
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !d.checkImage(jaegerImage + jaegerVersion) {
			if err := d.pull(ctx, jaegerImage+jaegerVersion); err != nil {
				d.Println("error-pull", jaegerImage+jaegerVersion, err.Error())
				errs <- err
			}
		}
	}()

	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	if !d.checkImage(zipKinRep + zipkinImage + zipkinVersion) {
	//		if err := d.pull(ctx, zipKinRep+zipkinImage+zipkinVersion); err != nil {
	//			d.Println("error-pull", zipKinRep+zipkinImage+zipkinVersion, err.Error())
	//			errs <- err
	//		}
	//	}
	//}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !d.checkImage(databaseImage) {
			if err := d.pull(ctx, databaseImage); err != nil {
				d.Println("error-pull", databaseImage, err.Error())
				errs <- err
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !d.checkImage(proxyImage) {
			if err := d.pull(ctx, proxyImage); err != nil {
				d.Println("error-pull", proxyImage, err.Error())
				errs <- err
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !d.checkImage(natsImage + natsVersion) {
			if err := d.pull(ctx, natsImage+natsVersion); err != nil {
				d.Println("error-pull", natsImage+natsVersion, err.Error())
				errs <- err
			}
		}
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	return true
}

func (d *docker) startJaeger() error {
	ctx := context.Background()

	name := jaegerContainerName()
	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: jaegerImage + jaegerVersion, Hostname: name, Labels: labels}

	// bind container port to host port
	hostBinding1 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: jaegerHostPort1,
	}
	containerPort1, err := nat.NewPort("tcp", jaegerHostPort1)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	// bind container port to host port
	hostBinding2 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: jaegerHostPort2,
	}
	containerPort2, err := nat.NewPort("tcp", jaegerHostPort2)
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

	return nil

}
func (d *docker) startZipkin() error {
	ctx := context.Background()

	name := zipkinContainerName()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: zipKinRep + zipkinImage + zipkinVersion, Hostname: name, Labels: labels}

	// bind container port to host port
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: zipkinHostPort,
	}
	containerPort, err := nat.NewPort("tcp", zipkinHostPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	hostConfig.PortBindings = nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

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

	return nil
}

func (d *docker) startNats() error {
	ctx := context.Background()

	name := "rfa-" + natsImage

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: natsImage + natsVersion, Hostname: name, Labels: labels}

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

	return nil
}

func (d *docker) startProxy() error {
	ctx := context.Background()

	name := proxyContainerName()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: proxyImage, Hostname: name, ExposedPorts: map[nat.Port]struct{}{
		proxyHttpHostPort + "/tcp": {},
		proxyRpcHostPort + "/tcp":  {},
	}, Env: defaultEnv, Labels: labels}

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
	return nil
}

func (d *docker) startDatabase() error {
	ctx := context.Background()

	var config = new(container.Config)
	var hostConfig = new(container.HostConfig)
	var networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	config = &container.Config{Image: databaseImage, Hostname: trimVersion(databaseImage), Env: defaultEnv, Labels: labels, ExposedPorts: map[nat.Port]struct{}{
		databaseHostRpcPort + "/tcp":  {},
		databaseHostHttpPort + "/tcp": {},
	}}

	// bind container port to host port
	hostBinding1 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: databaseHostRpcPort,
	}
	hostBinding2 := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: databaseHostHttpPort,
	}
	containerPort1, err := nat.NewPort("tcp", databaseHostRpcPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	containerPort2, err := nat.NewPort("tcp", databaseHostHttpPort)
	if err != nil {
		d.Println("Unable to get the port", err.Error())
		return err
	}
	hostConfig.PortBindings = nat.PortMap{containerPort1: []nat.PortBinding{hostBinding1},
		containerPort2: []nat.PortBinding{hostBinding2}}

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

	return nil
}
