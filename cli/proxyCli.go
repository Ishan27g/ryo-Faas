package main

import (
	"log"
	"os"

	"github.com/Ishan27g/ryo-Faas/cli/cmd"
)

func main() {
	app := cmd.Init()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
