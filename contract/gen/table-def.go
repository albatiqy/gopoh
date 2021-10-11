package gen

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/albatiqy/gopoh/contract/gen/driver"
	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/pkg/lib/decimal"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
	"github.com/albatiqy/gopoh/pkg/lib/null"
)

type TableDef struct {
	DBEnvKey             string
	DBDriver             string
	TableName            string
	EntityName           string
	FieldDefs            map[string]driver.FieldDef
	KeyAttr              string
	KeyAuto              bool
	KeyCanUpdate         bool
	SoftDelete           bool
	EntityAttrs          []string
	OverridesStructField map[string]string
	OverridesType        map[string]interface{}
	OverridesJSON        map[string]string
	OverridesLabel       map[string]string
}

var (
	//go:embed _embed/main.txt
	txtMain string
	//go:embed _embed/entity-def.txt
	txtEntityDef string
)

func (obj TableDef) GenerateDBTarget(pathPrjDir, dbDriver string) {

	if util.GetModName(pathPrjDir) == "" {
		log.Fatal("direktori project tidak valid")
	}

	if _, ok := driver.LoadedDrivers[dbDriver]; !ok {
		log.Fatalf("gen driver\"%s\" tidak diload", dbDriver)
	}

	lowerDbDriver := strings.ToLower(dbDriver)

	pathSaveRoot := filepath.Join(pathPrjDir, "_APPFS_/gopoh-gen/db-target", lowerDbDriver)
	if success, err := fs.MkDirIfNotExists(pathSaveRoot); !success {
		log.Fatal(err)
	}

	nsTableName := strings.Replace(obj.TableName, ".", "_", 1)

	nsName := obj.DBEnvKey + "_" + nsTableName
	fnameQuery := filepath.Join(pathSaveRoot, nsName+".go")

	if obj.KeyAttr == "" {
		log.Fatal("TableDef.GenerateDBTarget: KeyAttr tidak didefinisikan")
	}

	var (
		strFields []string
		imports   []string
		useImport = map[string]string{}
	)

	fullAttrLen := len(obj.FieldDefs)
	if len(obj.EntityAttrs) == 0 {
		strFields = make([]string, fullAttrLen)
		obj.EntityAttrs = make([]string, fullAttrLen)
		for attr, field := range obj.FieldDefs {
			obj.EntityAttrs[field.Ordinal] = attr
		}
	} else {
		strFields = make([]string, len(obj.EntityAttrs))

		attrFields := map[string]int{}
		for i, attr := range obj.EntityAttrs {
			attrFields[attr] = i
		}
		if _, ok := attrFields[obj.KeyAttr]; !ok {
			log.Fatal("TableDef.GenerateDBTarget: KeyAttr not included")
		}
		for attr, field := range obj.FieldDefs {
			if field.DBRequired {
				if _, ok := attrFields[attr]; !ok {
					log.Fatal("TableDef.GenerateDBTarget: requidred attr not included")
				}
			}
		}
	}

	newJsonKeyAttr := obj.KeyAttr

	for i, attr := range obj.EntityAttrs {
		if field, ok := obj.FieldDefs[attr]; ok {
			switch field.Type.(type) {
			case *time.Time:
				useImport["time"] = ""
			case *null.String:
				useImport["null"] = ""
			case *decimal.Decimal, *decimal.NullDecimal:
				useImport["decimal"] = ""
			}

			overrideType := reflect.TypeOf(field.Type).Elem().String()
			if obj.OverridesType != nil {
				if fieldType, ok := obj.OverridesType[attr]; ok {
					overrideType = reflect.TypeOf(fieldType).Elem().String()
				}
			}

			overrideJSON := field.JSON
			if obj.OverridesJSON != nil {
				if fieldJSON, ok := obj.OverridesJSON[attr]; ok {
					overrideJSON = fieldJSON
				}
			}

			if attr == obj.KeyAttr {
				newJsonKeyAttr = overrideJSON
			}

			overrideLabel := field.Label
			if obj.OverridesLabel != nil {
				if fieldLabel, ok := obj.OverridesLabel[attr]; ok {
					overrideLabel = fieldLabel
				}
			}

			overrideStructField := strings.ReplaceAll(overrideLabel, " ", "")
			if obj.OverridesStructField != nil {
				if structField, ok := obj.OverridesStructField[attr]; ok {
					overrideStructField = structField
				}
			}

			dbRequiredStr := ""
			if field.DBRequired {
				dbRequiredStr = ", DBRequired: true"
			}

			// harus sinkron dengan driver
			strFields[i] = fmt.Sprintf("\t\t\t"+`"%[1]s": {StructField: "%[2]s", Col: "%[1]s", Type: (*%[3]s)(nil), JSON: "%[1]s", Label: "%[4]s", Ordinal: %[5]d%[6]s},`, overrideJSON, overrideStructField, overrideType, overrideLabel, i, dbRequiredStr)
		} else {
			log.Fatal("TableDef.GenerateDBTarget: wrong EntityAttrs entry")
		}
	}

	for impk, impv := range useImport {
		if impv != "" {
			imports = append(imports, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				imports = append(imports, "\t\""+impl+`"`)
			}
		}
	}

	keyAutoStr := "false"
	if obj.KeyAuto {
		keyAutoStr = "true"
		obj.KeyCanUpdate = false
	}

	keyCanUpdateStr := "false"
	if obj.KeyCanUpdate {
		keyCanUpdateStr = "true"
	}

	softDeleteStr := "false"
	if obj.SoftDelete {
		softDeleteStr = "true"
	}

	completeAttrList := make([]string, fullAttrLen)
	for attr, field := range obj.FieldDefs {
		completeAttrList[field.Ordinal] = attr
	}

	tplData := map[string]string{
		"fields":           strings.Join(strFields, "\n"),
		"imports":          strings.Join(imports, "\n"),
		"nsName":           nsName,
		"tableName":        obj.TableName,
		"nsTableName":      nsTableName,
		"entityName":       obj.EntityName,
		"keyAttr":          newJsonKeyAttr,
		"keyAutoStr":       keyAutoStr,
		"keyCanUpdateStr":  keyCanUpdateStr,
		"softDeleteStr":    softDeleteStr,
		"completeAttrList": strings.Join(completeAttrList, ","),
		"selectedDBDriver": dbDriver,
	}

	util.WriteTplFile(fnameQuery, txtEntityDef, tplData)

	fnameMain := filepath.Join(pathSaveRoot, "main.go")

	if fs.FileInfo(fnameMain) == nil {
		tplData := map[string]string{}
		util.WriteTplFile(fnameMain, txtMain, tplData)
	}
}
