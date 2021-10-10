package sqlserver

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
	TableCatalog           string
	TableSchema            string
	TableName              string
	ColumnName             string
	OrdinalPosition        int
	ColumnDefault          null.String
	IsNullable             string
	DataType               string
	CharacterMaximumLength null.Int32
	CharacterOctetLength   null.Int32
	NumericPrecision       null.Int32
	NumericPrecisionRadix  null.Int32
	NumericScale           null.Int32
	DatetimePrecision      null.Int32
	CharacterSetCatalog    null.String
	CharacterSetSchema     null.String
	CharacterSetName       null.String
	CollationCatalog       null.String
	CollationSchema        null.String
	CollationName          null.String
	DomainCatalog          null.String
	DomainSchema           null.String
	DomainName             null.String
	Ordinal                uint16
}

type Driver struct {
}

func (d Driver) ReadTable(tblName, keyCol string, db *sql.DB) (*driver.TableData, error) {
	var tblSchema string
	if schema := strings.Split(tblName, "."); len(schema) != 2 {
		return nil, fmt.Errorf("tblName harus mengandung schema ex: schema.nama_tabel")
	} else {
		tblSchema = schema[0]
		tblName = schema[1]
	}

	fields := make(map[string]rawField)

	rows, err := db.Query(fmt.Sprintf(`
	SELECT TABLE_CATALOG,TABLE_SCHEMA,TABLE_NAME,COLUMN_NAME,ORDINAL_POSITION,COLUMN_DEFAULT,IS_NULLABLE,DATA_TYPE,CHARACTER_MAXIMUM_LENGTH,CHARACTER_OCTET_LENGTH,NUMERIC_PRECISION,NUMERIC_PRECISION_RADIX,NUMERIC_SCALE,DATETIME_PRECISION,CHARACTER_SET_CATALOG,CHARACTER_SET_SCHEMA,CHARACTER_SET_NAME,COLLATION_CATALOG,COLLATION_SCHEMA,COLLATION_NAME,DOMAIN_CATALOG,DOMAIN_SCHEMA,DOMAIN_NAME
	FROM INFORMATION_SCHEMA.columns
	WHERE TABLE_NAME = '%s' AND TABLE_SCHEMA='%s'
	ORDER BY ORDINAL_POSITION
	`,
		tblName, tblSchema))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if keyCol == "" {
		type primaryKey struct {
			ColumnName string
		}

		var primaryKeys []primaryKey

		rows1, err := db.Query(fmt.Sprintf(`
		select C.COLUMN_NAME FROM
		INFORMATION_SCHEMA.TABLE_CONSTRAINTS T
		JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE C
		ON C.CONSTRAINT_NAME=T.CONSTRAINT_NAME
		WHERE
		C.TABLE_NAME='%s' and C.TABLE_SCHEMA='%s'
		and T.CONSTRAINT_TYPE='PRIMARY KEY'`,
			tblName, tblSchema))
		if err != nil {
			return nil, err
		}
		defer rows1.Close()

		for rows1.Next() {
			pk := primaryKey{}
			err := rows1.Scan(&pk.ColumnName)
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
		err := rows.Scan(&field.TableCatalog, &field.TableSchema, &field.TableName, &field.ColumnName, &field.OrdinalPosition, &field.ColumnDefault, &field.IsNullable, &field.DataType, &field.CharacterMaximumLength, &field.CharacterOctetLength, &field.NumericPrecision, &field.NumericPrecisionRadix, &field.NumericScale, &field.DatetimePrecision, &field.CharacterSetCatalog, &field.CharacterSetSchema, &field.CharacterSetName, &field.CollationCatalog, &field.CollationSchema, &field.CollationName, &field.DomainCatalog, &field.DomainSchema, &field.DomainName)
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
		switch field.DataType {
		case "varchar", "char":
			if nullable {
				ftype = "null.String"
				useImport["null"] = ""
			} else {
				ftype = "string"
			}
		case "date", "datetime2": // time, smalldatetime
			if nullable {
				ftype = "null.Time"
				useImport["null"] = ""
			} else {
				ftype = "time.Time"
				useImport["time"] = ""
			}
		case "int", "smallint", "tinyint":
			if nullable {
				ftype = "null.Int32"
				useImport["null"] = ""
			} else {
				ftype = "int32"
			}
		case "bigint":
			if nullable {
				ftype = "null.Int64"
				useImport["null"] = ""
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
		case "real", "float":
			if nullable {
				ftype = "null.Float64"
				useImport["null"] = ""
			} else {
				ftype = "float64"
			}
		case "uniqueidentifier":
			if nullable {
				ftype = "sqlserver.NullUniqueIdentifier"
			} else {
				ftype = "sqlserver.UniqueIdentifier"
			}
			useImport["sqlserver"] = ""
		default:
			{
				return nil, fmt.Errorf("type " + field.DataType + " tidak terdefinisi")
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
	driver.LoadedDrivers["sqlserver"] = Driver{}
	log.Debugf("sqlserver gen: %s", "initialized")
}
