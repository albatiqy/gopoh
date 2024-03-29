package mysql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/albatiqy/gopoh/contract/log"
	"github.com/albatiqy/gopoh/contract/repository"
	"github.com/albatiqy/gopoh/contract/repository/sqldb"
)

type DriverSpec struct {
}

func init() {
	sqldb.DriversSpec["mysql"] = DriverSpec{}
	log.Debugf("sqldb: %s", "mysql driver initialized")
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

// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

func (spec DriverSpec) Open(dbSetting *sqldb.DBSetting) *sql.DB {

	//DSN: "user:password@/database?parseTime=true&loc=UTC"

	db, err := sql.Open("mysql", dbSetting.DSN)
	if err != nil {
		log.Debugf(`sqldb: %s`, err)
		return nil
	}

	return db
}

func (spec DriverSpec) BuildFinderCursorQuery(cursorID string, isPrevNav bool, baseQuery string, finderOptionCursor repository.FinderOptionCursor, whereRaws []string, colsMap *sqldb.ColsMap) (string, []interface{}, error) {
	var args []interface{}
	var bbBuilder strings.Builder
	if len(finderOptionCursor.Filters) > 0 {
		for _, qFilter := range finderOptionCursor.Filters {
			if fType, ok := filterType[qFilter.Type]; ok {
				if fCol, ok := colsMap.Cols[qFilter.Attr]; ok {
					if fType == "like" { // atomic ops !! =================================
						bbBuilder.WriteString("(" + fCol + " LIKE ?) AND")
						args = append(args, "%"+qFilter.Val+"%")
					} else {
						bbBuilder.WriteString("(" + fCol + fType + "?) AND")
						args = append(args, qFilter.Val)
					}
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
			whereRaws = append(whereRaws, keyCol+" < ?")
		} else {
			whereRaws = append(whereRaws, keyCol+" > ?")
		}
		args = append(args, cursorID)
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

	baseQuery += " LIMIT 0," + strconv.FormatUint(uint64(finderOptionCursor.PageSize+1), 10)

	return baseQuery, args, nil
}
