package app

import (
	"../handlers"
	"../storage"
	"../utils"
	"log"
	"net/http"
)

const holder = "storage.csv"

func App() {
	st, err := storage.New(holder)
	if err != nil {
		log.Fatalf("err:> %s\n", err.Error())
	}
	defer func() {
		if err := st.File.Close(); err != nil {
			log.Fatalf("err:> %s\n", err.Error())
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/", handlers.New(st))

	http.ListenAndServe(utils.Host, mux)
}
