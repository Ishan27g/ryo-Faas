package main

import (
	"log"
	"os"

	"github.com/Ishan27g/ryo-Faas/cmd"
)

func main() {
	app := cmd.Init()
	err := app.Run(os.Args)
	app.EnableBashCompletion = true
	if err != nil {
		log.Fatal(err)
	}
}
