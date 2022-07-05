package base

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"shortener/internal/handlers"

	"github.com/labstack/echo/v4"
)

/*
	POST uri/ 201-400
	GET uri/{id} 307-400
*/

type Handler struct {
	s    handlers.Storage
	base string
}

type Options func(h *Handler) error

func WithBase(base string) Options {
	return func(h *Handler) error {
		h.base = base
		return nil
	}
}

func New(s handlers.Storage, opts ...Options) (*Handler, error) {
	h := &Handler{
		s: s,
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *Handler) Error(writer http.ResponseWriter, err error) {
	log.Printf("warn:> unexpected error: %s\n", err.Error())
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := r.Method

	switch m {
	case "GET":
		h.Get(w, r)
	case "POST":
		h.Post(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write(nil); err != nil {
			h.Error(w, err)
			return
		}
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	short := r.URL.Path[1:]
	if uri, ok := h.s.Get(short); ok {
		w.Header().Set(echo.HeaderLocation, uri)
		w.WriteHeader(http.StatusTemporaryRedirect)
		if _, err := w.Write(nil); err != nil {
			h.Error(w, err)
			return
		}
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write(nil); err != nil {
		h.Error(w, err)
		return
	}
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.Error(w, err)
		return
	}
	fmt.Println(string(b))
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("warn:> unexpected error: %s\n", err.Error())
		}
	}()

	auth, err := r.Cookie("uri-auth")
	if err != nil {
		h.Error(w, err)
		//return
	}
	short, err := h.s.Push(string(b), auth.Value)

	if err != nil {
		h.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(fmt.Sprintf("%s/%s", h.base, short))); err != nil {
		h.Error(w, err)
		return
	}
}
