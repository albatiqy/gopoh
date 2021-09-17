package id

import (
	"crypto/rand"
	"fmt"
)

// 6ecd77df6feeb63f7cc3167de86909650c2d3ff1
func RandomString(length int) string {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", key)
}