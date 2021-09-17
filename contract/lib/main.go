package lib

type IDGenUint64 interface {
	Next() (uint64, error)
}
