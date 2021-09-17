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

/*
func (d Driver) GenerateQByID(tblName string, selectCols []string, keyCol string, softDelete bool) string {
	andSoftDelete := ""
	if softDelete {
		andSoftDelete = " AND " + softDeleteCol + " IS NULL"
	}
	return "SELECT " + strings.Join(selectCols, ",") + " FROM " + tblName + " WHERE " + keyCol + "=?" + andSoftDelete
}
*/

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
	driver.LoadedDrivers["mysql"] = Driver{}
	log.Debugf("mysql gen: %s", "initialized")
}

/*
func CreateTableFromMssql(tblName string, fields map[string]mssql.Field, customFields map[string]string) (int64, error) {
	var strFields []string
	for _, field := range fields {
		strField := ""
		customField, ok := customFields[field.ColumnName]
		if ok {
			strField = customField
		} else {
			null := "NULL"
			if !field.Nullable() {
				null = "NOT NULL"
			}
			switch field.DataType {
			case "varchar", "char":
				dataType := "VARCHAR"
				if field.DataType == "char" {
					dataType = "CHAR"
				}
				strField = field.ColumnName + " " + dataType + "(" + strconv.FormatInt(int64(field.CharacterMaximumLength.Int32), 10) + ") " + null
			case "date":
				strField = field.ColumnName + " DATE " + null
			case "datetime2":
				strField = field.ColumnName + " DATETIME " + null
			case "smallint":
				strField = field.ColumnName + " SMALLINT " + null
			case "int":
				strField = field.ColumnName + " INT " + null
			case "decimal":
				strField = field.ColumnName + " DECIMAL(" + strconv.FormatInt(int64(field.NumericPrecision.Int32), 10) + "," + strconv.FormatInt(int64(field.NumericScale.Int32), 10) + ") " + null
			case "uniqueidentifier":
				strField = field.ColumnName + " CHAR(38) NULL"
			}
		}
		strFields = append(strFields, "\t"+strField)
	}

	createTableStatement := "CREATE TABLE IF NOT EXISTS " + tblName + " (\n" + strings.Join(strFields, ",\n") + "\n) ENGINE=INNODB;"

	log.Println(createTableStatement)

	lib := gopohSqlDB.GetMySql()
	db := lib.Connect()
	if db == nil {
		log.Fatal("tidak dapat terkoneksi dengan mysql")
	}
	defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, createTableStatement)
	if err != nil {
		log.Fatalf("Error %s when creating table", err)
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error %s when getting rows affected", err)
		return 0, err
	}
	log.Printf("Rows affected when creating table: %d", rows)
	return rows, nil
}
*/
