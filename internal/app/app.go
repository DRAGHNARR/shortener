package app

import (
	"net/http"
	"shortener/internal/handlers"
	"shortener/internal/utils"
	"sync"
)

func App() {
	st := sync.Map{}

	mux := http.NewServeMux()
	mux.Handle("/", handlers.New(&st))

	http.ListenAndServe(utils.Host, mux)
}
