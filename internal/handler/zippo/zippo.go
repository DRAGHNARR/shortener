package zippo

import (
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
)

type zippo struct {
	echo.Context
	Writer io.Writer
}

func (z *zippo) Write(b []byte) (int, error) {
	return z.Writer.Write(b)
}

func ZippoWriter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Println(c.Request().Header.Get(echo.HeaderAcceptEncoding))
			fmt.Println(c.Response().Header().Get(echo.HeaderAcceptEncoding))
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
