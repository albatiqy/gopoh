package mysql

import (
	"fmt"
	"database/sql/driver"
)

type Uint8Bool bool

func (t *Uint8Bool) Scan(v interface{}) error { // harus prepared statement
	d, ok := v.(int64)
	if !ok {
		return fmt.Errorf("unable to scan value from database")
	}
	*t = d > 0
	return nil
}

func (t Uint8Bool) Value() (driver.Value, error) {
	if t {
		return int64(1), nil
	}

	return int64(0), nil
}