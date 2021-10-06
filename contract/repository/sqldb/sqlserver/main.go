package sqlserver

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/albatiqy/gopoh/contract/repository"
	"github.com/albatiqy/gopoh/contract/repository/sqldb"

	// "github.com/albatiqy/gopoh/pkg/lib/env"
	"github.com/albatiqy/gopoh/contract/log"
)

type DriverSpec struct {
}

func init() {
	sqldb.DriversSpec["sqlserver"] = DriverSpec{}
	log.Debugf("sqldb: %s", "sqlserver driver initialized")
}

var (
	filterType = map[string]string{
		"like": "",
		"eq":   "=",
		"neq":  "<>",
		"lt":   "<",
		"lte":  "<=",
		"gt":   ">",
		"gte":  ">=",
	}
	orderType = map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}
)

// dsn := "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
// db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})

func (spec DriverSpec) Open(dbSetting *sqldb.DBSetting) *sql.DB {

	// sqlserver://sa:mypass@localhost:1234?database=master&connection+timeout=30
	// sqlserver://sa:my%7Bpass@somehost?connection+timeout=30

	//DSN: "sqlserver://@host/instance?database=database&connection+timeout=30&parseTime=true&loc=UTC"

	db, err := sql.Open("sqlserver", dbSetting.DSN)
	if err != nil {
		log.Debugf(`sqldb: %s`, err)
		return nil
	}

	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(25)
	// db.SetConnMaxLifetime(5*time.Minute)

	return db
}

func (spec DriverSpec) BuildFinderCursorQuery(cursorID string, isPrevNav bool, baseQuery string, finderOptionCursor repository.FinderOptionCursor, whereRaws []string, colsMap *sqldb.ColsMap) (string, []interface{}, error) {
	var args []interface{}
	var bbBuilder strings.Builder
	z := int(1)
	if len(finderOptionCursor.Filters) > 0 {
		for _, qFilter := range finderOptionCursor.Filters {
			if fType, ok := filterType[qFilter.Type]; ok {
				if fCol, ok := colsMap.Cols[qFilter.Attr]; ok {
					if fType == "like" { // atomic ops !! =================================
						bbBuilder.WriteString("(" + fCol + " LIKE $" + strconv.Itoa(z) +") AND")
						args = append(args, "%"+qFilter.Val+"%")
					} else {
						bbBuilder.WriteString("(" + fCol + fType + "$" + strconv.Itoa(z) +") AND")
						args = append(args, qFilter.Val) // sync args and z???
					}
					z++
				}
			}
		}
	}

	keyCol, ok := colsMap.Cols[colsMap.KeyAttr]
	if !ok {
		return "", nil, fmt.Errorf("key attr not defined")
	}

	if cursorID != "" {
		if isPrevNav { // atomic===============================
			whereRaws = append(whereRaws, keyCol+" < $" + strconv.Itoa(z))
		} else {
			whereRaws = append(whereRaws, keyCol+" > $" + strconv.Itoa(z))
		}
		args = append(args, cursorID)
		z++
	}

	for _, where := range whereRaws {
		bbBuilder.WriteString("(" + where + ") AND")
	}

	if bbBuilder.Len() > 0 {
		strWhere := bbBuilder.String()
		baseQuery += " WHERE " + strWhere[:bbBuilder.Len()-4] // hati2 rune length
	}

	bbBuilder.Reset()
	if len(finderOptionCursor.Orders) > 0 {
		var oKey bool
		for _, qOrder := range finderOptionCursor.Orders {
			if oType, ok := orderType[qOrder.Type]; ok {
				if oCol, ok := colsMap.Cols[qOrder.Attr]; ok {
					if qOrder.Attr == colsMap.KeyAttr {
						keyOrder := "ASC"
						if isPrevNav {
							keyOrder = "DESC"
						}
						bbBuilder.WriteString(keyCol + " " + keyOrder + ",")
						oKey = true
					} else {
						bbBuilder.WriteString(oCol + " " + oType + ",")
					}
				}
			}
		}
		if bbBuilder.Len() > 0 {
			if !oKey {
				keyOrder := "ASC"
				if isPrevNav {
					keyOrder = "DESC"
				}
				bbBuilder.WriteString(keyCol + " " + keyOrder + ",")
			}
			strOrder := bbBuilder.String()
			baseQuery += " ORDER BY " + strOrder[:bbBuilder.Len()-1] // hati2 rune length
		}
	}

	if bbBuilder.Len() == 0 {
		keyOrder := "ASC"
		if isPrevNav {
			keyOrder = "DESC"
		}
		baseQuery += " ORDER BY " + keyCol + " " + keyOrder
	}

	baseQuery += " LIMIT " + strconv.FormatUint(uint64(finderOptionCursor.PageSize+1), 10) + " OFFSET 0"

	return baseQuery, args, nil
}
