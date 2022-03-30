package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
)

const Host = "localhost:8080"

func Shotifier(st *sync.Map, input string) (string, error) {
	hash := md5.Sum([]byte(input))
	stringified := hex.EncodeToString(hash[:])

	for i := 0; i < len(stringified)-7; i++ {
		short := stringified[i : i+7]

		orig, ok := st.Load(short)

		switch {
		case !ok:
			st.Store(short, input)
			return short, nil
		case orig == input:
			return short, nil
		default:
			continue
		}
	}

	return "", fmt.Errorf("cannot shortify URL %s", input)
}
