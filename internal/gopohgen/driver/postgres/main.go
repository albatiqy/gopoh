package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/internal/gopohgen/driver"
	"github.com/albatiqy/gopoh/pkg/lib/null"
)

const (
	softDeleteCol = "deleted_at"
)

type rawField struct {
	ColumnName             string
	OrdinalPosition        int
	ColumnDefault          null.String
	IsNullable             string
	UDTName                string
	CharacterMaximumLength null.Int32
	CharacterOctetLength   null.Int32
	NumericPrecision       null.Int32
	NumericPrecisionRadix  null.Int32
	NumericScale           null.Int32
	DatetimePrecision      null.Int32
	Ordinal                uint16
}

type Driver struct {
}

func (d Driver) ReadTable(tblName, keyCol string, db *sql.DB) (*driver.TableData, error) {
	fields := make(map[string]rawField)

	rows, err := db.Query(fmt.Sprintf(`
	SELECT column_name, ordinal_position, column_default, is_nullable, udt_name, character_maximum_length, character_octet_length, numeric_precision, numeric_precision_radix, numeric_scale, datetime_precision
	FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name ='%s' ORDER BY ordinal_position`,
		tblName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if keyCol == "" {
		type primaryKey struct {
			ColumnName string
			DataType   string
		}

		rows1, err := db.Query(fmt.Sprintf(`
	SELECT
	pg_attribute.attname,
	format_type(pg_attribute.atttypid, pg_attribute.atttypmod)
	FROM pg_index, pg_class, pg_attribute, pg_namespace
	WHERE
	pg_class.oid = '%s'::regclass AND
	indrelid = pg_class.oid AND
	nspname = 'public' AND
	pg_class.relnamespace = pg_namespace.oid AND
	pg_attribute.attrelid = pg_class.oid AND
	pg_attribute.attnum = any(pg_index.indkey)
	AND indisprimary`,
			tblName))
		if err != nil {
			return nil, err
		}
		defer rows1.Close()

		var primaryKeys []primaryKey
		for rows1.Next() {
			pk := primaryKey{}
			err := rows1.Scan(&pk.ColumnName, &pk.DataType)
			if err != nil {
				return nil, err
			}
			primaryKeys = append(primaryKeys, pk)
		}
		if len(primaryKeys) == 1 {
			keyCol = primaryKeys[0].ColumnName
		} else {
			return nil, fmt.Errorf("primary key tidak ditemukan atau lebih dari 1")
		}
	}

	ordinal := uint16(0)
	for rows.Next() {
		field := rawField{}
		err := rows.Scan(&field.ColumnName, &field.OrdinalPosition, &field.ColumnDefault, &field.IsNullable, &field.UDTName, &field.CharacterMaximumLength, &field.CharacterOctetLength, &field.NumericPrecision, &field.NumericPrecisionRadix, &field.NumericScale, &field.DatetimePrecision)
		if err != nil {
			return nil, err
		}
		field.Ordinal = ordinal
		ordinal++
		fields[field.ColumnName] = field
	}

	colsData := make([]driver.ColData, len(fields))

	keyAuto := false
	softDelete := false
	useImport := map[string]string{}
	for _, field := range fields {
		nullable := (field.IsNullable == "YES")
		colsData[field.Ordinal].Nullable = nullable

		required := true
		if field.ColumnDefault.Valid || nullable {
			required = false
		}
		colsData[field.Ordinal].DBRequired = required

		ftype := ""
		switch field.UDTName {
		case "bool":
			if nullable {
				ftype = "null.Bool"
				useImport["null"] = ""
			} else {
				ftype = "bool"
			}
		case "bpchar", "varchar":
			if nullable {
				ftype = "null.String"
				useImport["null"] = ""
			} else {
				ftype = "string"
			}
		case "time", "timetz", "timestamp": // time, smalldatetime // tz??
			if nullable {
				ftype = "null.Time"
				useImport["null"] = ""
			} else {
				ftype = "time.Time"
				useImport["time"] = ""
			}
		case "int2": //smallint
			if nullable {
				ftype = "null.Int32"
				useImport["null"] = ""
			} else {
				ftype = "int32"
			}
		case "int4":
			if nullable {
				ftype = "null.Int32"
				useImport["null"] = ""
			} else {
				ftype = "int32"
			}
		case "int8": // int8 => bigint
			if nullable {
				ftype = "null.Int64"
				useImport["null"] = ""
			} else {
				ftype = "int64"
			}
		case "numeric":
			if nullable {
				ftype = "decimal.NullDecimal"
			} else {
				ftype = "decimal.Decimal"
			}
			useImport["decimal"] = ""
		default:
			{
				return nil, fmt.Errorf("type " + field.UDTName + " tidak terdefinisi")
			}
		}

		if field.ColumnName == softDeleteCol {
			softDelete = true
		}

		structField := field.ColumnName
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
	driver.LoadedDrivers["postgres"] = Driver{}
	log.Debugf("postgres gen: %s", "initialized")
}
