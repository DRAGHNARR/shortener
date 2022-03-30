package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"shortener/internal/storage"
	"shortener/internal/utils"
)

type ShortHandler struct {
	st storage.Storage
}

func New(st storage.Storage) ShortHandler {
	return ShortHandler{
		st: st,
	}
}

func (handler ShortHandler) Error(writer http.ResponseWriter, err error) {
	log.Printf("warn:> %s\n", err.Error())
	http.Error(writer, err.Error(), http.StatusFailedDependency) //http.StatusInternalServerError)
}

func (handler ShortHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	method := request.Method

	switch method {
	case "GET":
		handler.Get(writer, request)
	case "POST":
		handler.Post(writer, request)
	default:
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(nil)
	}
}

func (handler ShortHandler) Get(writer http.ResponseWriter, request *http.Request) {
	if original, ok := handler.st[request.URL.Path[1:]]; ok {
		writer.Header().Set("Location", original)
		writer.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		writer.WriteHeader(http.StatusUnauthorized)
	}
	writer.Write(nil)
}

func (handler ShortHandler) Post(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := request.Body.Close(); err != nil {
			log.Printf("warn> %s\n", err.Error())
		}
	}()
	original, err := io.ReadAll(request.Body)
	if err != nil {
		handler.Error(writer, err)
		return
	}
	short, err := utils.Shotifier(handler.st, string(original))
	if err != nil {
		handler.Error(writer, err)
		return
	}
	body := struct {
		URL string `json:"url"`
	}{
		URL: fmt.Sprintf("%s/%s", utils.Host, short),
	}

	jbody, err := json.Marshal(body)
	if err != nil {
		handler.Error(writer, err)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write(jbody)
}
