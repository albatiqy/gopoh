package postgres

import (
	"strings"
	"strconv"

	// "github.com/albatiqy/gopoh/pkg/lib/decimal"
	// "github.com/albatiqy/gopoh/pkg/lib/null"

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
	byIDSelect := baseSelect + " WHERE " + colsMap[uint16(keyIdx)] + "=$1" + andSoftDelete

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
	ph := make([]string, len(storeCols))
	for i := range storeCols {
		ph[i] = "$" + strconv.Itoa(i+1)
	}
	return "INSERT INTO " + tblName + " (" + strings.Join(storeCols, ",") + ") VALUES (" + strings.Join(ph, ",") + ")"
}

func (d Driver) GenerateEUpdate(tblName string, updateCols []string, keyCol string) string {
	strUpdate := make([]string, len(updateCols))
	z := int(1)
	for i, col := range updateCols {
		strUpdate[i] = col + "=$" + strconv.Itoa(z)
		z++
	}
	return "UPDATE " + tblName + " SET " + strings.Join(strUpdate, ",") + " WHERE " + keyCol + "=$" + strconv.Itoa(z)
}

func (d Driver) GenerateEDelete(tblName string, keyCol string, softDelete bool) string {
	// jika sudah ada tidak harus didelete
	if softDelete {
		return "UPDATE " + tblName + " SET " + softDeleteCol + "=$1 WHERE " + keyCol + "=$2"
	} else {
		return "DELETE FROM " + tblName + " WHERE " + keyCol + "=$1"
	}
}

/*
func (d Driver) GenerateInsertSQLPlaceholders(val interface{}) string {
	switch val.(type) {
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		return `%d`
	case *float32, *float64:
		return `%f`
	case *string:
		return `'%s'`
	default:
		return `%s` // need quote
	}
}
*/

func init() {
	driver.LoadedDrivers["postgres"] = Driver{}
	log.Debugf("postgres gen: %s", "initialized")
}
