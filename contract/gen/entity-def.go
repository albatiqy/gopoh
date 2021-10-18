package gen

import (
	_ "embed"
	"fmt"
	"os"
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
	// "github.com/albatiqy/gopoh/pkg/lib/decimal"
)

type EntityDef struct {
	TableName    string
	FieldDefs    map[string]driver.FieldDef
	KeyAttr      string
	KeyAuto      bool
	KeyCanUpdate bool
	SoftDelete   bool
	EntityName   string
	EAttrs       []string
}

var (
	//go:embed _embed/e-contr.txt
	txtEContr string
	//go:embed _embed/e-impl-mysql.txt
	txtEImplMysql string
	//go:embed _embed/e-impl-postgres.txt
	txtEImplPostgres string
	//go:embed _embed/e-impl-sqlserver.txt
	txtEImplSqlserver string
	//go:embed _embed/e-service.txt
	txtEService string
	//go:embed _embed/repo-utils.txt
	txtRepoUtils string
	//go:embed _embed/service-utils.txt
	txtServiceUtils string
)

func (obj EntityDef) GenerateContr(pathPrjDir string) {
	modName := util.GetModName(pathPrjDir)
	if modName == "" {
		log.Fatal("direktori project tidak valid")
	}

	pathContractDir := filepath.Join(pathPrjDir, "internal/core")
	if success, err := fs.MkDirIfNotExists(pathContractDir); !success {
		log.Fatal(err)
	}

	pathBakDir := filepath.Join(pathPrjDir, "_APPFS_/gopoh-gen/_bak") // _APPFS_ dari env aja=================================
	if success, err := fs.MkDirIfNotExists(pathBakDir); !success {
		log.Fatal(err)
	}

	//fs.BackupDir(pathContractDir, filepath.Join(pathBakDir, "e-contract.zip")) // q tidak usah

	lowerEntityName := strings.ToLower(obj.EntityName)

	fnameE := filepath.Join(pathContractDir, lowerEntityName+"-e.go")
	if fs.FileInfo(fnameE) != nil {
		log.Fatal("QueryDef.Generate: file sudah ada") //====================
	}
	/*
		if success, err := fs.BackupIfExist(fnameE, filepath.Join(pathBakDir, lowerEntityName+"-e.txt")); !success {
			log.Fatal(err)
		}
	*/

	var pickedFields map[string]driver.FieldDef

	if len(obj.EAttrs) == 0 {
		pickedFields = obj.FieldDefs
	} else {
		pickedFields = map[string]driver.FieldDef{}
		var newOrdinal uint16
		for attr, field := range obj.FieldDefs {
			field.Ordinal = newOrdinal
			pickedFields[attr] = field
			newOrdinal++
		}
	}

	var (
		varKeyName, keyStructField, keyType string
		keyIdx                              int = -1
	)
	if obj.KeyAttr == "" {
		log.Fatal("EntityDef.GenerateEContr: KeyAttr tidak didefinisikan")
	} else {
		if attrField, ok := pickedFields[obj.KeyAttr]; !ok {
			log.Fatal("EntityDef.GenerateEContr: missing KeyAttr in EAttrs")
		} else {
			keyStructField = attrField.StructField
			r := []rune(attrField.StructField)
			r[0] = unicode.ToLower(r[0])
			varKeyName = string(r)
			keyType = reflect.TypeOf(attrField.Type).Elem().String()
			keyIdx = int(attrField.Ordinal)
		}
	}

	fieldLen := len(pickedFields)

	strFields := make([]string, fieldLen)
	strFieldsTags := make([]string, fieldLen)

	var (
		strFieldsEInput []string
		strFieldsInput  []string
		importsE        []string
		useImportE      = map[string]string{}
	)

	for _, field := range pickedFields {
		fieldType := reflect.TypeOf(field.Type).Elem().String()
		switch field.Type.(type) {
		case *time.Time:
			useImportE["time"] = ""
			/*
				case *null.String: //  dan??
					fieldType = strings.Replace(fieldType, "contract.", "gopoh.", 1)
			*/
		}
		if int(field.Ordinal) == keyIdx {
			switch field.Type.(type) {
			case *uint64, *uint32, *uint, *int64, *int32, *int:
				strFields[field.Ordinal] = fmt.Sprintf("\t// %[1]s %[2]s", field.StructField+"String", "string")
				strFieldsTags[field.Ordinal] = fmt.Sprintf("`validate:\"%[1]s\" json:\"%[2]s\"` warning!, js bug conversion", "", field.JSON)
			default:
				strFields[field.Ordinal] = fmt.Sprintf("\t// %[1]s %[2]s", field.StructField, fieldType)
				strFieldsTags[field.Ordinal] = fmt.Sprintf("`validate:\"%[1]s\" json:\"%[2]s\"`", "", field.JSON)
			}
		} else {
			strFields[field.Ordinal] = fmt.Sprintf("\t%[1]s %[2]s", field.StructField, fieldType)
			strFieldsTags[field.Ordinal] = fmt.Sprintf("`validate:\"%[1]s\" json:\"%[2]s\"`", "", field.JSON)
		}
	}

	if obj.KeyCanUpdate {
		strFieldsEInput = make([]string, fieldLen)
		strFieldsInput = make([]string, fieldLen)
		// copy(strFieldsEInput, strFields)

		for i := range strFields {
			strFieldsEInput[i] = strFields[i] + " " + strFieldsTags[i]
			strFieldsInput[i] = strFields[i]
		}

	} else {
		newLen := fieldLen - 1
		var newIdx int
		strFieldsEInput = make([]string, newLen)
		strFieldsInput = make([]string, newLen)
		for i := range strFields {
			if i != keyIdx {
				strFieldsEInput[newIdx] = strFields[i] + " " + strFieldsTags[i]
				strFieldsInput[newIdx] = strFields[i]
				newIdx++
			}
		}
	}

	for impk, impv := range useImportE {
		if impv != "" {
			importsE = append(importsE, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsE = append(importsE, "\t\""+impl+`"`)
			}
		}
	}

	entityStructName := strings.ReplaceAll(obj.EntityName, "_", " ")
	entityStructName = strings.Title(entityStructName)
	entityStructName = strings.ReplaceAll(entityStructName, " ", "")

	tplData := map[string]string{
		"fieldsEInput":     strings.Join(strFieldsEInput, "\n"),
		"fieldsInput":      strings.Join(strFieldsInput, "\n"),
		"imports":          strings.Join(importsE, "\n"),
		"entityStructName": entityStructName,
		"keyStructField":   keyStructField,
		"varKeyName":       varKeyName,
		"keyType":          keyType,
	}

	util.WriteTplFile(fnameE, txtEContr, tplData)
}

