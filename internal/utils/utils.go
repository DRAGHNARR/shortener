package utils

import (
	"../storage"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const Host = "localhost:8080"

func Shotifier(st *storage.Storage, input string) (string, error) {
	hash := md5.Sum([]byte(input))
	stringified := hex.EncodeToString(hash[:])

	for i := 0; i < len(stringified)-7; i++ {
		short := stringified[i : i+7]

		orig, ok := st.Map[short]

		switch {
		case !ok:
			st.Append(short, input)
			return short, nil
		case orig == input:
			return short, nil
		default:
			continue
		}
	}

	return "", fmt.Errorf("Cannot shortify URL %s", input)
}
