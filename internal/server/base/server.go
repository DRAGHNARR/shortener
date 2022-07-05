package base

import "net/http"

func New() *http.Server {
	s := &http.Server{
		Addr: "localhost:8080",
	}
	return s
}
