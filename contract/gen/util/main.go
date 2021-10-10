package util

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/albatiqy/gopoh/contract/log"
)

var (
	ImportsMap = map[string]string{
		"time":      "time",
		"strconv":   "strconv",
		"null":      "github.com/albatiqy/gopoh/pkg/lib/null",
		"decimal":   "github.com/albatiqy/gopoh/pkg/lib/decimal",
		"sqlserver": "github.com/albatiqy/gopoh/contract/repository/sqldb/sqlserver",
		"mysql":     "github.com/albatiqy/gopoh/contract/repository/sqldb/mysql",
	}
)

func GetModName(pathPrjDir string) string {
	fgomodName := filepath.Join(pathPrjDir, "go.mod")

	fgomod, err := os.Open(fgomodName)
	if err != nil {
		log.Errorf("file \"%s\" tidak dapat dibuka: %s", fgomodName, err)
		return ""
	}
	defer fgomod.Close()

	scanner := bufio.NewScanner(fgomod)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()
	fgomodfirstline := strings.Split(scanner.Text(), " ")

	return fgomodfirstline[1]
}

func WriteTplFile(fname, tplStr string, tplData map[string]string) {
	tmp1 := template.New("Template_1")
	tmp1, err := tmp1.Parse(tplStr)
	if err != nil {
		log.Fatal("WriteTplFile: ", err)
	}
	fOut, err := os.Create(fname)
	if err != nil {
		log.Fatal("WriteTplFile: ", err)
	}
	defer fOut.Close()

	if err = tmp1.Execute(fOut, tplData); err != nil {
		log.Fatal("WriteTplFile: ", err)
	}
}
