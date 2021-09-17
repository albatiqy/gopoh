package driver

import (
	"database/sql"
	"fmt"
	"strings"
)

var LoadedDrivers = make(map[string]interface{})

type Driver interface {
	ReadTable(tblName string, db *sql.DB) (*TableData, error)
}

type ColData struct {
	Name                string
	CompatibleGoTypeStr string
	Nullable            bool
	DBRequired          bool
	JSON                string
	Label               string
}

type TableData struct {
	strFields               []string
	strOverridesType        []string
	strOverridesJSON        []string
	strOverridesLabel       []string
	strOverridesStructField []string
	UseImport               map[string]string
	KeyCol                  string
	KeyAuto                 bool
	SoftDelete              bool
	colsData                []ColData
}

func (tableData TableData) StrFields() []string {
	return tableData.strFields
}

func (tableData TableData) StrOverridesType() []string {
	return tableData.strOverridesType
}

func (tableData TableData) StrOverridesJSON() []string {
	return tableData.strOverridesJSON
}

func (tableData TableData) StrOverridesLabel() []string {
	return tableData.strOverridesLabel
}

func (tableData TableData) StrOverridesStructField() []string {
	return tableData.strOverridesStructField
}

func (tableData TableData) ColsData() []ColData {
	return tableData.colsData
}

func NewTableData(colsData []ColData, keyCol string, keyAuto bool, softDelete bool, useImport map[string]string) *TableData {
	tableData := &TableData{}
	tableData.KeyCol = keyCol
	tableData.KeyAuto = keyAuto
	tableData.SoftDelete = softDelete
	tableData.UseImport = useImport
	lenCol := len(colsData)
	tableData.strFields = make([]string, lenCol)
	tableData.strOverridesType = make([]string, lenCol)
	tableData.strOverridesJSON = make([]string, lenCol)
	tableData.strOverridesLabel = make([]string, lenCol)
	tableData.strOverridesStructField = make([]string, lenCol)
	for i, colData := range colsData {
		dbRequired := ""
		if colData.DBRequired {
			dbRequired = ", DBRequired: true"
		}
		tableData.strFields[i] = fmt.Sprintf("\t\t"+`"%[1]s": {Col: "%[1]s", Type: (*%[2]s)(nil), JSON: "%[3]s", Label: "%[4]s", Ordinal: %[5]d%[6]s},`, colData.Name, colData.CompatibleGoTypeStr, colData.JSON, colData.Label, i, dbRequired)
		tableData.strOverridesType[i] = fmt.Sprintf("\t\t\t\t"+`// "%[1]s":  (*%[2]s)(nil),`, colData.Name, colData.CompatibleGoTypeStr)
		tableData.strOverridesJSON[i] = fmt.Sprintf("\t\t\t\t"+`// "%[1]s":  "%[1]s",`, colData.JSON)
		tableData.strOverridesLabel[i] = fmt.Sprintf("\t\t\t\t"+`// "%[1]s":  "%[2]s",`, colData.Name, colData.Label)
		tableData.strOverridesStructField[i] = fmt.Sprintf("\t\t\t\t"+`// "%[1]s":  "%[2]s",`, colData.Name, strings.ReplaceAll(colData.Label, " ", ""))
	}
	tableData.colsData = colsData
	return tableData
}

func Get(driverName string) Driver {
	driver, ok := LoadedDrivers[driverName]
	if ok {
		return driver.(Driver)
	}
	return nil
}
