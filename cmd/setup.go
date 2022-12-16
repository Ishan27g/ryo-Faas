package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/urfave/cli/v2"
)

var prune = false
var stopRyoFaas = cli.Command{
	Name:            "stopFaas",
	Usage:           "stop ryo-Faas",
	Aliases:         []string{"sto"},
	ArgsUsage:       "server-cmd stopFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "prune",
			Value:       false,
			Usage:       "prune function image",
			Destination: &prune,
		},
	},
	Action: func(c *cli.Context) error {
		d := docker.New()
		if !d.StatusAny() {
			return nil
		}
		var rsp []types.FunctionJsonRsp
		err := json.Unmarshal(sendHttp("/details"), &rsp)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		for _, v := range rsp {
			fmt.Printf("%s\t%s\t%s\n", v.Url, v.Name, v.Status)
			d.StopFunction(v.Name, prune)
		}
		d.Stop()
		fmt.Println("Stopped ryo-Faas")
		return nil
	},
}
var startRyoFaas = cli.Command{
	Name:            "startFaas",
	Aliases:         []string{"sta"},
	Usage:           "start ryo-Faas",
	ArgsUsage:       "server-cmd startFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		_, err := os.Stat(getDir())
		if err != nil && os.IsNotExist(err) {
			fmt.Println("run init command")
			return err
		}

		// todo --flag forcePull

		d := docker.New()

		if isProxyLocal() {
			d.SetLocalProxy()
		}
		if isDbLocal() {
			d.SetLocalDb()
		}

		if !d.Start() {
			d.Stop()
			fmt.Println("Unable to start")
			return nil
		}
		fmt.Println("Started ryo-Faas : Proxy running at http://localhost:9999")
		<-time.After(100 * time.Millisecond)
		d.StatusAll()
		return nil
	},
}

var pruneRyoFaas = cli.Command{
	Name:            "prune",
	Aliases:         []string{"p"},
	Usage:           "remove all images",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		d := docker.New()
		d.Stop()
		d.PruneImages()
		return nil
	},
}
