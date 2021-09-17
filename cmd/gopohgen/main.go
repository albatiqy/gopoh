package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/contract/repository/sqldb"
	_ "github.com/albatiqy/gopoh/contract/repository/sqldb/postgres"
	"github.com/albatiqy/gopoh/internal/gopohgen/driver"
	_ "github.com/albatiqy/gopoh/internal/gopohgen/driver/postgres"
	"github.com/albatiqy/gopoh/pkg/lib/env"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
)

var (
	//go:embed _embed/main.txt
	txtMain string
	//go:embed _embed/field-defs.txt
	txtFieldDefs string
)

func main() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	if len(os.Args) < 2 {
		ui.Error(`
An nama_tabel must be specified.`)
		ui.Error("")
		os.Exit(1)
	}

	var dbEnvKey string

	tableName := os.Args[1]

	flag.StringVar(&dbEnvKey, "d", "DEFAULT", "DB env key")
	flag.Usage = func() {
		ui.Output(`
Usage: gopohgen nama_tabel [options]
	Dispatches a custom event across the Serf cluster.
Options:
	-d=dbEnvKey             (default "DEFAULT")
	`)
	}
	flag.Parse()

	workingDir := fs.WorkingDir()

	if util.GetModName(workingDir) == "" {
		ui.Error("direktori project tidak valid")
		os.Exit(1)
	}

	pathEnv := filepath.Join(workingDir, ".env")
	if fs.FileInfo(pathEnv) == nil {
		ui.Error("file \".env\" tidak ditemukan pada root project")
		os.Exit(1)
	}
	env.Load(pathEnv)

	pathGopohGen := filepath.Join(workingDir, "_APPFS_/gopoh-gen")
	if success, err := fs.MkDirIfNotExists(pathGopohGen); !success {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	} else {
		if err == nil {
			fs.WriteTextFile("*", filepath.Join(pathGopohGen, ".gitignore"))
		}
	}

	pathTableDef := filepath.Join(pathGopohGen, "table-def")
	if success, err := fs.MkDirIfNotExists(pathTableDef); !success {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	pathBakDir := filepath.Join(pathTableDef, "_bak")
	if success, err := fs.MkDirIfNotExists(pathBakDir); !success {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	nsName := dbEnvKey + "_" + tableName
	pathSaveRoot := filepath.Join(pathTableDef, nsName)
	if success, err := fs.MkDirIfNotExists(pathSaveRoot); !success {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	fnameFieldDefs := filepath.Join(pathSaveRoot, "field-defs.go")
	fnameMain := filepath.Join(pathSaveRoot, "main.go")

	genMain := true

	if success, err := fs.BackupIfExist(fnameFieldDefs, filepath.Join(pathBakDir, nsName+"-field-defs.txt")); !success {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	if fs.FileInfo(fnameMain) != nil {
		genMain = false
	}

	dbConn := sqldb.NewConn(getDBSetting(dbEnvKey))
	if dbConn == nil {
		ui.Error("tidak dapat terkoneksi dengan database")
		os.Exit(1)
	}
	dbDriver := dbConn.DriverName()

	genDriver := driver.Get(dbDriver)

	tableData, err := genDriver.ReadTable(tableName, dbConn.DB)
	if err != nil {
		ui.Error(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	keyCanUpdateStr := "false"
	keyAutoStr := "false"
	if tableData.KeyAuto {
		keyAutoStr = "true"
		keyCanUpdateStr = "false"
	}
	softDeleteStr := "false"
	if tableData.SoftDelete {
		softDeleteStr = "true"
	}

	if genMain {
		tplData := map[string]string{
			"dbEnvKey":             dbEnvKey,
			"dbDriver":             dbDriver,
			"tableName":            tableName,
			"keyAttr":              tableData.KeyCol,
			"keyAutoStr":           keyAutoStr,
			"keyCanUpdateStr":      keyCanUpdateStr,
			"softDeleteStr":        softDeleteStr,
			"overridesStructField": strings.Join(tableData.StrOverridesStructField(), "\n"),
			"overridesType":        strings.Join(tableData.StrOverridesType(), "\n"),
			"overridesJSON":        strings.Join(tableData.StrOverridesJSON(), "\n"),
			"overridesLabel":       strings.Join(tableData.StrOverridesLabel(), "\n"),
		}

		util.WriteTplFile(fnameMain, txtMain, tplData)
	}

	var imports []string

	for impk, impv := range tableData.UseImport {
		if impv != "" {
			imports = append(imports, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				imports = append(imports, "\t\""+impl+`"`)
			}
		}
	}

	tplData := map[string]string{
		"fieldDefs": strings.Join(tableData.StrFields(), "\n"),
		"imports":   strings.Join(imports, "\n"),
	}

	util.WriteTplFile(fnameFieldDefs, txtFieldDefs, tplData)
}

func getDBSetting(envKey string) *sqldb.DBSetting {
	driverName := os.Getenv("DB_" + envKey + "_DRIVER")
	if driverName == "" {
		return nil
	}

	setting := &sqldb.DBSetting{
		Host:       os.Getenv("DB_" + envKey + "_HOST"),
		User:       os.Getenv("DB_" + envKey + "_USER"),
		Passwd:     os.Getenv("DB_" + envKey + "_PASSWORD"),
		Database:   os.Getenv("DB_" + envKey + "_DATABASE"),
		DriverName: driverName,
	}

	str := os.Getenv("DB_" + envKey + "_PORT")
	if str != "" {
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			setting.Port = uint(val)
		}
	}

	return setting
}
