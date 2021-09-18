package id

import (
	"fmt"
	"crypto/rand"

	// "github.com/segmentio/ksuid"
	// "github.com/rs/xid"
)

/*
// 1sHUttQAiqFD2bv0q6TPpxl6lFE
func NextKSUID() string {
	id := ksuid.New()
	return id.String()
}

// c2bl0ihisk41o34lmfkg
func NextXID() string {
	id := xid.New()
	return id.String()
}
*/

// be6b758692fa91824053e7ddef4b8e6f
func NextGUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}