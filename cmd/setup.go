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

var stopRyoFaas = cli.Command{
	Name:            "stopFaas",
	Usage:           "stop ryo-Faas",
	Aliases:         []string{"sto"},
	ArgsUsage:       "server-cmd stopFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		d := docker.New()
		if !d.StatusAny() {
			return nil
		}
		var rsp []types.FunctionJsonRsp
		json.Unmarshal(sendHttp("/details", ""), &rsp)
		for _, v := range rsp {
			fmt.Printf("%s\t\t%s\t%s\n", v.Url, v.Name, v.Status)
			d.StopFunction(v.Name)
		}
		d.Stop()

		fmt.Println("Stopped ryo-Faas")
		return nil
	},
}
var runProxy bool
var startRyoFaas = cli.Command{
	Name:            "startFaas",
	Aliases:         []string{"sta"},
	Usage:           "start ryo-Faas",
	ArgsUsage:       "server-cmd startFaas",
	HideHelp:        false,
	HideHelpCommand: false,
	Flags: []cli.Flag{&cli.BoolFlag{
		Name:        "proxy",
		Value:       false,
		Usage:       "with proxy?",
		Destination: &runProxy,
	}},
	Action: func(c *cli.Context) error {
		_, err := os.Stat(getDir())
		if err != nil && os.IsNotExist(err) {
			fmt.Println("run init command")
			return err
		}
		d := docker.New()
		// d.SetLocalProxy()
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
