package gen

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/albatiqy/gopoh/contract/gen/driver"
	"github.com/albatiqy/gopoh/contract/gen/util"
	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/pkg/lib/fs"
	"github.com/albatiqy/gopoh/pkg/lib/null"
)

var (
	//go:embed _embed/qe-contr.txt
	txtQEContr string
	//go:embed _embed/qe-colsmap-mysql.txt
	txtQEColsMapMysql string
	//go:embed _embed/qe-colsmap-postgres.txt
	txtQEColsMapPostgres string
	//go:embed _embed/qe-fieldsmap.txt
	txtQEFieldsMap string
	//go:embed _embed/qe-impl-mysql.txt
	txtQEImplMysql string
	//go:embed _embed/qe-impl-postgres.txt
	txtQEImplPostgres string
)

type FieldOverride struct {
	Label       string
	JSON        string
	StructField string
}

type QueryEntity struct {
	EntityDef       EntityDef
	ParentEntityDef *EntityDef
	EntityAttrs     []string
	FieldsOverride  map[string]FieldOverride
}

type QueryDef struct {
	QueryName     string
	QueryEntities []QueryEntity
}

// softDelete, joinType ???
func (obj QueryDef) Generate(pathPrjDir, fName, dbDriver string) { // harus dihapus tidak boleh nimpa
	modName := util.GetModName(pathPrjDir)
	if modName == "" {
		log.Fatal("direktori project tidak valid")
	}

	qeLen := len(obj.QueryEntities)
	if qeLen < 1 {
		log.Fatal("QueryDef.Generate: paling tidak harus ada 1 EntityDef") // 2
	}

	pathContractDir := filepath.Join(pathPrjDir, "internal/core")
	if success, err := fs.MkDirIfNotExists(pathContractDir); !success {
		log.Fatal(err)
	}

	pathRepositoryDir := filepath.Join(pathPrjDir, "internal/repository")
	if success, err := fs.MkDirIfNotExists(pathRepositoryDir); !success {
		log.Fatal(err)
	}

	pathBakDir := filepath.Join(pathPrjDir, "_APPFS_/gopoh-gen/_bak")
	if success, err := fs.MkDirIfNotExists(pathBakDir); !success {
		log.Fatal(err)
	}

	if obj.QueryName == "" {
		obj.QueryName = obj.QueryEntities[0].EntityDef.EntityName
	}

	lowerDbDriver := strings.ToLower(dbDriver)
	lowerQueryName := strings.ToLower(obj.QueryName)

	fnameQE := filepath.Join(pathContractDir, lowerQueryName+"-qe.go")
	if fs.FileInfo(fnameQE) != nil {
		log.Fatal("QueryDef.Generate: file sudah ada") //====================
	}

	fnameColsMap := filepath.Join(pathRepositoryDir, lowerQueryName+"-qe-colsmap-"+lowerDbDriver+".go")
	fnameImpl := filepath.Join(pathRepositoryDir, lowerQueryName+"-qe-"+lowerDbDriver+".go")

	fnameQCmd := filepath.Join(pathPrjDir, "_APPFS_/gopoh-gen/db-target", lowerDbDriver, fName+".go")
	// backup self
	if _, err := fs.BackupIfExist(fnameQCmd, filepath.Join(pathBakDir, "qe-"+fName+".txt")); err != nil {
		log.Fatal(err)
	}

	keyAttr1 := obj.QueryEntities[0].EntityDef.KeyAttr
	keyStructField := ""
	if keyAttrField, ok := obj.QueryEntities[0].EntityDef.FieldDefs[keyAttr1]; ok {
		keyStructField = keyAttrField.StructField
	}

	tableSelectDefs := make([]driver.TableSelectDef, qeLen)

	pickedFields := make(map[string]driver.FieldDef)

	keyAttr := ""

	var newFieldOrdinal uint16
	for i, queryEntity := range obj.QueryEntities {
		tableSelectDefs[i].TableName = queryEntity.EntityDef.TableName
		if queryEntity.EntityDef.KeyAttr != "" {
			if keyField, ok := queryEntity.EntityDef.FieldDefs[queryEntity.EntityDef.KeyAttr]; ok { // harusnya ga perlu if
				tableSelectDefs[i].KeyCol = keyField.Col
			} else {
				log.Fatal("QueryDef.Generate: wrong KeyAttr")
			}
		} else {
			log.Fatal("QueryDef.Generate: KeyAttr tidak terdefinisi")
		}

		if len(queryEntity.EntityAttrs) == 0 {
			queryEntity.EntityAttrs = make([]string, len(queryEntity.EntityDef.FieldDefs))
			for jsonAttr, field := range queryEntity.EntityDef.FieldDefs {
				queryEntity.EntityAttrs[field.Ordinal] = jsonAttr
			}
		}
		for _, jsonAttr := range queryEntity.EntityAttrs {
			fieldDef, ok := queryEntity.EntityDef.FieldDefs[jsonAttr]
			if ok {
				newAttr := jsonAttr
				if queryEntity.FieldsOverride != nil {
					if fieldOverride, ok := queryEntity.FieldsOverride[jsonAttr]; ok {
						if fieldOverride.JSON == "" {
							log.Fatal("QueryDef.Generate: jsonAttr kosong")
						}
						if fieldOverride.JSON == jsonAttr {
							log.Fatal("QueryDef.Generate: jsonAttr tidak berubah")
						}
						newAttr = fieldOverride.JSON
						fieldDef.JSON = fieldOverride.JSON
						if fieldOverride.Label != "" {
							fieldDef.Label = fieldOverride.Label
						}
						if fieldOverride.StructField != "" {
							fieldDef.StructField = fieldOverride.StructField
						}
					}
				}
				if _, ok := pickedFields[newAttr]; ok {
					log.Fatal("QueryDef.Generate: jsonAttr bentrok")
				}
				if i == 0 {
					if keyAttr1 == jsonAttr {
						keyAttr = newAttr
					}
				}
				fieldDef.Ordinal = newFieldOrdinal
				tableSelectDefs[i].ColSelectDefs = append(tableSelectDefs[i].ColSelectDefs, driver.ColSelectDef{ // ini....
					Col:      fieldDef.Col,
					ColAlias: fieldDef.JSON,
					Ordinal:  newFieldOrdinal,
				})
				pickedFields[newAttr] = fieldDef
				newFieldOrdinal++
			} else {
				log.Fatal("QueryDef.Generate: jsonAttr tidak terdefinisi")
			}
		}
	}

	var (
		varKeyName, keyType, keyStructFieldFindAll string
		keyIdx                                     int
	)
	if keyAttr == "" {
		log.Fatal("QueryDef.Generate: KeyAttr tidak didefinisikan")
	} else {
		if attrField, ok := pickedFields[keyAttr]; !ok {
			log.Fatal("QueryDef.Generate: missing KeyAttr in EAttrs")
		} else {
			keyStructFieldFindAll = keyStructField
			keyType = reflect.TypeOf(attrField.Type).Elem().String()
			keyIdx = int(attrField.Ordinal)
			r := []rune(attrField.StructField)
			r[0] = unicode.ToLower(r[0])
			varKeyName = string(r)
		}
	}

	fieldLen := len(pickedFields)

	strMap := make([]string, fieldLen)
	strFieldsE := make([]string, fieldLen)
	fieldScansQ := make([]string, fieldLen)
	fieldScansQByID := make([]string, fieldLen)

	var (
		strFieldsQ       []string
		importsQE        []string
		importsImpl      []string
		importsColsMap   []string
		qTimeLocalA      []string
		eTimeLocalA      []string
		useImportQE      = map[string]string{}
		useImportImpl    = map[string]string{}
		useImportColsMap = map[string]string{}
	)

	genDriver := driver.Get(dbDriver)
	sqlSelectAll, softDeleteSelectAll, sqlSelectByID, colsMap := genDriver.GenerateQESelects(tableSelectDefs, obj.QueryEntities[0].EntityDef.SoftDelete)

	for newAttr, field := range pickedFields {
		fieldType := reflect.TypeOf(field.Type).Elem().String()
		switch field.Type.(type) { // uneficient code
		case *time.Time:
			useImportQE["time"] = ""
			/*
				case *null.String: // dan??
					fieldType = strings.Replace(fieldType, "contract.", "gopoh.", 1)
			*/
		}

		strMap[field.Ordinal] = "\t\t\t\"" + newAttr + "\": \"" + colsMap[field.Ordinal] + "\","

		if int(field.Ordinal) == keyIdx {
			switch field.Type.(type) {
			case *uint64, *uint32, *uint, *int64, *int32, *int:
				strFieldsE[field.Ordinal] = fmt.Sprintf("\t%[1]s %[2]s `json:\"%[3]s\"`", field.StructField+"String", "string", field.JSON)
				fieldScansQ[field.Ordinal] = "\t\t\t&record." + field.StructField + "String,"
				fieldScansQByID[field.Ordinal] = "\t\t&record." + field.StructField + "String,"
				keyStructFieldFindAll = keyStructField + "String"
			default:
				strFieldsE[field.Ordinal] = fmt.Sprintf("\t%[1]s %[2]s `json:\"%[3]s\"`", field.StructField, fieldType, field.JSON)
				fieldScansQ[field.Ordinal] = "\t\t\t&record." + field.StructField + ","
				fieldScansQByID[field.Ordinal] = "\t\t&record." + field.StructField + ","
			}
		} else {
			switch field.Type.(type) {
			case *time.Time, *null.Time:
				useImportQE["time"] = ""
				fieldScansQ[field.Ordinal] = "\t\t\t&record." + field.StructField + ", // warning from UTC result"
				fieldScansQByID[field.Ordinal] = "\t\t&record." + field.StructField + ", // warning from UTC result"
				qTimeLocalA = append(qTimeLocalA, "\t\trecord."+field.StructField+" = record."+field.StructField+".Local() // convert to local")
				eTimeLocalA = append(eTimeLocalA, "\t\trecord."+field.StructField+" = record."+field.StructField+".Local() // convert to local")
			default:
				fieldScansQ[field.Ordinal] = "\t\t\t&record." + field.StructField + ","
				fieldScansQByID[field.Ordinal] = "\t\t&record." + field.StructField + ","
			}
			strFieldsE[field.Ordinal] = fmt.Sprintf("\t%[1]s %[2]s `json:\"%[3]s\"`", field.StructField, fieldType, field.JSON)
		}
	}

	strFieldsQ = make([]string, len(strFieldsE))
	copy(strFieldsQ, strFieldsE)
	/*
		for i := range strFieldsE {
			strFieldsQ[i] = strFieldsE[i]
		}
	*/

	for impk, impv := range useImportQE {
		if impv != "" {
			importsQE = append(importsQE, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsQE = append(importsQE, "\t\""+impl+`"`)
			}
		}
	}

	queryStructName := strings.ReplaceAll(obj.QueryName, "_", " ")
	queryStructName = strings.Title(queryStructName)
	queryStructName = strings.ReplaceAll(queryStructName, " ", "")

	tplDataQE := map[string]string{
		"imports":         strings.Join(importsQE, "\n"),
		"fieldsE":         strings.Join(strFieldsE, "\n"),
		"fieldsQ":         strings.Join(strFieldsQ, "\n"),
		"queryStructName": queryStructName,
		"varKeyName":      varKeyName,
		"keyType":         keyType,
	}

	util.WriteTplFile(fnameQE, txtQEContr, tplDataQE)

	useImportImpl["core"] = modName + "/internal/core"

	if obj.QueryEntities[0].EntityDef.SoftDelete {
		useImportImpl["time"] = ""
	}

	for impk, impv := range useImportImpl {
		if impv != "" {
			importsImpl = append(importsImpl, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsImpl = append(importsImpl, "\t\""+impl+`"`)
			}
		}
	}

	r := []rune(lowerDbDriver)
	r[0] = unicode.ToUpper(r[0])
	dbDriverStr := string(r)

	tplDataImpl := map[string]string{
		//"fields":                strings.Join(strFields, "\n"),
		"imports":               strings.Join(importsImpl, "\n"),
		"queryStructName":       queryStructName,
		"fieldScansQ":           strings.Join(fieldScansQ, "\n"),
		"qTimeLocal":            strings.Join(qTimeLocalA, "\n"),
		"fieldScansQByID":       strings.Join(fieldScansQByID, "\n"),
		"eTimeLocal":            strings.Join(eTimeLocalA, "\n"),
		"sqlSelectAll":          sqlSelectAll,
		"sqlSelectByID":         sqlSelectByID,
		"whereSoftDelete":       softDeleteSelectAll,
		"keyAttr":               keyAttr,
		"keyStructField":        keyStructField,
		"dbDriverStr":           dbDriverStr,
		"varKeyName":            varKeyName,
		"keyStructFieldFindAll": keyStructFieldFindAll,
		"keyType":               keyType,
	}

	var txtQEImpl, txtQEColsMap string
	switch dbDriver {
	case "mysql":
		txtQEImpl = txtQEImplMysql
		txtQEColsMap = txtQEColsMapMysql
	case "postgres":
		txtQEImpl = txtQEImplPostgres
		txtQEColsMap = txtQEColsMapPostgres
	}

	util.WriteTplFile(fnameImpl, txtQEImpl, tplDataImpl)

	for impk, impv := range useImportColsMap {
		if impv != "" {
			importsColsMap = append(importsColsMap, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsColsMap = append(importsColsMap, "\t\""+impl+`"`)
			}
		}
	}

	tplDataColsMap := map[string]string{
		"maps":            strings.Join(strMap, "\n"),
		"imports":         strings.Join(importsColsMap, "\n"),
		"queryStructName": queryStructName,
		"keyAttr":         keyAttr,
		"dbDriverStr":     dbDriverStr,
	}

	util.WriteTplFile(fnameColsMap, txtQEColsMap, tplDataColsMap)
}

