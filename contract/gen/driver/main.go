package driver

var LoadedDrivers = make(map[string]Driver) // interface{}

type Driver interface {
	GenerateQ(tblName string, selectCols []string, softDelete bool) (string, string)
	//GenerateQByID(tblName string, selectCols []string, keyCol string, softDelete bool) string
	GenerateQESelects(tableSelectDefs []TableSelectDef, softDelete bool) (string, string, string, map[uint16]string)
	GenerateEStore(tblName string, storeCols []string) string
	GenerateEUpdate(tblName string, updateCols []string, keyCol string) string
	GenerateEDelete(tblName string, keyCol string, softDelete bool) string
}

type FieldDef struct {
	Col         string
	Ordinal     uint16
	StructField string
	Type        interface{}
	Searchable  bool //unused
	DBRequired  bool
	DBFn        string //unused
	JSON        string
	Label       string
}

type ColSelectDef struct {
	Ordinal  uint16
	Col      string
	ColAlias string
}

type TableSelectDef struct {
	TableName     string
	KeyCol        string
	ColSelectDefs []ColSelectDef
}

func Get(driverName string) Driver {
	driver, ok := LoadedDrivers[driverName]
	if ok {
		return driver //.(Driver)
	}
	return nil
}
