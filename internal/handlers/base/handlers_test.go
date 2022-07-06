package base

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sb "shortener/internal/storage/base"
)

const message = "wanted: %v, got: %v"

func TestHandlerGet(t *testing.T) {
	st := sb.New()
	short, err := st.Push("http://exists.io", "test")
	if err != nil {
		log.Fatalf("err:> unexpected error (st.Push): %s", err.Error())
	}
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name  string
		value string
		want
	}{
		{
			name:  "positive#1: Exists",
			value: short,
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "http://exists.io",
			},
		},
		{
			name:  "positive#2: Not exists",
			value: "",
			want: want{
				code:     http.StatusUnauthorized,
				location: "",
			},
		},
	}

	h, err := New(st, WithBase("http://localhost:8080"))
	if err != nil {
		log.Fatalf("err:> unexpected error (h New): %s", err.Error())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", test.value), nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			func() {
				res := w.Result()
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("err:> unexpected error (res.Body.Close): %s", err.Error())
					}
				}()

				if res.StatusCode != test.want.code {
					t.Errorf(message, test.want.code, res.StatusCode)
				}

				if location := res.Header.Get("Location"); location != test.want.location {
					t.Errorf(message, test.want.location, location)
				}
			}()
		})
	}
}

func TestHandlerPost(t *testing.T) {
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		value string
		want
	}{
		{
			name:  "positive #1 - Not exists",
			value: "http://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/15a9c59",
			},
		},
		{
			name:  "positive #2 - Exists",
			value: "http://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/15a9c59",
			},
		},
	}

	st := sb.New()
	h, err := New(st, WithBase("http://localhost:8080"))
	if err != nil {
		log.Fatalf("err:> unexpected error (h New): %s", err.Error())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.value))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			func() {
				res := w.Result()
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("err:> unexpected error (res.Body.Close): %s", err.Error())
					}
				}()

				if body, _ := io.ReadAll(res.Body); assert.NotNil(t, body) {
					assert.Equal(t, test.want.body, string(body), message, test.want.body, string(body))
					assert.Equal(t, test.want.code, res.StatusCode, message, res.StatusCode, test.want.code)
				}
			}()
		})
	}
}

func TestHandlerUnexpectedMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   int
	}{
		{
			name:   "negative#1 - Not exists",
			method: http.MethodHead,
			want:   http.StatusBadRequest,
		},
		{
			name:   "negative #2 - Not exists",
			method: http.MethodPatch,
			want:   http.StatusBadRequest,
		},
	}

	st := sb.New()
	h, err := New(st, WithBase("http://localhost:8080"))
	if err != nil {
		log.Fatalf("err:> unexpected error (h New): %s", err.Error())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, "/", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			func() {
				res := w.Result()
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("err:> unexpected error (res.Body.Close): %s", err.Error())
					}
				}()

				assert.Equal(t, test.want, res.StatusCode, message, res.StatusCode, test.want)
			}()
		})
	}
}
