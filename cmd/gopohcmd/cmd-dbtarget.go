package main

import (
	"bufio"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
)

type dbtargetCmd struct {
	Ui cli.Ui
}

func (cmd *dbtargetCmd) Help() string {
	helpText := `
Usage: gopohcmd dbtarget nama_tabel [options]
	Dispatches a custom event across the Serf cluster.
Options:
	-d=dbEnvKey             (default "DEFAULT")
`
	return strings.TrimSpace(helpText)
}

func (cmd *dbtargetCmd) Run(args []string) int {

	// args = cmdFlags.Args()
	if len(args) < 1 {
		cmd.Ui.Error("An event table_name must be specified.")
		cmd.Ui.Error("")
		cmd.Ui.Error(cmd.Help())
		return 1
	}

	tableName := args[0]

	cmdFlags := flag.NewFlagSet("dbtarget", flag.ContinueOnError)
	cmdFlags.Usage = func() { cmd.Ui.Output(cmd.Help()) }

	var dbEnvKey string
	cmdFlags.StringVar(&dbEnvKey, "e", "DEFAULT", "dbEnvKey")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if util.GetModName(workingDir) == "" {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	pathTableDefDir := filepath.Join(workingDir, "_APPFS_/gopoh-gen/table-def", dbEnvKey+"_"+tableName)
	fnameMain := filepath.Join(pathTableDefDir, "main.go")
	if fs.FileInfo(fnameMain) != nil {
		ps := exec.Command("go", "run", pathTableDefDir)
		stdout, err := ps.StdoutPipe()
		if err != nil {
			cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
			return 1
		}

		if err := ps.Start(); err != nil {
			cmd.Ui.Error(fmt.Sprintf("Error: %s", err))
			return 1
		}
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			cmd.Ui.Output(m)
		}
		ps.Wait()
	}

	return 0
}

func (cmd *dbtargetCmd) Synopsis() string {
	return "Send a custom event through the Serf cluster"
}

func init() {
	cmds["dbtarget"] = func() (cli.Command, error) {
		return &dbtargetCmd{Ui: ui}, nil
	}
}
