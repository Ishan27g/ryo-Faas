package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ishan27g/ryo-Faas/docker"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/urfave/cli/v2"
)

var stopRyoFaas = cli.Command{
	Name:            "stopFaas",
	Usage:           "stop ryo-Faas",
	Aliases:         []string{"sto"},
	ArgsUsage:       "server-cli stopFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		d := docker.New()
		d.Stop()
		if !d.StatusAny() {
			return nil
		}
		var rsp []types.FunctionJsonRsp
		json.Unmarshal(sendHttp("/details", ""), &rsp)
		for _, v := range rsp {
			fmt.Printf("%s\t\t%s\t%s\n", v.Url, v.Name, v.Status)
			d.StopFunction(v.Name)
		}

		fmt.Println("Stopped ryo-Faas")
		return nil
	},
}
var startRyoFaas = cli.Command{
	Name:            "startFaas",
	Aliases:         []string{"sta"},
	Usage:           "start ryo-Faas",
	ArgsUsage:       "server-cli startFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		d := docker.New()
		if !d.Start() {
			fmt.Println("Unable to start")
			return nil
		}
		fmt.Println("Started ryo-Faas : Proxy running at http://localhost:9999")
		<-time.After(500 * time.Millisecond)
		d.StatusAll()
		return nil
	},
}
