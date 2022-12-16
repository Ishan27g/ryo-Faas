package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/docker"
	"github.com/go-git/go-git/v5"
	"github.com/mitchellh/go-homedir"
	cp "github.com/otiai10/copy"
	"github.com/urfave/cli/v2"
)

//var cwdMock = "/Users/ishan/go/src/github.com/Ishan27g/ryo-Faas"
var cwdMock = "/Users/ishan/Documents/Drive/golang"

const DirKey = "RYO_FAAS"

const LocalProxy = "LOCAL_PROXY"
const LocalDB = "LOCAL_DB"

const ReuseFnImg = "REUSE_FN_IMG"

var repositoryUrl = "https://github.com/Ishan27g/registry.git"

//var local = flag.Bool("create in cwd", true, "")
//var copy = flag.Bool("clone or copy", true, "")

var local = true
var copy = true
var dirName = "/ryo-Faas" // todo .
var gitDir = "/gitDir"
var directory = ""

var isProxyLocal = func() bool { return os.Getenv(LocalProxy) == "YES" }
var isDbLocal = func() bool { return os.Getenv(LocalDB) == "YES" }
var isFnImgBuilt = func() bool { return os.Getenv(ReuseFnImg) == "YES" }

func getDir() string {
	if directory != "" && os.Getenv(DirKey) != "" {
		return os.Getenv(DirKey)
	}
	// flag.Parse()
	if !local {
		dir, err := homedir.Dir()
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
		cwdMock, err = homedir.Expand(dir)
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
	}
	directory = cwdMock + dirName + gitDir
	os.Setenv(DirKey, directory)
	return directory
}

func cloneRepo() bool {
	if getDir() == "" {
		return false
	}
	var path = getDir()
	err := os.RemoveAll(path)
	if err != nil {
		log.Println("ok", err.Error())
	}
	if !copy {
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL: repositoryUrl,
			//Progress:   os.Stdout,
			Depth:      1,
			RemoteName: "main",
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		err = os.Setenv(DirKey, path)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	} else {
		err := cp.Copy("/Users/ishan/go/src/github.com/Ishan27g/ryo-Faas/", path)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		err = os.Setenv(DirKey, path)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	return true
}
func setupHomeDir() bool {
	if getDir() == "" {
		return false
	}
	return makeDir()
}

func makeDir() bool {
	if getDir() == "" {
		return false
	}
	err := os.MkdirAll(getDir(), os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Println("created - ", getDir())
	return true
}

var initRfaFaasCmd = cli.Command{
	Name:            "init",
	Aliases:         []string{"i"},
	Usage:           "initialise rfa-Faas",
	ArgsUsage:       "proxyCli init",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		now := time.Now()

		if !setupHomeDir() {
			return cli.Exit("cannot create directory", 1)
		}
		if !cloneRepo() {
			return cli.Exit("unable to clone repository", 1)
		}
		d := docker.New()
		d.SetSilent()

		go d.Setup()

		if d.CheckImages() {
			return cli.Exit("", 0)
		} else {
			fmt.Println("still setting up... ", time.Since(now).Seconds())
		}
		for {
			select {
			case <-time.After(10 * time.Second):
				if d.CheckImages() {
					return cli.Exit("Done", 0)
				} else {
					fmt.Println("still setting up... ", time.Since(now).Seconds())
				}
			}
		}
	},
}

var envCmd = cli.Command{
	Name:            "env",
	Aliases:         []string{"e"},
	Usage:           "print dir/ used by rfa-Faas",
	ArgsUsage:       "proxyCli env",
	HideHelp:        false,
	HideHelpCommand: false,
	Action: func(c *cli.Context) error {
		dir := getDir()
		fmt.Println(DirKey + "=" + dir)
		fmt.Println(LocalProxy+"=", isProxyLocal())
		fmt.Println(LocalDB+"=", isDbLocal())
		return nil
	},
}
