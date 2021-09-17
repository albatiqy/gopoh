package lib

import (
	"github.com/sony/sonyflake"
)

type sonyflakeGen struct {
	obj *sonyflake.Sonyflake
}

// 353998630367055873
func (gen sonyflakeGen) Next() (uint64, error) {
	id, err := gen.obj.NextID()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NewSonyflakeIDGen() IDGenUint64 {
	return sonyflakeGen{
		obj: sonyflake.NewSonyflake(sonyflake.Settings{}),
	}
}
