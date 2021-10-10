package mysql

import (
	"strings"

	"github.com/albatiqy/gopoh/contract/gen/driver"
	"github.com/albatiqy/gopoh/contract/log"
)

const (
	softDeleteCol = "deleted_at" // check compat
	timeStampCol  = "created_at"
)

type Driver struct {
}

func (d Driver) GenerateQESelects(tableSelectDefs []driver.TableSelectDef, softDelete bool) (string, string, string, map[uint16]string) {
	aliasesString := "abcdefghijklmnopqrstuvwxyz"
	aliasesRunes := []rune(aliasesString)
	selectFrom := ""

	var colSelectDefs []driver.ColSelectDef
	aliasMap := map[uint16]string{}
	colsMap := map[uint16]string{}
	keyIdx := -1
	leftTableAlias := ""

	for i, tableSelectDef := range tableSelectDefs {
		tbAlias := string(aliasesRunes[i])
		if i == 0 {
			selectFrom = tableSelectDef.TableName + " " + tbAlias
			leftTableAlias = tbAlias
		} else {
			selectFrom += " LEFT JOIN " + tableSelectDef.TableName + " " + tbAlias + " ON " + string(aliasesRunes[i-1]) + "." + tableSelectDef.KeyCol + "=" + tbAlias + "." + tableSelectDef.KeyCol
		}
		for _, colSelectDef := range tableSelectDef.ColSelectDefs {
			colSelectDefs = append(colSelectDefs, colSelectDef)
			aliasMap[colSelectDef.Ordinal] = tbAlias
			if i == 0 {
				if tableSelectDef.KeyCol == colSelectDef.Col {
					keyIdx = int(colSelectDef.Ordinal)
				}
			}
		}
	}

	selects := make([]string, len(colSelectDefs))
	for _, colSelectDef := range colSelectDefs {
		aliasSuffix := ""
		if colSelectDef.ColAlias != colSelectDef.Col {
			aliasSuffix = " " + colSelectDef.ColAlias //=============
		}
		colMap := aliasMap[colSelectDef.Ordinal] + "." + colSelectDef.Col
		selects[colSelectDef.Ordinal] = colMap + aliasSuffix
		colsMap[colSelectDef.Ordinal] = colMap
	}

	baseSelect := "SELECT " + strings.Join(selects, ",") + " FROM " + selectFrom
	andSoftDelete := ""
	whereSoftDelete := ""
	if softDelete {
		whereSoftDelete = leftTableAlias + "." + softDeleteCol + " IS NULL"
		andSoftDelete = " AND " + whereSoftDelete
	}
	byIDSelect := baseSelect + " WHERE " + colsMap[uint16(keyIdx)] + "=?" + andSoftDelete

	return baseSelect, whereSoftDelete, byIDSelect, colsMap
}

func (d Driver) GenerateQ(tblName string, selectCols []string, softDelete bool) (string, string) {
	whereSoftDelete := ""
	if softDelete {
		whereSoftDelete = softDeleteCol + " IS NULL"
	}
	return "SELECT " + strings.Join(selectCols, ",") + " FROM " + tblName, whereSoftDelete
}

func (d Driver) GenerateEStore(tblName string, storeCols []string) string {
	return "INSERT INTO " + tblName + " (" + strings.Join(storeCols, ",") + ") VALUES (" + strings.TrimSuffix(strings.Repeat("?,", len(storeCols)), ",") + ")"
}

func (d Driver) GenerateEUpdate(tblName string, updateCols []string, keyCol string) string {
	strUpdate := make([]string, len(updateCols))
	for i, col := range updateCols {
		strUpdate[i] = col + "=?"
	}
	return "UPDATE " + tblName + " SET " + strings.Join(strUpdate, ",") + " WHERE " + keyCol + "=?"
}

func (d Driver) GenerateEDelete(tblName string, keyCol string, softDelete bool) string {
	// jika sudah ada tidak harus didelete
	if softDelete {
		return "UPDATE " + tblName + " SET " + softDeleteCol + "=? WHERE " + keyCol + "=?"
	} else {
		return "DELETE FROM " + tblName + " WHERE " + keyCol + "=?"
	}
}

func init() {
	driver.LoadedDrivers["sqlserver"] = Driver{}
	log.Debugf("sqlserver gen: %s", "initialized")
}
