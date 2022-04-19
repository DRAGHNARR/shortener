package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const Host = "localhost:8080"

func Shotifier(box *map[string]string, orig string) (string, bool, error) {
	hash := md5.Sum([]byte(orig))
	stringified := hex.EncodeToString(hash[:])

	for i := 0; i < len(stringified)-7; i++ {
		shorty := stringified[i : i+7]

		sOrig, ok := (*box)[shorty]

		switch {
		case !ok:
			(*box)[shorty] = orig
			return shorty, true, nil
		case sOrig == orig:
			return shorty, false, nil
		default:
			continue
		}
	}

	return "", false, fmt.Errorf("cannot shortify URL %s", orig)
}
