package sqlserver

import (
	"database/sql/driver"
	"encoding/json"
	mssqlDriver "github.com/denisenkom/go-mssqldb"
)

type NullUniqueIdentifier struct {
	UniqueIdentifier mssqlDriver.UniqueIdentifier
	Valid            bool
}

func (t *NullUniqueIdentifier) Scan(value interface{}) error {
	if value == nil {
		t.Valid = false
		return nil
	}
	tmp := mssqlDriver.UniqueIdentifier{}
	err := tmp.Scan(value)
	if err != nil {
		return err
	}

	t.Valid = true
	t.UniqueIdentifier = tmp // value.(mssql.UniqueIdentifier)

	return nil
}

func (t NullUniqueIdentifier) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.UniqueIdentifier, nil
}

func (t NullUniqueIdentifier) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return json.Marshal(t.UniqueIdentifier)
	}
	return json.Marshal(nil)
}

type UniqueIdentifier struct {
	mssqlDriver.UniqueIdentifier
}

func (t UniqueIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}