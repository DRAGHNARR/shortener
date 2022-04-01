package shorty

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestShorty_Get(t *testing.T) {
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
			name:  "get#unit#1: URL exists",
			value: "8b8d84b",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://exists.io",
			},
		},
		{
			name:  "get#unit#2: URL not exists",
			value: "123123123",
			want: want{
				code:     http.StatusUnauthorized,
				location: "",
			},
		},
	}

	h := New()
	h.box.Store("8b8d84b", "https://exists.io")

	fmt.Println(h.box)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:url")
			c.SetParamNames("url")
			c.SetParamValues(test.value)

			if assert.NoError(t, h.Get(c)) {
				assert.Equal(t, test.want.code, rec.Code)
				assert.Equal(t, test.want.location, rec.Header().Get(echo.HeaderLocation))
			}
		})
	}
}

func TestShorty_Post(t *testing.T) {
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
			name:  "post#unit#1: URL not exists",
			value: "https://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/8b8d84b",
			},
		},
		{
			name:  "post#unit#2: URL exists",
			value: "https://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/8b8d84b",
			},
		},
		{
			name:  "post#unit#3: URL is big and nicely",
			value: "https://AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/afbfbd3",
			},
		},
		{
			name:  "post#unit#4: URL is short and perfect",
			value: "https://b.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/54b4951",
			},
		},
	}

	h := New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.value))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if assert.NoError(t, h.Post(c)) {
				assert.Equal(t, test.want.code, rec.Code)
				assert.Equal(t, test.want.body, rec.Body.String())
			}
		})
	}
}
