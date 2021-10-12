package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/contract/repository/sqldb"
	_ "github.com/albatiqy/gopoh/contract/repository/sqldb/mysql"
	_ "github.com/albatiqy/gopoh/contract/repository/sqldb/postgres"
	_ "github.com/albatiqy/gopoh/contract/repository/sqldb/sqlserver"
	"github.com/albatiqy/gopoh/internal/gopohgen/driver"
	_ "github.com/albatiqy/gopoh/internal/gopohgen/driver/mysql"
	_ "github.com/albatiqy/gopoh/internal/gopohgen/driver/postgres"
	_ "github.com/albatiqy/gopoh/internal/gopohgen/driver/sqlserver"
	"github.com/albatiqy/gopoh/pkg/lib/env"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
)

var (
	//go:embed _embed/main.txt
	txtMain string
	//go:embed _embed/dev-main.txt
	txtDevMain string
	//go:embed _embed/field-defs.txt
	txtFieldDefs string
	//go:embed _embed/sqlgen.txt
	txtSqlGen string
)

func main() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	if len(os.Args) < 2 {
		ui.Error(`
An nama_tabel must be specified.`)
		ui.Error("")
		os.Exit(1)
	}

	var (
		dbEnvKey, keyCol string
		sqlGenDev        bool
	)

	// pastikan flag terlebih dahulu
	flag.StringVar(&dbEnvKey, "d", "DEFAULT", "DB env key")
	flag.StringVar(&keyCol, "k", "", "table primary column")
	flag.BoolVar(&sqlGenDev, "q", false, "SQL Gen Development")
	flag.Usage = func() {
		ui.Output(`
Usage: gopohgen nama_tabel [options]
	Dispatches a custom event across the Serf cluster.
Options:
	-d=dbEnvKey             (default "DEFAULT")
	-k=primaryKeyColumn             (default [blank])
	-q             SQL Gen Development
	`)
	}
	flag.Parse()

	tableName := flag.Arg(0)

	workingDir := fs.WorkingDir()

	modName := util.GetModName(workingDir)
	if modName == "" {
		os.Exit(1)
	}

	// periksa FLAG!!!!!!!!!

	pathEnv := filepath.Join(workingDir, ".env")
	if fs.FileInfo(pathEnv) == nil {
		ui.Error("file \".env\" tidak ditemukan pada root project")
		os.Exit(1)
	}
	env.Load(pathEnv)

	if sqlGenDev {
		pathSqlDevGen := filepath.Join(workingDir, "_APPFS_/sql-dev")
		if success, err := fs.MkDirIfNotExists(pathSqlDevGen); !success {
			ui.Error(fmt.Sprintf("Error: %s", err))
			os.Exit(1)
		} else {
			if err == nil {
				fs.WriteTextFile("*", filepath.Join(pathSqlDevGen, ".gitignore"))
			}
		}

		pathSqlGenDir := filepath.Join(workingDir, "internal/sqlgen")
		if success, err := fs.MkDirIfNotExists(pathSqlGenDir); !success {
			ui.Error(fmt.Sprintf("Error: %s", err))
			os.Exit(1)
		}

		nsTableName := strings.Replace(tableName, ".", "_", 1)

		nsName := dbEnvKey + "_" + nsTableName

		pathSaveRoot := filepath.Join(pathSqlDevGen, nsName)
		if success, err := fs.MkDirIfNotExists(pathSaveRoot); !success {
			ui.Error(fmt.Sprintf("Error: %s", err))
			os.Exit(1)
		}

		fnameFieldDefs := filepath.Join(pathSaveRoot, "field-defs.go")
		fnameMain := filepath.Join(pathSaveRoot, "main.go")

		genMain := true

		if fs.FileInfo(fnameMain) != nil {
			genMain = false
		}

		fnameGen := filepath.Join(pathSqlGenDir, nsName+".go")
		genGen := true

		if fs.FileInfo(fnameGen) != nil {
			genGen = false
		}

		dbConn := sqldb.NewConn(getDBSetting(dbEnvKey))
		if dbConn == nil {
			ui.Error("tidak dapat terkoneksi dengan database")
			os.Exit(1)
		}
		dbDriver := dbConn.DriverName()

		genDriver := driver.Get(dbDriver)

		tableData, err := genDriver.ReadTable(tableName, keyCol, dbConn.DB)
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

		genStructName := strings.ReplaceAll(nsName, "_", " ")
		genStructName = strings.ToLower(genStructName)
		genStructName = strings.Title(genStructName)
		genStructName = strings.ReplaceAll(genStructName, " ", "")

		if genMain {
			var (
				useImport = map[string]string{}
				imports   []string
			)

			useImport["sqlgen"] = modName + "/internal/sqlgen"

			for impk, impv := range useImport {
				if impv != "" {
					imports = append(imports, "\t\""+impv+`"`)
				} else {
					if impl, ok := util.ImportsMap[impk]; ok {
						imports = append(imports, "\t\""+impl+`"`)
					}
				}
			}

			dbDriverEngine := modName + "/pkg/gendriver/" + dbDriver

			tplData := map[string]string{
				"imports":              strings.Join(imports, "\n"),
				"dbDriverEngine":       dbDriverEngine,
				"genStructName":        genStructName,
				"dbEnvKey":             dbEnvKey,
				"dbDriver":             dbDriver,
				"tableName":            tableName,
				"entityName":           nsTableName,
				"keyAttr":              tableData.KeyCol,
				"keyAutoStr":           keyAutoStr,
				"keyCanUpdateStr":      keyCanUpdateStr,
				"softDeleteStr":        softDeleteStr,
				"overridesStructField": strings.Join(tableData.StrOverridesStructField(), "\n"),
				"overridesType":        strings.Join(tableData.StrOverridesType(), "\n"),
				"overridesJSON":        strings.Join(tableData.StrOverridesJSON(), "\n"),
				"overridesLabel":       strings.Join(tableData.StrOverridesLabel(), "\n"),
			}

			util.WriteTplFile(fnameMain, txtDevMain, tplData)
		}

		if genGen {
			var (
				useImport = map[string]string{}
				imports   []string
			)

			useImport["sqldriver"] = modName + "/pkg/gendriver"

			for impk, impv := range useImport {
				if impv != "" {
					imports = append(imports, "\t\""+impv+`"`)
				} else {
					if impl, ok := util.ImportsMap[impk]; ok {
						imports = append(imports, "\t\""+impl+`"`)
					}
				}
			}

			tplData := map[string]string{
				"imports":       strings.Join(imports, "\n"),
				"genStructName": genStructName,
			}

			util.WriteTplFile(fnameGen, txtSqlGen, tplData)
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
	} else {
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

		nsTableName := strings.Replace(tableName, ".", "_", 1)

		nsName := dbEnvKey + "_" + nsTableName
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

		tableData, err := genDriver.ReadTable(tableName, keyCol, dbConn.DB)
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
				"entityName":           nsTableName,
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
}

func getDBSetting(envKey string) *sqldb.DBSetting {
	driverName := os.Getenv("DB_" + envKey + "_DRIVER")
	if driverName == "" {
		return nil
	}

	setting := &sqldb.DBSetting{
		DSN:        os.Getenv("DB_" + envKey + "_DSN"),
		DriverName: driverName,
	}

	return setting
}
