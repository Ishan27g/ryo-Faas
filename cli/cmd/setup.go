package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/ryo-Faas/docker"
	"github.com/urfave/cli/v2"
)

var stopRyoFaas = cli.Command{
	Name:            "stopFaas",
	Usage:           "stop ryo-Faas",
	ArgsUsage:       "server-cli stopFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StopDatabase()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StopProxy()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StopNats()
		}()
		wg.Wait()
		fmt.Println("Stopped ryo-Faas")
		return nil
	},
}
var startRyoFaas = cli.Command{
	Name:            "startFaas",
	Usage:           "start ryo-Faas",
	ArgsUsage:       "server-cli startFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		d := docker.New()
		//if err := d.Pull(); err != nil {
		//	fmt.Println(err.Error())
		//	return cli.Exit("Cannot pull images from remote", 1)
		//}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StartDatabase()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StartProxy()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			docker.New().StartNats()
		}()
		wg.Wait()
		fmt.Println("Started ryo-Faas : Proxy running at http://localhost:9999")
		<-time.After(500 * time.Millisecond)
		d.Status()
		return nil
	},
}
