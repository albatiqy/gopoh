package main

import (
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/pkg/lib/env"
	"github.com/albatiqy/gopoh/pkg/lib/fs"

	_ "github.com/lib/pq"
)

var ()

type buildrolesCmd struct {
	Ui cli.Ui
}

func (cmd *buildrolesCmd) Help() string {
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

func (cmd *buildrolesCmd) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("buildroles", flag.ContinueOnError)
	cmdFlags.Usage = func() { cmd.Ui.Output(cmd.Help()) }

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if util.GetModName(workingDir) == "" {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	env.Load()



	// init config



	cfg := configuration{
		AppFSRoot:  "./_APPFS_",
		Ui:         cmd.Ui,
		dbSettings: map[string]*dbSetting{},
	}
	cfg.readEnv(workingDir)

	pathServiceDir := filepath.Join(workingDir, "internal/core/service")
	if fs.FileInfo(pathServiceDir) == nil {
		cmd.Ui.Error("direktori project tidak valid")
		return 1
	}

	db := cfg.getPostgresDB("DEFAULT")
	if db == nil {
		cmd.Ui.Error("koneksi database error")
		return 1
	}
	defer db.Close()

	if _, err := db.Exec("TRUNCATE TABLE roles RESTART IDENTITY"); err != nil {
		cmd.Ui.Warn(err.Error())
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

	stmt, err := db.Prepare("INSERT INTO roles (role_id) VALUES ($1)")
	if err != nil {
		cmd.Ui.Error(err.Error())
		return 1
	}
	defer stmt.Close()

	for _, f := range d {
		myDoc := doc.New(f, "./", doc.AllMethods)
		for _, service := range myDoc.Types {
			for _, mthd := range service.Methods {
				if mthd.Doc == "" {
					continue
				}
				if match := regex.MatchString(mthd.Doc); match {
					fmt.Print(service.Name, ".", mthd.Name, "\n")
					if _, err := stmt.Exec(service.Name + "." + mthd.Name); err != nil {
						cmd.Ui.Warn(err.Error())
					}
				}
			}
		}
	}

	return 0
}

func (cmd *buildrolesCmd) Synopsis() string {
	return "Send a custom event through the Serf cluster"
}

func init() {
	cmds["buildroles"] = func() (cli.Command, error) {
		return &buildrolesCmd{Ui: ui}, nil
	}
}

type dbSetting struct {
	DriverName string
	DSN       string
}

type configuration struct {
	AppFSRoot      string
	HTTPListenBind string
	HTTPBasePath   string
	Ui             cli.Ui
	dbSettings     map[string]*dbSetting
}

func (cfg *configuration) readEnv(envDir string) {
	if val, ok := os.LookupEnv("APP_FS_ROOT"); ok {
		cfg.AppFSRoot = val
	}
	if strings.HasPrefix(cfg.AppFSRoot, ".") {
		cfg.AppFSRoot = filepath.Join(envDir, cfg.AppFSRoot)
	}
	fileInfo := fs.FileInfo(cfg.AppFSRoot)
	if fileInfo == nil {
		cfg.Ui.Warn("FS ROOT tidak ditemukan")
	}
	if !fileInfo.IsDir() {
		cfg.Ui.Warn("FS ROOT bukan direktori")
	}
	// readEnvKeyString("HTTP_LISTEN_BIND", &cfg.HTTPListenBind)
	// readEnvKeyString("HTTP_BASE_PATH", &cfg.HTTPBasePath)
}

func (cfg *configuration) getDBSetting(envKey string) *dbSetting {
	setting, ok := cfg.dbSettings[envKey]
	if !ok {
		driverName := os.Getenv("DB_" + envKey + "_DRIVER")
		if driverName == "" {
			return nil
		}

		setting = &dbSetting{
			DSN:       os.Getenv("DB_" + envKey + "_DSN"),
			DriverName: driverName,
		}

		cfg.dbSettings[envKey] = setting
	}
	return setting
}

/*
func readEnvKeyString(envKey string, cfgField *string) {
	if val, ok := os.LookupEnv(envKey); ok {
		*cfgField = val
	}
}
*/

func (cfg *configuration) getPostgresDB(envKey string) *sql.DB {
	dbSetting := cfg.getDBSetting(envKey)
	if dbSetting.DriverName != "postgres" {
		cfg.Ui.Warn(`sqldb: driver tidak cocok`)
		return nil
	}
	// Use DSN string to open
	db, err := sql.Open("postgres", dbSetting.DSN)
	if err != nil {
		cfg.Ui.Warn(`sqldb: ` + err.Error())
		return nil
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
