package zippo

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type zippo struct {
	http.ResponseWriter
	Writer io.Writer
}

func (z zippo) Write(b []byte) (int, error) {
	return z.Writer.Write(b)
}

func ZippoReader() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get(echo.HeaderContentEncoding) != "gzip" {
				return next(c)
			}

			gz, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return err
			}
			defer func() {
				c.Logger().Error(gz.Close())
			}()

			c.Request().Body = gz
			return next(c)
		}
	}
}

func ZippoWriter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), "gzip") {
				return next(c)
			}

			gz, err := gzip.NewWriterLevel(c.Response().Writer, gzip.BestSpeed)
			if err != nil {
				return err
			}
			defer func() {
				if err := gz.Close(); err != nil {
					c.Logger().Error(err)
				}
			}()

			c.Response().Writer = zippo{
				c.Response().Writer,
				gz,
			}
			c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")
			c.Response().Header().Set(echo.HeaderVary, echo.HeaderAcceptEncoding)
			//c.Response().Header().Set(echo.HeaderVary, echo.HeaderContentEncoding)
			c.Response().Header().Del(echo.HeaderContentLength) // wtf??? check
			return next(c)
		}
	}
}
