package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
	cp "github.com/otiai10/copy"
)

var proxyAddress string // rpc address of proxy (default :9001)
var proxyHttpAddr = "localhost" + proxy.DefaultHttp

var bypass bool
var isAsync = false
var isMain = false

type definition struct {
	Deploy []struct {
		Name       string `json:"name"`
		FilePath   string `json:"filePath"`
		PackageDir string `json:"packageDir"`
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
	for _, fn := range d.Deploy {
		dir, _ := filepath.Split(fn.FilePath)
		df = append(df, &deploy.Function{
			Entrypoint: fn.Name,
			FilePath:   fn.FilePath,
			Dir:        dir,
		})
	}

	cwd := getDir() + "/"
	err = os.Chdir(cwd)
	if err != nil {
		fmt.Println(err.Error())
	}

	os.MkdirAll(cwd+tmpDir, os.ModePerm)

	valid, _ := generateFile(cwd+tmpDir, df)
	if !valid {
		log.Fatal("Invalid definition ")
	}

	for _, fn := range d.Deploy {
		dir, fName := filepath.Split(fn.FilePath)
		pn := filepath.Base(dir)
		if err := cp.Copy(dir, cwd+tmpDir+pn); err != nil {
			log.Fatal("Error copying files ", err.Error())
		}
		fn.PackageDir = cwd + tmpDir
		fn.FilePath = cwd + tmpDir + pn + "/" + fName
		fns.Deploy = append(fns.Deploy, fn)
	}

	return fns, isMain
}

func printResonse(response *deploy.DeployResponse) {
	for _, fn := range response.Functions {
		fmt.Printf("%s %s [%s]\n", fn.Entrypoint, fn.Url, fn.Status)
	}
}

func sendHttp(url string) []byte {
	resp, err := http.Get("http://" + proxyHttpAddr + url)
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
