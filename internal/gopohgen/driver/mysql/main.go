package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/internal/gopohgen/driver"
	"github.com/albatiqy/gopoh/pkg/lib/null"
)

const (
	softDeleteCol = "deleted_at"
)

type rawField struct {
	Field         string
	Type          string
	DataType      string
	MaxCharLength string
	Null          string
	Key           string
	Default       null.String
	Extra         string
	Unsigned      bool
	Ordinal       uint16
	SizeString    string
}

type Driver struct {
}

func (d Driver) ReadTable(tblName, keyCol string, db *sql.DB) (*driver.TableData, error) {
	fields := make(map[string]rawField)

	rows, err := db.Query("DESCRIBE " + tblName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regex, err := regexp.Compile(`(\w+)\(*([^\)]*)\)*`)
	if err != nil {
		log.Fatalf("db error: %s", err)
	}

	ordinal := uint16(0)
	for rows.Next() {
		field := rawField{}
		err := rows.Scan(&field.Field, &field.Type, &field.Null, &field.Key, &field.Default, &field.Extra)
		if err != nil {
			return nil, err
		}
		match := regex.FindStringSubmatch(field.Type)
		field.DataType = match[1]
		switch field.DataType {
		case "decimal":
			field.SizeString = match[2]
		case "char", "varchar":
			field.MaxCharLength = match[2]
		}
		if strings.Contains(field.Type, "unsigned") { // filter numerik type??
			field.Unsigned = true
		}
		field.Ordinal = ordinal
		ordinal++
		fields[field.Field] = field
	}

	colsData := make([]driver.ColData, len(fields))

	keyAuto := false
	softDelete := false
	useImport := map[string]string{}
	for _, field := range fields {
		nullable := (field.Null == "YES")
		colsData[field.Ordinal].Nullable = nullable
		unsigned := strings.Contains(field.Type, "unsigned")

		required := true
		if field.Default.Valid || nullable {
			required = false
		}
		colsData[field.Ordinal].DBRequired = required

		ftype := ""
		switch field.DataType { // HANDLE NULL TYPE ===================================
		case "varchar", "char":
			if nullable {
				ftype = "null.String"
				useImport["null"] = ""
			} else {
				ftype = "string"
			}
		case "datetime", "timestamp", "date":
			if nullable {
				ftype = "null.Time"
				useImport["null"] = ""
			} else {
				ftype = "time.Time"
				useImport["time"] = ""
			}
		case "int": // isnull???
			if unsigned {
				ftype = "uint32"
			} else {
				ftype = "int32"
			}
		case "tinyint":
			if unsigned {
				ftype = "uint8"
			} else {
				ftype = "byte"
			}
		case "smallint":
			if unsigned {
				ftype = "uint16"
			} else {
				ftype = "int16"
			}
		case "bigint":
			if unsigned {
				ftype = "uint64"
			} else {
				ftype = "int64"
			}
		case "decimal":
			if nullable {
				ftype = "decimal.NullDecimal"
			} else {
				ftype = "decimal.Decimal"
			}
			useImport["decimal"] = ""
		case "float":
			ftype = "float32"
		case "double":
			ftype = "float64"
		default:
			return nil, fmt.Errorf("type " + field.DataType + " tidak terdefinisi")
		}
		if keyCol == "" {
			if field.Key == "PRI" {
				keyCol = field.Field
				if strings.Contains(field.Extra, "AUTO_INCREMENT") {
					keyAuto = true
				}
			}
		}
		if field.Field == softDeleteCol {
			softDelete = true
		}

		structField := field.Field
		colsData[field.Ordinal].Name = structField // JSON dan
		colsData[field.Ordinal].JSON = structField // Col ditentukan disini
		structField = strings.ReplaceAll(structField, "_", " ")
		structField = strings.Title(structField)
		colsData[field.Ordinal].Label = structField
		colsData[field.Ordinal].CompatibleGoTypeStr = ftype
	}
	tableData := driver.NewTableData(colsData, keyCol, keyAuto, softDelete, useImport)
	return tableData, nil
}

func init() {
	driver.LoadedDrivers["mysql"] = Driver{}
	log.Debugf("mysql gen: %s", "initialized")
}
