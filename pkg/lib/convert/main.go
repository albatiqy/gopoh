package convert

import (
	"strconv"
)

func String2Int64(input string) int64 {
	if input != "" {
		if val, err := strconv.ParseInt(input, 10, 64); err == nil {
			return val
		}
	}
	return 0
}