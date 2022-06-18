package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/urfave/cli/v2"
)

var deployCmd = cli.Command{

	Name:    "deploy",
	Aliases: []string{"d"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "async",
			Value:       false,
			Usage:       "deploy async function",
			Destination: &isAsync,
		},
		&cli.BoolFlag{
			Name:        "main",
			Value:       false,
			Usage:       "deploy Init method",
			Destination: &isMain,
		},
		&cli.BoolFlag{
			Name:        "bypass",
			Value:       false,
			Usage:       "bypass deployment to proxy",
			Destination: &bypass,
		},
	},
	Usage:           "deploy a definition",
	ArgsUsage:       "proxyCli deploy {path to definition.json}",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return cli.Exit("filename not provided", 1)
		}

		df, isMain := read(c.Args().First())

		// process entire definition (all function per deploy) into a single container
		var fns []*deploy.Function
		for _, s := range df.Deploy {
			s.Name = strings.ToLower(s.Name)
			df := &deploy.Function{
				Entrypoint: s.Name,
				FilePath:   s.FilePath,
				Dir:        s.PackageDir,
				Async:      isAsync,
				IsMain:     isMain,
			}
			fns = append(fns, df)
		}

		var proxy transport.AgentWrapper
		if !bypass {
			proxy = getProxy()
			if proxy == nil {
				return cli.Exit("cannot connect to "+proxyAddress, 1)
			}
		}

		// run definition as single container
		d := docker.New()

		if d.BuildFunction(fns[0].Entrypoint) != nil {
			log.Fatal("cannot run container" + fns[0].Entrypoint)
		}
		if !bypass {
			// add container proxy
			deployResponse, err := proxy.Deploy(c.Context, &deploy.DeployRequest{Functions: fns})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			printResponse(deployResponse)
		}
		// delete dir to `go build` latest during next run
		os.RemoveAll("deployment/tmp/")
		return nil
	},
}
var detailsProxyCmd = cli.Command{
	Name:            "details",
	Usage:           "get current details",
	ArgsUsage:       "server-cmd details",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		proxy := getProxy()
		if proxy == nil {
			return cli.Exit("cannot connect to "+proxyAddress, 1)
		}
		rsp, err := proxy.Details(context.Background(), &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: ""}})
		if err != nil {
			return cli.Exit("cannot get details", 1)
		}
		for _, f := range rsp.Functions {
			fmt.Printf("%s %20s ", f.Url, f.Entrypoint)
			if f.IsMain {
				fmt.Printf("[Main]")
			}
			if f.Async {
				fmt.Printf("[Async]")
			}
			fmt.Printf("\n")
		}
		return nil
	},
}
var stopCmd = cli.Command{
	Name:            "stop",
	Aliases:         []string{"s"},
	Usage:           "stop a function",
	ArgsUsage:       "server-cmd stop {entrypoint}",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return cli.Exit("entrypoint not provided", 1)
		}
		fmt.Println(c.Args().First())

		proxy := getProxy()
		if proxy == nil {
			return cli.Exit("cannot connect to "+proxyAddress, 1)
		}
		for _, s := range c.Args().Slice() {
			response, err := proxy.Stop(c.Context, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: s}})
			if err != nil {
				fmt.Println(err.Error())
			}
			printResponse(response)
		}
		return nil
	},
}
var proxyResetCmd = cli.Command{
	Name:            "reset",
	Aliases:         []string{"r"},
	Usage:           "reset the proxy",
	ArgsUsage:       "server-cmd reset",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		sendHttp("/reset")
		return nil
	},
}

func Init() *cli.App {
	app := &cli.App{Commands: []*cli.Command{
		&initRfaFaasCmd, &envCmd, &proxyResetCmd,
		&startRyoFaas, &stopRyoFaas, &pruneRyoFaas,
		&deployCmd, &stopCmd, &detailsProxyCmd},
		HideHelp:             true,
		HideHelpCommand:      true,
		HideVersion:          true,
		EnableBashCompletion: true,
	}
	return app
}
