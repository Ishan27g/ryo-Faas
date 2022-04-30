package cmd

import (
	"fmt"
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

const DIR_KEY = "RYA_FAAS"

var repositoryUrl = "https://github.com/Ishan27g/registry.git"

//var local = flag.Bool("create in cwd", true, "")
//var copy = flag.Bool("clone or copy", true, "")

var local = true
var copy = true
var dirName = "/ryo-Faas" // todo .
var gitDir = "/gitDir"
var directory = ""

func getDir() string {
	if directory != "" {
		return directory
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
	directory = cwdMock + dirName
	os.Setenv(DIR_KEY, directory)
	return directory
}

func cloneRepo() bool {
	if getDir() == "" {
		return false
	}
	var path = getDir() + gitDir
	os.RemoveAll(path + "/")
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
		err = os.Setenv(DIR_KEY, path)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	} else {
		err := cp.Copy("/Users/ishan/go/src/github.com/Ishan27g/ryo-Faas/", getDir())
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		err = os.Setenv(DIR_KEY, getDir())
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
	err := os.MkdirAll(getDir()+gitDir, os.ModePerm)
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
			return cli.Exit("Done", 0)
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
		fmt.Println(DIR_KEY + "=" + dir)
		return nil
	},
}
