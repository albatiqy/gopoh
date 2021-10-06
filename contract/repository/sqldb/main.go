package sqldb

import (
	"context"
	"database/sql"

	"github.com/albatiqy/gopoh/contract/repository"
)

var (
	DriversSpec = make(map[string]DriverSpec)
)

type DBSetting struct {
	DriverName string
	DSN        string
}

type DriverSpec interface {
	Open(dbSetting *DBSetting) *sql.DB
	BuildFinderCursorQuery(cursorID string, isPrevNav bool, baseQuery string, finderOptionCursor repository.FinderOptionCursor, whereRaws []string, colsMap *ColsMap) (string, []interface{}, error)
}

type Conn struct {
	DB         *sql.DB
	driverSpec DriverSpec
	dbSetting  *DBSetting
}

func (cn Conn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return cn.DB.PrepareContext(ctx, query)
}

func (cn Conn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return cn.DB.QueryContext(ctx,
		query,
		args...,
	)
}

func (cn Conn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return cn.DB.QueryRowContext(ctx, query, args...)
}

func (cn Conn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return cn.DB.ExecContext(ctx, query, args...)
}

func (cn Conn) Tx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := cn.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (cn Conn) DriverName() string {
	return cn.dbSetting.DriverName
}

func (cn Conn) NewFinderCursorQueryBuilder(baseQuery string, finderOptionCursor repository.FinderOptionCursor, colsMap *ColsMap) FinderCursorQueryBuilder {
	return FinderCursorQueryBuilder{
		driverSpec:         cn.driverSpec,
		baseQuery:          baseQuery,
		finderOptionCursor: finderOptionCursor,
		colsMap:            colsMap,
	}
}

type ColsMap struct {
	KeyAttr string
	Cols    map[string]string
}

func NewConn(dbSetting *DBSetting) *Conn {
	if dbSetting == nil {
		return nil
	}
	spec := DriversSpec[dbSetting.DriverName] // panic kalo blm diload
	db := spec.Open(dbSetting)

	/*
		if err := db.Ping(); err != nil {
			return nil
		}
	*/

	return &Conn{
		DB:         db,
		driverSpec: spec,
		dbSetting:  dbSetting,
	}
}
