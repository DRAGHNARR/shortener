package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"shortener/internal/handlers"
	"time"
)

type Handler struct {
	st   handlers.Storage
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
		st: s,
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *Handler) GetPlain(c echo.Context) error {
	if orig, ok := h.st.Get(c.Param("url")); ok {
		c.Response().Header().Set(echo.HeaderLocation, orig)
		return c.NoContent(http.StatusTemporaryRedirect)
	}

	return c.NoContent(http.StatusUnauthorized)
}

func (h *Handler) GetJSON(c echo.Context) error {
	var m handlers.Message
	if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
		return err
	}

	var a handlers.Message
	if orig, ok := h.st.Get(m.URL); ok {
		a.Result = orig
		body, err := json.Marshal(a)
		if err != nil {
			return err
		}
		return c.JSONBlob(http.StatusTemporaryRedirect, body)
	}
	return c.NoContent(http.StatusUnauthorized)
}

func (h Handler) Get(c echo.Context) error {
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

func (h *Handler) PostPlain(c echo.Context) error {
	orig, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	auth, err := c.Request().Cookie("uri-auth")
	if err != nil {
		return err
	}

	shorty, err := h.st.Push(string(orig), auth.Value)
	if err != nil {
		return err
	}

	return c.String(http.StatusCreated, fmt.Sprintf("%s/%s", h.base, shorty))
}

func (h *Handler) PostJSON(c echo.Context) error {
	var m handlers.Message
	if err := json.NewDecoder(c.Request().Body).Decode(&m); err != nil {
		return err
	}

	auth, err := c.Request().Cookie("uri-auth")
	if err != nil {
		return err
	}

	shorty, err := h.st.Push(m.URL, auth.Value)
	if err != nil {
		return err
	}

	var a handlers.Message
	a.Result = fmt.Sprintf("%s/%s", h.base, shorty)

	body, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return c.JSONBlob(http.StatusCreated, body)
}

func (h *Handler) Post(c echo.Context) error {
	defer func() {
		if err := c.Request().Body.Close(); err != nil {
			c.Logger().Error(err)
		}
	}()

	switch c.Request().Header.Get(echo.HeaderContentType) {
	case echo.MIMEApplicationJSON:
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

func (h Handler) Batch(c echo.Context) error {
	defer func() {
		if err := c.Request().Body.Close(); err != nil {
			c.Logger().Error(err)
		}
	}()
	switch c.Request().Header.Get(echo.HeaderContentType) {
	case echo.MIMEApplicationJSON:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		var mm []*handlers.Batch
		if err := json.NewDecoder(c.Request().Body).Decode(&mm); err != nil {
			return err
		}

		auth, err := c.Request().Cookie("uri-auth")
		if err != nil {
			return err
		}

		for _, m := range mm {
			short, err := h.st.Push(m.URI, auth.Value)
			if err != nil {
				return err
			}
			m.Short = fmt.Sprintf("%s/%s", h.base, short)
			m.URI = ""
			fmt.Println(m)
		}

		body, err := json.Marshal(mm)
		if err != nil {
			return err
		}

		return c.JSONBlob(http.StatusCreated, body)

	default:
		return c.NoContent(http.StatusUnauthorized)
	}
}

func (h *Handler) User(c echo.Context) error {
	auth, err := c.Request().Cookie("uri-auth")
	if err != nil {
		return err
	}

	if data := h.st.Users(h.base, auth.Value); len(data) != 0 {
		body, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return c.JSONBlob(http.StatusOK, body)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) Ping(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.st.Ping(ctx); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
