package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"shortener/internal/storage"
	"shortener/internal/utils"
	"strings"
	"testing"
)

const message = "wanted: %v, got: %v"
const holder = "./storage_test.csv"

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

	st, _ := storage.New(holder)
	defer func() {
		if err := st.File.Close(); err != nil {
			assert.Errorf(t, err, "cannot close test storage")
		}
		os.Remove(holder)
	}()
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
				code: http.StatusOK,
				body: `{"url": "localhost:8080/15a9c59"}`,
			},
		},
		{
			name:  "positive #2 - exists",
			value: "http://exists.io",
			want: want{
				code: http.StatusOK,
				body: `{"url": "localhost:8080/15a9c59"}`,
			},
		},
	}

	st, _ := storage.New(holder)
	defer func() {
		if err := st.File.Close(); err != nil {
			assert.Errorf(t, err, "cannot close test storage")
		}
		os.Remove(holder)
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(fmt.Sprintf(`{"url": "%v"}`, test.value)))
			writer := httptest.NewRecorder()
			handler := http.Handler(New(st))
			handler.ServeHTTP(writer, request)
			result := writer.Result()
			defer result.Body.Close()

			if body, _ := io.ReadAll(result.Body); assert.NotNil(t, body) {
				defer result.Body.Close()
				assert.JSONEqf(t, test.want.body, string(body), message, test.want.body, string(body))
				assert.Equal(t, result.StatusCode, http.StatusOK, message, test.want.code, http.StatusOK)
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

	st, _ := storage.New(holder)
	defer func() {
		if err := st.File.Close(); err != nil {
			assert.Errorf(t, err, "cannot close test storage")
		}
		os.Remove(holder)
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", nil)
			writer := httptest.NewRecorder()
			handler := http.Handler(New(st))
			handler.ServeHTTP(writer, request)
			result := writer.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, test.want, message, result.StatusCode, test.want)
		})
	}
}
