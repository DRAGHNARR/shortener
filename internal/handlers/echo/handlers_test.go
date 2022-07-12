package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"shortener/internal/storage/base"
)

func TestHandlers_GetPlain(t *testing.T) {
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
			name:  "getPlain#unit#1: URL exists",
			value: "8b8d84b",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://exists.io",
			},
		},
		{
			name:  "getPlain#unit#2: URL not exists",
			value: "1234567",
			want: want{
				code:     http.StatusUnauthorized,
				location: "",
			},
		},
	}

	st := base.New(
		base.WithFile("test.json"),
	)

	_, err := st.Push("https://exists.io", "test")
	assert.NoError(t, err, "cannot append url to storage")

	h, _ := New(
		st,
		WithBase("http://localhost:8080"),
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
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

	assert.NoError(t, st.File.Close(), "unexpected error")
	assert.NoError(t, os.Remove("test.json"), "unexpected error")
}

func TestShorty_GetJson(t *testing.T) {
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name string
		body string
		want
	}{
		{
			name: "getJson#unit#1: URL exists",
			body: `{"url":"8b8d84b"}`,
			want: want{
				code: http.StatusTemporaryRedirect,
				body: `{"result":"https://exists.io"}`,
			},
		},
		{
			name: "getJson#unit#2: URL not exists",
			body: `{"url":"1234567"}`,
			want: want{
				code: http.StatusUnauthorized,
				body: "",
			},
		},
	}

	st := base.New(
		base.WithFile("test.json"),
	)

	_, err := st.Push("https://exists.io", "test")
	assert.NoError(t, err, "cannot append url to storage")

	h, _ := New(
		st,
		WithBase("http://localhost:8080"),
	)

	cookie := new(http.Cookie)
	cookie.Name = "uri-auth"
	cookie.Value = "test"
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = "/"
	cookie.Secure = false

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(test.body))
			req.AddCookie(cookie)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/shorten")

			if assert.NoError(t, h.Get(c)) {
				assert.Equal(t, test.want.code, rec.Code)
				assert.Equal(t, test.want.body, rec.Body.String())
			}
		})
	}

	assert.NoError(t, st.File.Close(), "unexpected error")
	assert.NoError(t, os.Remove("test.json"), "unexpected error")
}

func TestShorty_PostPlain(t *testing.T) {
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name string
		body string
		want
	}{
		{
			name: "postPlain#unit#1: URL not exists",
			body: "https://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/8b8d84b",
			},
		},
		{
			name: "postPlain#unit#2: URL exists",
			body: "https://exists.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/8b8d84b",
			},
		},
		{
			name: "postPlain#unit#3: URL is big and nicely",
			body: "https://AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/afbfbd3",
			},
		},
		{
			name: "postPlain#unit#4: URL is short and perfect",
			body: "https://b.io",
			want: want{
				code: http.StatusCreated,
				body: "http://localhost:8080/54b4951",
			},
		},
	}

	st := base.New(
		base.WithFile("test.json"),
	)

	_, err := st.Push("https://exists.io", "test")
	assert.NoError(t, err, "cannot append url to storage")

	h, _ := New(
		st,
		WithBase("http://localhost:8080"),
	)

	cookie := new(http.Cookie)
	cookie.Name = "uri-auth"
	cookie.Value = "test"
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = "/"
	cookie.Secure = false

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req.AddCookie(cookie)
			req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if assert.NoError(t, h.Post(c)) {
				assert.Equal(t, test.want.code, rec.Code)
				assert.Equal(t, test.want.body, rec.Body.String())
			}
		})
	}

	assert.NoError(t, st.File.Close(), "unexpected error")
	assert.NoError(t, os.Remove("test.json"), "unexpected error")
}

func TestShorty_PostJson(t *testing.T) {
	type want struct {
		code int
		body string
	}
	tests := []struct {
		name string
		body string
		want
	}{
		{
			name: "postPlain#unit#1: URL not exists",
			body: `{"url":"https://exists.io"}`,
			want: want{
				code: http.StatusCreated,
				body: `{"result":"http://localhost:8080/8b8d84b"}`,
			},
		},
		{
			name: "postPlain#unit#2: URL exists",
			body: `{"url":"https://exists.io"}`,
			want: want{
				code: http.StatusCreated,
				body: `{"result":"http://localhost:8080/8b8d84b"}`,
			},
		},
		{
			name: "postPlain#unit#3: URL is big and nicely",
			body: `{"url":"https://AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.io"}`,
			want: want{
				code: http.StatusCreated,
				body: `{"result":"http://localhost:8080/afbfbd3"}`,
			},
		},
		{
			name: "postPlain#unit#4: URL is short and perfect",
			body: `{"url":"https://b.io"}`,
			want: want{
				code: http.StatusCreated,
				body: `{"result":"http://localhost:8080/54b4951"}`,
			},
		},
	}

	st := base.New(
		base.WithFile("test.json"),
	)

	_, err := st.Push("https://exists.io", "test")
	assert.NoError(t, err, "cannot append url to storage")

	h, _ := New(
		st,
		WithBase("http://localhost:8080"),
	)

	cookie := new(http.Cookie)
	cookie.Name = "uri-auth"
	cookie.Value = "test"
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = "/"
	cookie.Secure = false

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req.AddCookie(cookie)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/shorten")

			if assert.NoError(t, h.Post(c)) {
				assert.Equal(t, test.want.code, rec.Code)
				assert.Equal(t, test.want.body, rec.Body.String())
			}
		})
	}

	assert.NoError(t, st.File.Close(), "unexpected error")
	assert.NoError(t, os.Remove("test.json"), "unexpected error")
}
