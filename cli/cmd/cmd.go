package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"

	"github.com/Ishan27g/ryo-Faas/agent/registry"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/urfave/cli/v2"
)

var proxyAddress string // rpc address of proxy (default :9001)
var proxyHttpAddr = "localhost" + proxy.DefaultHttp

type definition struct {
	Deploy []struct {
		Name       string `json:"name"`
		FilePath   string `json:"filePath"`
		PackageDir string `json:"packageDir"`
		Async      bool   `json:"Async"`
	} `json:"deploy"`
}

var getProxy = func() transport.AgentWrapper {
	if proxyAddress == "" {
		proxyAddress = proxy.DefaultRpc
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
	var df []*deploy.Function
	for _, fn := range d.Deploy {
		df = append(df, &deploy.Function{
			Entrypoint: fn.Name,
			FilePath:   fn.FilePath,
			Dir:        fn.PackageDir,
			Async:      fn.Async,
		})
	}
	cwd, _ := os.Getwd()
	registry.ModFile = func() string {
		return "/Users/ishan/go/src/github.com/Ishan27g/ryo-Faas/template/template.go"
	}
	registry.ImportPath = "github.com/Ishan27g/ryo-Faas/deployments/"
	valid, genFile := registry.AstLocalCopy(cwd+"/../deployments/", df)
	if !valid {
		log.Fatal("Invalid definition ")
	}
	fmt.Println("Generated file", genFile)
	for _, fn := range d.Deploy {
		dir, _ := filepath.Split(fn.FilePath)
		pn := filepath.Base(dir)
		if err := cp.Copy(fn.PackageDir, cwd+"/../deployments/"+pn); err != nil {
			log.Fatal("Error copying files ")
		}
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
func printResonse(response *deploy.DeployResponse) {
	//printJson(response)
	for _, fn := range response.Functions {
		fmt.Printf("%s %s [%s]\n", fn.Entrypoint, fn.Url, fn.Status)
	}
}

// localhost:9000
func sendHttp(url, agentAddr string) []byte {
	resp, err := http.Get("http://" + proxyHttpAddr + url + agentAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
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
		_ = read(c.Args().First())
		// proxy := getProxy()
		// if proxy == nil {
		// 	return cli.Exit("cannot connect to "+proxyAddress, 1)
		// }

		// var fns []*deploy.Function

		// for _, s := range df.Deploy {
		// 	df := &deploy.Function{
		// 		Entrypoint: s.Name,
		// 		FilePath:   s.FilePath,
		// 		Dir:        s.PackageDir,
		// 		Async:      s.Async,
		// 	}
		// 	fns = append(fns, df)
		// 	if !transport.UploadDir(proxy, context.Background(), df.Dir, df.Entrypoint) {
		// 		return cli.Exit("cannot upload directory to proxy "+df.Dir, 1)
		// 	}
		// }
		// deployResponse, err := proxy.Deploy(c.Context, &deploy.DeployRequest{Functions: fns})
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	return err
		// }
		// printResonse(deployResponse)
		return nil
	},
}
var detailsProxyCmd = cli.Command{
	Name:            "details",
	Usage:           "list current details",
	ArgsUsage:       "server-cli details",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		var rsp []types.FunctionJsonRsp
		json.Unmarshal(sendHttp("/details", ""), &rsp)
		for _, v := range rsp {
			fmt.Printf("%s\t\t%s\t%s\n", v.Url, v.Name, v.Status)
		}
		return nil
	},
}
var functionStatusProxyCmd = cli.Command{
	Name:            "details",
	Usage:           "list current details",
	ArgsUsage:       "server-cli details",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		var rsp []types.FunctionJsonRsp
		json.Unmarshal(sendHttp("/details", ""), &rsp)
		for _, v := range rsp {
			fmt.Printf("%s\t\t%s\t%s\n", v.Url, v.Name, v.Status)
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
		response, err := proxy.List(context.Background(), &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: c.Args().First()}})
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		printResonse(response)
		return nil
	},
}
var stopCmd = cli.Command{
	Name:            "stop",
	Aliases:         []string{"s"},
	Usage:           "stop a function",
	ArgsUsage:       "server-cli stop {entrypoint}",
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
			printResonse(response)
		}
		return nil
	},
}
var logsCmd = cli.Command{
	Name:            "log",
	Aliases:         []string{"l"},
	Usage:           "log a function",
	ArgsUsage:       "server-cli log {entrypoint}",
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

		response, err := proxy.Logs(c.Context, &deploy.Function{Entrypoint: c.Args().First()})
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		printJson(response)
		// printResonse(response)
		return nil
	},
}
var agentAddCmd = cli.Command{
	Name:            "add",
	Aliases:         []string{"a"},
	Usage:           "add an agent",
	ArgsUsage:       "server-cli add {address}",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return cli.Exit("address not provided", 1)
		}
		fmt.Println(c.Args().First())
		sendHttp("/addAgent?address=", c.Args().First())
		return nil
	},
}
var proxyResetCmd = cli.Command{
	Name:            "reset",
	Usage:           "reset the proxy",
	ArgsUsage:       "server-cli reset",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		sendHttp("/reset", "")
		return nil
	},
}

func Init() *cli.App {
	app := &cli.App{Commands: []*cli.Command{&deployCmd, &listCmd, &stopCmd, &agentAddCmd, &detailsProxyCmd,
		&logsCmd, &proxyResetCmd, &startRyoFaas, &stopRyoFaas},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "proxy",
				Aliases:     []string{"p"},
				DefaultText: "RPC port of the proxy server, default ",
				Destination: &proxyAddress,
			},
		}}
	return app
}
