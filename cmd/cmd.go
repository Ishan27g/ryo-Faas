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
	"strings"

	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
	cp "github.com/otiai10/copy"
	"github.com/urfave/cli/v2"
)

var proxyAddress string // rpc address of proxy (default :9001)
var proxyHttpAddr = "localhost" + proxy.DefaultHttp

type definition struct {
	Deploy []struct {
		Name        string `json:"name"`
		FilePath    string `json:"filePath"`
		PackageDir  string `json:"packageDir"`
		Async       bool   `json:"Async"`
		MainProcess bool   `json:"mainProcess"`
	} `json:"deploy"`
}

var getProxy = func() transport.AgentWrapper {
	if proxyAddress == "" {
		proxyAddress = proxy.DefaultRpc
	}
	// return transport.ProxyGrpcClient(proxyAddress)
	return transport.ProxyGrpcClient(proxyAddress)
}

var read = func(defFile string) (definition, bool) {
	var d definition
	var fns definition

	content, err := ioutil.ReadFile(defFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = json.Unmarshal(content, &d)
	if err != nil {
		log.Fatal(err.Error())
	}
	var df []*deploy.Function
	var isMain = false
	for _, fn := range d.Deploy {
		df = append(df, &deploy.Function{
			Entrypoint: fn.Name,
			FilePath:   fn.FilePath,
			Dir:        fn.PackageDir,
			Async:      fn.Async,
		})
		if fn.MainProcess {
			isMain = true
		}
	}

	cwd := getDir() + "/"
	err = os.Chdir(cwd)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = os.MkdirAll(cwd+tmpDir, os.ModePerm)
	valid, genFile := GenerateFile(isMain, cwd+tmpDir, df)
	if !valid {
		log.Fatal("Invalid definition ")
	}
	fmt.Println("Generated file", genFile)
	for _, fn := range d.Deploy {
		dir, fName := filepath.Split(fn.FilePath)
		pn := filepath.Base(dir)
		if err := cp.Copy(fn.PackageDir, cwd+tmpDir+pn); err != nil {
			log.Fatal("Error copying files ", err.Error())
		}
		fn.PackageDir = cwd + tmpDir
		fn.FilePath = cwd + tmpDir + pn + "/" + fName
		fns.Deploy = append(fns.Deploy, fn)
	}
	return fns, isMain
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
		fmt.Println(err.Error())
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return body
}

var bypass bool
var deployCmd = cli.Command{

	Name:    "deploy",
	Aliases: []string{"d"},
	Flags: []cli.Flag{
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

		fmt.Println(bypass)

		df, isMain := read(c.Args().First())

		// process entire definition (all function per deploy) into a single container
		var fns []*deploy.Function
		for _, s := range df.Deploy {
			s.Name = strings.ToLower(s.Name)
			df := &deploy.Function{
				Entrypoint: s.Name,
				FilePath:   s.FilePath,
				Dir:        s.PackageDir,
				Async:      s.Async,
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
		fmt.Println("Starting container ...")

		d := docker.New()
		// d.SetLocalProxy()

		if d.RunFunction(fns[0].Entrypoint) != nil {
			log.Fatal("cannot run container" + fns[0].Entrypoint)
		}
		if !bypass {
			// add container proxy
			deployResponse, err := proxy.Deploy(c.Context, &deploy.DeployRequest{Functions: fns})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			printResonse(deployResponse)
		}
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
		rsp, err := getProxy().Details(context.Background(), &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: ""}})
		if err != nil {
			return cli.Exit("cannot get details", 1)
		}
		for _, f := range rsp.Functions {
			fmt.Println(f)
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
			printResonse(response)
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
		sendHttp("/reset", "")
		return nil
	},
}

func Init() *cli.App {
	app := &cli.App{Commands: []*cli.Command{&initRfaFaasCmd, &envCmd, &deployCmd, &stopCmd, &detailsProxyCmd,
		&proxyResetCmd, &startRyoFaas, &stopRyoFaas},
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