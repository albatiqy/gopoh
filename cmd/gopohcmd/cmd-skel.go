package main

import (
	_ "embed"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
)

var (
	//go:embed _embed/gitignore-root.txt
	txtGitignoreRoot string
	//go:embed _embed/gitignore-appfs.txt
	txtGitignoreAppfs string
	/*
		//go:embed _embed/env.txt
		txtEnv string
	*/
	//go:embed _embed/env-sample.txt
	txtEnvSample string
)

type skelCmd struct {
	Ui cli.Ui
}

func (cmd *skelCmd) Help() string {
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

func (cmd *skelCmd) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("skel", flag.ContinueOnError)
	cmdFlags.Usage = func() { cmd.Ui.Output(cmd.Help()) }
	//cmdFlags.BoolVar(&coalesce, "coalesce", true, "coalesce")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// args = cmdFlags.Args()
	/*
		if len(args) < 1 {
			c.Ui.Error("An event name must be specified.")
			c.Ui.Error("")
			c.Ui.Error(c.Help())
			return 1
		} else if len(args) > 2 {
			c.Ui.Error("Too many command line arguments. Only a name and payload must be specified.")
			c.Ui.Error("")
			c.Ui.Error(c.Help())
			return 1
		}
	*/

	// c.Ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))

	// c.Ui.Output(fmt.Sprintf("Event '%s' dispatched! Coalescing enabled: %#v", event, coalesce))

	const (
		gitignoreParentOverrideIncludeAll  = "!*"
		gitignoreParentOverrideIncludeNone = "*"
	)

	mkdir := func(pth string) error {
		if success, err := fs.MkDirIfNotExists(pth); !success {
			return err
		}
		return nil
	}
	writeFile := func(pth, strContent string) error {
		if err := fs.WriteTextFile(strContent, pth); err != nil {
			return err
		}
		return nil
	}

	if util.GetModName(workingDir) == "" {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	pathBakDir := filepath.Join(workingDir, "_APPFS_/_bak")
	if err := mkdir(pathBakDir); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	if err := mkdir(filepath.Join(workingDir, "_APPFS_/_archive")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "_APPFS_/_refs")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "_APPFS_/log")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "_APPFS_/tmp")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "_APPFS_/gopoh-gen")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "cmd/appname")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := mkdir(filepath.Join(workingDir, "internal")); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := writeFile(filepath.Join(workingDir, ".gitignore"), txtGitignoreRoot); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := writeFile(filepath.Join(workingDir, "_APPFS_/.gitignore"), txtGitignoreAppfs); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := writeFile(filepath.Join(workingDir, "_APPFS_/_archive/.gitignore"), gitignoreParentOverrideIncludeAll); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}
	if err := writeFile(filepath.Join(workingDir, "_APPFS_/log/.gitignore"), gitignoreParentOverrideIncludeNone); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	/*
		pathEnv := filepath.Join(pathPrjDir, ".env")
		if success, err := fs.MoveIfExist(pathEnv, filepath.Join(pathBakDir, "env-%s.txt")); !success {
			log.Fatal(err)
		}
		writeFile(pathEnv, txtEnv)
	*/
	if err := writeFile(filepath.Join(workingDir, ".env-sample"), txtEnvSample); err != nil {
		cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
		return 1
	}

	return 0

	/*
		var format string
		cmdFlags := flag.NewFlagSet("info", flag.ContinueOnError)
		cmdFlags.Usage = func() { i.Ui.Output(i.Help()) }
		cmdFlags.StringVar(&format, "format", "text", "output format")
		rpcAddr := RPCAddrFlag(cmdFlags)
		rpcAuth := RPCAuthFlag(cmdFlags)
		if err := cmdFlags.Parse(args); err != nil {
			return 1
		}

		client, err := RPCClient(*rpcAddr, *rpcAuth)
		if err != nil {
			i.Ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))
			return 1
		}
		defer client.Close()

		stats, err := client.Stats()
		if err != nil {
			i.Ui.Error(fmt.Sprintf("Error querying agent: %s", err))
			return 1
		}

		output, err := formatOutput(StatsContainer(stats), format)
		if err != nil {
			i.Ui.Error(fmt.Sprintf("Encoding error: %s", err))
			return 1
		}

		i.Ui.Output(string(output))
		return 0
	*/

	// return cmd.RunResult
}

func (cmd *skelCmd) Synopsis() string {
	return "Send a custom event through the Serf cluster"
}

func init() {
	cmds["skel"] = func() (cli.Command, error) {
		return &skelCmd{Ui: ui}, nil
	}
}
