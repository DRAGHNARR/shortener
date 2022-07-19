package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"shortener/internal/storage"
)

type Getty interface {
	Get(string) (*storage.URIsItem, bool)
}

func Shorty(st Getty, uri string) (string, error) {
	hash := md5.Sum([]byte(uri))
	toShort := hex.EncodeToString(hash[:])

	for i := 0; i < len(toShort)-7; i++ {
		short := toShort[i : i+7]
		if suri, ok := st.Get(short); !ok || suri.URI == uri {
			return short, nil
		}
	}

	return "", fmt.Errorf("cannot shortify uri %s", uri)
}
