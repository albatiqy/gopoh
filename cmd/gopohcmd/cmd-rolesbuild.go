package main

import (
	_ "embed"
	"flag"
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
)

var ()

type rolesbuildCmd struct {
	Ui cli.Ui
}

func (cmd *rolesbuildCmd) Help() string {
	helpText := `
Usage: serf event [options] name payload
  Dispatches a custom event across the Serf cluster.
Options:
  -coalesce=true/false      Whether this event can be coalesced. This means
                            that repeated events of the same name within a
                            short period of time are ignored, except the last
                            one received. Default is true.
  -rpc-addr=127.0.0.1:7373  RPC address of the Serf agent.
  -rpc-auth=""              RPC auth token of the Serf agent.
`
	return strings.TrimSpace(helpText)
}

func (cmd *rolesbuildCmd) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("rolesbuild", flag.ContinueOnError)
	cmdFlags.Usage = func() { cmd.Ui.Output(cmd.Help()) }

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if util.GetModName(workingDir) == "" {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	pathServiceDir := filepath.Join(workingDir, "internal/core/service")
	if fs.FileInfo(pathServiceDir) == nil {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	fset := token.NewFileSet() // positions are relative to fset

	d, err := parser.ParseDir(fset, pathServiceDir, nil, parser.ParseComments)
	if err != nil {
		cmd.Ui.Error(err.Error())
		return 1
	}

	regex, err := regexp.Compile(`\s*\[\s*@service\s*\]\s*`)
	if err != nil {
		cmd.Ui.Error(err.Error())
		return 1
	}

	for _, f := range d {
		myDoc := doc.New(f, "./", doc.AllMethods)
		for _, service := range myDoc.Types {
			for _, mthd := range service.Methods {
				if mthd.Doc == "" {
					continue
				}
				if match := regex.MatchString(mthd.Doc); match {
					fmt.Print(service.Name, ".", mthd.Name, "\n")
				}
			}
		}
	}

	return 0
}

func (cmd *rolesbuildCmd) Synopsis() string {
	return "Send a custom event through the Serf cluster"
}

func init() {
	cmds["rolesbuild"] = func() (cli.Command, error) {
		return &rolesbuildCmd{Ui: ui}, nil
	}
}
