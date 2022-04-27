package shorty

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"shortener/internal/storage"
)

type message struct {
	URL    string `json:"url,omitempty"`
	Result string `json:"result,omitempty"`
}

type option func(h *Shorty) error

type Shorty struct {
	storage *storage.Storage
	db      *sql.DB
	base    string
}

func New(s *storage.Storage, opts ...option) *Shorty {
	h := &Shorty{
		storage: s,
		base:    "http://localhost:8080",
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			log.Printf("warn>: %s\n", err.Error())
		}
	}

	return h
}

func WithBase(base string) option {
	return func(h *Shorty) error {
		h.base = base
		return nil
	}
}

func WithDBase(db *sql.DB) option {
	return func(h *Shorty) error {
		h.db = db
		return nil
	}
}

func (h *Shorty) GetPlain(c echo.Context) error {
	if orig, ok := h.storage.Get(c.Param("url")); ok {
		c.Response().Header().Set(echo.HeaderLocation, orig)
		return c.NoContent(http.StatusTemporaryRedirect)
	}

	return c.NoContent(http.StatusUnauthorized)
}

func (h *Shorty) GetJSON(c echo.Context) error {
	var m message
	if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
		return err
	}

	var a message
	if orig, ok := h.storage.Get(m.URL); ok {
		a.Result = orig
		body, err := json.Marshal(a)
		if err != nil {
			return err
		}
		return c.JSONBlob(http.StatusTemporaryRedirect, body)
	}
	return c.NoContent(http.StatusUnauthorized)
}

func (h *Shorty) PostPlain(c echo.Context) error {
	orig, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	auth, err := c.Request().Cookie("url-auth")
	if err != nil {
		return err
	}

	shorty, err := h.storage.Append(string(orig), auth.Value)
	if err != nil {
		return err
	}

	return c.String(http.StatusCreated, fmt.Sprintf("%s/%s", h.base, shorty))
}

func (h *Shorty) PostJSON(c echo.Context) error {
	var m message
	if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
		return err
	}

	auth, err := c.Request().Cookie("url-auth")
	if err != nil {
		return err
	}

	shorty, err := h.storage.Append(m.URL, auth.Value)
	if err != nil {
		return err
	}

	var a message
	a.Result = fmt.Sprintf("%s/%s", h.base, shorty)

	body, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return c.JSONBlob(http.StatusCreated, body)
}

func (h *Shorty) Get(c echo.Context) error {
	defer func() {
		if err := c.Request().Body.Close(); err != nil {
			c.Logger().Error(err)
		}
	}()

	switch c.Request().Header.Get(echo.HeaderContentType) {
	case echo.MIMEApplicationJSON:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return h.GetJSON(c)

	case echo.MIMETextPlainCharsetUTF8:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		return h.GetPlain(c)

	case echo.MIMETextPlain:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		return h.GetPlain(c)

	default:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		return h.GetPlain(c)
	}
}

func (h *Shorty) DBPing(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (h *Shorty) GetByUser(c echo.Context) error {
	auth, err := c.Request().Cookie("url-auth")
	if err != nil {
		return err
	}

	if data := h.storage.GetByUser(auth.Value); len(data) != 0 {
		body, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return c.JSONBlob(http.StatusOK, body)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Shorty) Post(c echo.Context) error {
	defer func() {
		if err := c.Request().Body.Close(); err != nil {
			c.Logger().Error(err)
		}
	}()

	switch c.Request().Header.Get(echo.HeaderContentType) {
	case echo.MIMEApplicationJSON:
		fmt.Println(123)
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return h.PostJSON(c)

	case echo.MIMETextPlainCharsetUTF8:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		return h.PostPlain(c)

	case echo.MIMETextPlain:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		return h.PostPlain(c)

	default:
		// return c.NoContent(http.StatusUnauthorized)
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)
		return h.PostPlain(c)
	}
}
