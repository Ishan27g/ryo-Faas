package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/urfave/cli/v2"
)

var proxyAddress string // rpc address of proxy (default :9001)

type definition struct {
	Deploy []struct {
		Name       string `json:"name"`
		FilePath   string `json:"filePath"`
		PackageDir string `json:"packageDir"`
	} `json:"deploy"`
}

var getProxy = func() transport.AgentWrapper {
	if proxyAddress == "" {
		proxyAddress = ":9001"
	}
	return transport.ProxyGrpcClient(proxyAddress)
}

var read = func(defFile string) definition {
	var d definition
	content, err := ioutil.ReadFile(defFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = json.Unmarshal(content, &d)
	if err != nil {
		log.Fatal(err.Error())
	}
	return d
}

func printJson(js interface{}) {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(data))
}

var deployCmd = cli.Command{
	Name:            "run",
	Aliases:         []string{"r"},
	Usage:           "run a definition",
	ArgsUsage:       "proxy-cli run {path to definition.json}",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return cli.Exit("filename not provided", 1)
		}
		df := read(c.Args().First())
		proxy := getProxy()
		if proxy == nil {
			return cli.Exit("cannot connect to "+proxyAddress, 1)
		}
		for _, s := range df.Deploy {
			deployResponse, err := proxy.Deploy(c.Context, &deploy.DeployRequest{Functions: &deploy.Function{
				Entrypoint: s.Name,
				FilePath:   s.FilePath,
				Dir:        s.PackageDir,
			}})
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			printJson(deployResponse)
		}
		return nil
	},
}
var listCmd = cli.Command{
	Name:            "list",
	Aliases:         []string{"l"},
	Usage:           "list function details",
	ArgsUsage:       "server-cli list {entrypoint}",
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
		response, err := proxy.List(c.Context, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: c.Args().First()}})
		if err != nil {
			return err
		}
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		printJson(response)
		return nil
	},
}

func main() {
	app := &cli.App{Commands: []*cli.Command{&deployCmd, &listCmd}, Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "proxy",
			Aliases:     []string{"p"},
			DefaultText: "RPC port of the proxy server, default ",
			Destination: &proxyAddress,
		},
	}}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
