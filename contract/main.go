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
	RequireAuth(ctx context.Context) error
	RequireRole(ctx context.Context, roleName string) error
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
