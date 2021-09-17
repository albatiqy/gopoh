package main

import (
	"log"
	"os"

	"github.com/mitchellh/cli"
)

var (
	app = cli.NewCLI("app", "1.0.0")
	cmds = map[string]cli.CommandFactory{}
	workingDir string
	ui = &cli.BasicUi{Writer: os.Stdout}
)

func main() {
	app.Args = os.Args[1:]
	app.Commands = cmds
	// app.Autocomplete = true

	var err error
	workingDir, err = os.Getwd()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	exitStatus, err := app.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}