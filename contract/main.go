package contract

type UserMessageError interface {
	error
	RealError() error
}

type FieldLabel interface {
	GetLabel(structField string) string
}

type StructMap struct {
	KeyField string
	Attrs  map[string]string
}