func (obj EntityDef) GenerateImpl(pathPrjDir string, dbDriver string) {
	modName := util.GetModName(pathPrjDir)
	if modName == "" {
		log.Fatal("direktori project tidak valid")
	}

	pathServiceDir := filepath.Join(pathPrjDir, "internal/core/service")
	if success, err := fs.MkDirIfNotExists(pathServiceDir); !success {
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

	lowerDbDriver := strings.ToLower(dbDriver)
	lowerEntityName := strings.ToLower(obj.EntityName)

	fnameE := filepath.Join(pathRepositoryDir, lowerEntityName+"-e-"+lowerDbDriver+".go")
	if fs.FileInfo(fnameE) != nil {
		log.Fatal("QueryDef.Generate: file sudah ada") //====================
	}
	/*
		if success, err := fs.BackupIfExist(fnameE, filepath.Join(pathBakDir, lowerEntityName+"-e-"+lowerDbDriver+".txt")); !success {
			log.Fatal(err)
		}
	*/

	fnameRepoUtils := filepath.Join(pathRepositoryDir, "utils.go")
	if fs.FileInfo(fnameRepoUtils) == nil {
		fOut, err := os.Create(fnameRepoUtils)
		if err != nil {
			log.Fatal("EntityDef.GenerateEImpl: ", err)
		}
		defer fOut.Close()
		fOut.WriteString(txtRepoUtils)
	}

	fnameServiceUtils := filepath.Join(pathServiceDir, "utils.go")
	if fs.FileInfo(fnameServiceUtils) == nil {
		fOut, err := os.Create(fnameServiceUtils)
		if err != nil {
			log.Fatal("EntityDef.GenerateEImpl: ", err)
		}
		defer fOut.Close()
		fOut.WriteString(txtServiceUtils)
	}

	fnameEService := filepath.Join(pathServiceDir, lowerEntityName+"-e.go")

	var pickedFields map[string]driver.FieldDef

	if len(obj.EAttrs) == 0 {
		pickedFields = obj.FieldDefs
	} else {
		pickedFields = map[string]driver.FieldDef{}
		var newOrdinal uint16
		for attr, field := range obj.FieldDefs {
			field.Ordinal = newOrdinal
			pickedFields[attr] = field
			newOrdinal++
		}
	}

	var (
		varKeyName, keyType, keyCol, keyStructField, keyStructFieldStore string
		keyIdx                                                           int
	)
	if obj.KeyAttr == "" {
		log.Fatal("EntityDef.GenerateEImpl: KeyAttr tidak didefinisikan")
	} else {
		if attrField, ok := pickedFields[obj.KeyAttr]; !ok {
			log.Fatal("EntityDef.GenerateEImpl: missing KeyAttr in EAttrs")
		} else {
			keyStructField = attrField.StructField
			r := []rune(keyStructField)
			r[0] = unicode.ToLower(r[0])
			varKeyName = string(r)
			keyType = reflect.TypeOf(attrField.Type).Elem().String()
			keyCol = attrField.Col
			keyIdx = int(attrField.Ordinal)
		}
	}

	fieldLen := len(pickedFields)

	storeCols := make([]string, fieldLen)
	fieldArgs := make([]string, fieldLen)
	serviceFieldsCopy := make([]string, fieldLen)
	dbRequiredNotice := make([]string, fieldLen)

	var (
		updateCols        []string
		updateFieldArgs   []string
		storeFieldArgs    []string
		importsE          []string
		importsService    []string
		useImportE        = map[string]string{}
		useImportService  = map[string]string{}
		useIDConvertionFn string
	)

	for _, field := range pickedFields {
		storeCols[field.Ordinal] = field.Col
		if field.DBRequired {
			dbRequiredNotice[field.Ordinal] = " // No default"
		}
		serviceFieldsCopy[field.Ordinal] = "\t\t" + field.StructField + ": input." + field.StructField + ","
		if int(field.Ordinal) == keyIdx {
			switch field.Type.(type) {
			case *uint64, *uint32, *uint, *int64, *int32, *int:
				fieldArgs[field.Ordinal] = "\t\tinput." + field.StructField + "String, // warning, u must convert this first"
				keyStructFieldStore = keyStructField + "String"
				useIDConvertionFn = "1"
			default:
				fieldArgs[field.Ordinal] = "\t\tinput." + field.StructField + ","
				if reflect.TypeOf(field.Type).Elem().String() == "string" {
					useIDConvertionFn = "0"
					keyStructFieldStore = keyStructField
				} else {
					useIDConvertionFn = "2"
				}
			}
		} else {
			switch field.Type.(type) {
			case *time.Time:
				useImportE["time"] = ""
				fieldArgs[field.Ordinal] = "\t\tinput." + field.StructField + ".UTC(), // warning to UTC convertion!!================"
			case *null.Time:
				// useImportE["null"] = ""
				fieldArgs[field.Ordinal] = "\t\tnull.NewTime(input."+field.StructField+".Time.UTC(), input."+field.StructField+".Valid), // warning to UTC convertion!!================"
			default:
				fieldArgs[field.Ordinal] = "\t\tinput." + field.StructField + ","
			}
		}
	}

	var newIdx int
	storeFieldArgs = make([]string, fieldLen-1)
	for i := range storeCols {
		if i != keyIdx {
			storeFieldArgs[newIdx] = fieldArgs[i] + dbRequiredNotice[i]
			newIdx++
		}
	}

	if obj.KeyCanUpdate {
		updateCols = make([]string, fieldLen)
		updateFieldArgs = make([]string, fieldLen)
		for i, col := range storeCols {
			updateCols[i] = col
			updateFieldArgs[i] = fieldArgs[i]
		}
	} else {
		newLen := fieldLen - 1
		var newIdx int
		updateCols = make([]string, newLen)
		updateFieldArgs = make([]string, newLen)
		for i, col := range storeCols {
			if i != keyIdx {
				updateCols[newIdx] = col
				updateFieldArgs[newIdx] = fieldArgs[i]
				newIdx++
			}
		}
	}

	useImportE["internal"] = modName + "/internal"
	useImportE["core"] = modName + "/internal/core"

	softDeleteStr := ""
	if obj.SoftDelete {
		softDeleteStr = "1"
		useImportE["time"] = ""
	}

	for impk, impv := range useImportE {
		if impv != "" {
			importsE = append(importsE, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsE = append(importsE, "\t\""+impl+`"`)
			}
		}
	}

	useImportService["core"] = modName + "/internal/core"
	if useIDConvertionFn == "1" {
		useImportService["strconv"] = ""
	}

	for impk, impv := range useImportService {
		if impv != "" {
			importsService = append(importsService, "\t\""+impv+`"`)
		} else {
			if impl, ok := util.ImportsMap[impk]; ok {
				importsService = append(importsService, "\t\""+impl+`"`)
			}
		}
	}

	entityStructName := strings.ReplaceAll(obj.EntityName, "_", " ")
	entityStructName = strings.Title(entityStructName)
	entityStructName = strings.ReplaceAll(entityStructName, " ", "")

	r := []rune(entityStructName)
	r[0] = unicode.ToLower(r[0])
	varEntityStructName := string(r)

	r = []rune(lowerDbDriver)
	r[0] = unicode.ToUpper(r[0])
	dbDriverStr := string(r)

	genDriver := driver.Get(dbDriver)

	tplDataE := map[string]string{
		"imports":             strings.Join(importsE, "\n"),
		"entityStructName":    entityStructName,
		"storeFieldArgs":      strings.Join(storeFieldArgs, "\n"),
		"updateFieldArgs":     strings.Join(updateFieldArgs, "\n"),
		"sqlEStore":           genDriver.GenerateEStore(obj.TableName, storeCols), // jika tidak softDelete cek id ada tidak
		"sqlEUpdate":          genDriver.GenerateEUpdate(obj.TableName, updateCols, keyCol),
		"sqlEDelete":          genDriver.GenerateEDelete(obj.TableName, keyCol, obj.SoftDelete),
		"varKeyName":          varKeyName,
		"keyType":             keyType,
		"softDeleteStr":       softDeleteStr,
		"keyStructField":      keyStructField,
		"keyStructFieldStore": keyStructFieldStore,
		"dbDriverStr":         dbDriverStr,
	}

	var txtEImpl string
	switch dbDriver {
	case "mysql":
		txtEImpl = txtEImplMysql
	case "postgres":
		txtEImpl = txtEImplPostgres
	case "sqlserver":
		txtEImpl = txtEImplSqlserver
	}

	util.WriteTplFile(fnameE, txtEImpl, tplDataE)

	tplDataService := map[string]string{
		"imports":             strings.Join(importsService, "\n"),
		"entityStructName":    entityStructName,
		"varEntityStructName": varEntityStructName,
		"serviceFieldsCopy":   strings.Join(serviceFieldsCopy, "\n"),
		"varKeyName":          varKeyName,
		"keyStructField":      keyStructField,
		"keyType":             keyType,
		"useIDConvertionFn":   useIDConvertionFn,
	}

	util.WriteTplFile(fnameEService, txtEService, tplDataService)
}
