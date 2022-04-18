package zippo

import (
	"github.com/labstack/echo/v4"
	"strings"
)

func DelContentLength() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), "gzip") {
				c.Response().Header().Set(echo.HeaderContentEncoding, "gzip")
				c.Response().Header().Set(echo.HeaderVary, "Accept-Encoding")
				c.Response().Header().Del(echo.HeaderContentLength)
			}
			return next(c)
		}
	}
}
