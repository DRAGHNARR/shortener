package zippo

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"io"
	"strings"
)

type zippo struct {
	echo.Context
	writer io.Writer
}

func Zippo() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), "gzip") {
				return next(c)
			}

			gz, err := gzip.NewWriterLevel(c.Response(), gzip.BestSpeed)
			if err != nil {
				return err
			}
			defer func() {
				if err := gz.Close(); err != nil {
					c.Logger().Error(err)
				}
			}()

			c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")
			z := zippo{
				c,
				gz,
			}

			return next(z)
		}
	}
}
