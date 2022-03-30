package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"shortener/internal/storage"
	"shortener/internal/utils"
	"strings"
	"testing"
)

const message = "wanted: %v, got: %v"

func TestShortHandler_Get(t *testing.T) {
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "positive #1 - exists",
			value: "15a9c59",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "http://exists.io",
			},
		},
		{
			name:  "positive #2 - not exists",
			value: "",
			want: want{
				code:     http.StatusUnauthorized,
				location: "",
			},
		},
	}

	st := storage.New()
	utils.Shotifier(st, "http://exists.io")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", test.value), nil)
			writer := httptest.NewRecorder()
			handler := http.Handler(New(st))
			handler.ServeHTTP(writer, request)
			result := writer.Result()
			defer result.Body.Close()

			if result.StatusCode != test.want.code {
				t.Errorf(message, result.StatusCode, test.want.code)
			}

			if location := result.Header.Get("Location"); location != test.want.location {
				t.Errorf(message, location, test.want.location)
			}
		})
	}
}

func TestShortHandler_Post(t *testing.T) {
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
			name:  "positive #1 - not exists",
			value: "http://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "localhost:8080/15a9c59",
			},
		},
		{
			name:  "positive #2 - exists",
			value: "http://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "localhost:8080/15a9c59",
			},
		},
	}

	st := storage.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.value))
			writer := httptest.NewRecorder()
			handler := http.Handler(New(st))
			handler.ServeHTTP(writer, request)
			result := writer.Result()
			defer result.Body.Close()

			if body, _ := io.ReadAll(result.Body); assert.NotNil(t, body) {
				defer result.Body.Close()
				assert.Equal(t, test.want.body, string(body), message, test.want.body, string(body))
				assert.Equal(t, test.want.code, result.StatusCode, message, result.StatusCode, test.want.code)
			}
		})
	}
}

func TestShortHandler_UnexpectedHTTPMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   int
	}{
		{
			name:   "negative #1 - not exists",
			method: http.MethodHead,
			want:   http.StatusBadRequest,
		},
		{
			name:   "negative #1 - not exists",
			method: http.MethodPatch,
			want:   http.StatusBadRequest,
		},
	}

	st := storage.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", nil)
			writer := httptest.NewRecorder()
			handler := http.Handler(New(st))
			handler.ServeHTTP(writer, request)
			result := writer.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want, result.StatusCode, message, result.StatusCode, test.want)
		})
	}
}
