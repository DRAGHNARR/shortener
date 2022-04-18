package zippo

import (
	"github.com/labstack/echo/v4"
)

func DelContentLength() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Del(echo.HeaderContentLength)
			return next(c)
		}
	}
}
