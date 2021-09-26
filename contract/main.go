package contract

import (
	"context"
)

type GateErrCode int

const (
	GateErrUnauthorized = GateErrCode(iota)
	GateErrInactive
	GateErrForbidden
)

type GateError interface {
	error
	ErrCode() GateErrCode
}

type Gate interface {
	RequireKey(ctx context.Context) (context.Context, error)
	Forbidden(message string) error
}

type UserMessageError interface {
	error
	Unwrap() error
}

type FieldsLabel interface {
	GetLabel(structField string) string
}

type FieldsMap struct {
	KeyAttr string
	Fields  map[string]string
}