func (obj QueryDef) GenerateFieldsMap(pathPrjDir string) {
	modName := util.GetModName(pathPrjDir)
	if modName == "" {
		log.Fatal("direktori project tidak valid")
	}

	qeLen := len(obj.QueryEntities)
	if qeLen < 1 {
		log.Fatal("QueryDef.Generate: paling tidak harus ada 1 EntityDef") // 2
	}

	pathServiceDir := filepath.Join(pathPrjDir, "internal/core/service")
	if success, err := fs.MkDirIfNotExists(pathServiceDir); !success {
		log.Fatal(err)
	}

	pathBakDir := filepath.Join(pathPrjDir, "_APPFS_/gopoh-gen/_bak")
	if success, err := fs.MkDirIfNotExists(pathBakDir); !success {
		log.Fatal(err)
	}

	if obj.QueryName == "" {
		obj.QueryName = obj.QueryEntities[0].EntityDef.EntityName
	}

	lowerQueryName := strings.ToLower(obj.QueryName)

	fnameMap := filepath.Join(pathServiceDir, lowerQueryName+"-qe-fieldsmap.go")
	if fs.FileInfo(fnameMap) != nil {
		log.Fatal("QueryDef.Generate: file sudah ada") //====================
	}

	keyAttr1 := obj.QueryEntities[0].EntityDef.KeyAttr

	tableSelectDefs := make([]driver.TableSelectDef, qeLen)

	pickedFields := make(map[string]driver.FieldDef)

	keyAttr := ""

	var newFieldOrdinal uint16
	for i, queryEntity := range obj.QueryEntities {
		tableSelectDefs[i].TableName = queryEntity.EntityDef.TableName
		if queryEntity.EntityDef.KeyAttr != "" {
			if keyField, ok := queryEntity.EntityDef.FieldDefs[queryEntity.EntityDef.KeyAttr]; ok {
				tableSelectDefs[i].KeyCol = keyField.Col
			} else {
				log.Fatal("QueryDef.Generate: wrong KeyAttr")
			}
		} else {
			log.Fatal("QueryDef.Generate: KeyAttr tidak terdefinisi")
		}

		if len(queryEntity.EntityAttrs) == 0 {
			queryEntity.EntityAttrs = make([]string, len(queryEntity.EntityDef.FieldDefs))
			for jsonAttr, field := range queryEntity.EntityDef.FieldDefs {
				queryEntity.EntityAttrs[field.Ordinal] = jsonAttr
			}
		}
		for _, jsonAttr := range queryEntity.EntityAttrs {
			fieldDef, ok := queryEntity.EntityDef.FieldDefs[jsonAttr]
			if ok {
				newAttr := jsonAttr
				if queryEntity.FieldsOverride != nil {
					if fieldOverride, ok := queryEntity.FieldsOverride[jsonAttr]; ok {
						if fieldOverride.JSON == "" {
							log.Fatal("QueryDef.Generate: jsonAttr kosong")
						}
						if fieldOverride.JSON == jsonAttr {
							log.Fatal("QueryDef.Generate: jsonAttr tidak berubah")
						}
						newAttr = fieldOverride.JSON
						fieldDef.JSON = fieldOverride.JSON
						if fieldOverride.Label != "" {
							fieldDef.Label = fieldOverride.Label
						}
						if fieldOverride.StructField != "" {
							fieldDef.StructField = fieldOverride.StructField
						}
					}
				}
				if _, ok := pickedFields[newAttr]; ok {
					log.Fatal("QueryDef.Generate: jsonAttr bentrok")
				}
				if i == 0 {
					if keyAttr1 == jsonAttr {
						keyAttr = newAttr
					}
				}
				fieldDef.Ordinal = newFieldOrdinal
				tableSelectDefs[i].ColSelectDefs = append(tableSelectDefs[i].ColSelectDefs, driver.ColSelectDef{
					Col:      fieldDef.Col,
					ColAlias: fieldDef.JSON,
					Ordinal:  newFieldOrdinal,
				})
				pickedFields[newAttr] = fieldDef
				newFieldOrdinal++
			} else {
				log.Fatal("QueryDef.Generate: jsonAttr tidak terdefinisi")
			}
		}
	}

	var (
		warningText string
		keyIdx      int
	)
	if keyAttr == "" {
		log.Fatal("QueryDef.Generate: KeyAttr tidak didefinisikan")
	} else {
		if attrField, ok := pickedFields[keyAttr]; !ok {
			log.Fatal("QueryDef.Generate: missing KeyAttr in EAttrs")
		} else {
			keyIdx = int(attrField.Ordinal)
		}
	}

	fieldLen := len(pickedFields)

	strFieldsMap := make([]string, fieldLen)
	strLabelsMap := make([]string, fieldLen)

	var (
		importsMap   []string
		useImportMap = map[string]string{}
	)

	for newAttr, field := range pickedFields {
		if int(field.Ordinal) == keyIdx {
			switch field.Type.(type) {
			case *uint64, *uint32, *uint, *int64, *int32, *int:
				strFieldsMap[field.Ordinal] = "\t\t\t\"" + newAttr + "\": \"" + field.StructField + "String\", // hati-hati konversi"
				strLabelsMap[field.Ordinal] = "\t\t\t\"" + field.StructField + "String\": \"" + field.Label + "\", // hati-hati konversi"
				warningText = "warning id convertion exists===================================="
			default:
				strFieldsMap[field.Ordinal] = "\t\t\t\"" + newAttr + "\": \"" + field.StructField + "\","
				strLabelsMap[field.Ordinal] = "\t\t\t\"" + field.StructField + "\": \"" + field.Label + "\","
			}
		} else {
			strFieldsMap[field.Ordinal] = "\t\t\t\"" + newAttr + "\": \"" + field.StructField + "\","
			strLabelsMap[field.Ordinal] = "\t\t\t\"" + field.StructField + "\": \"" + field.Label + "\","
		}
	}

	queryStructName := strings.ReplaceAll(obj.QueryName, "_", " ")
	queryStructName = strings.Title(queryStructName)
	queryStructName = strings.ReplaceAll(queryStructName, " ", "")

	for impk, impv := range useImportMap {
		if impv != "" {
			importsMap = append(importsMap, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsMap = append(importsMap, "\t\""+impl+`"`)
			}
		}
	}

	tplDataMap := map[string]string{
		"fieldsMap":       strings.Join(strFieldsMap, "\n"),
		"labelsMap":       strings.Join(strLabelsMap, "\n"),
		"imports":         strings.Join(importsMap, "\n"),
		"queryStructName": queryStructName,
		"keyAttr":         keyAttr,
		"warningText":     warningText,
	}

	util.WriteTplFile(fnameMap, txtQEFieldsMap, tplDataMap)
}
