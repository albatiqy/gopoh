package lib

type IDGenInt64 interface {
	Next() int64
}

type IDGenUint64 interface {
	Next() uint64
}
