package catcher

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type catcher struct {
	code int
}

func New() *catcher {
	return &catcher{
		code: http.StatusBadRequest,
	}
}

func (cat *catcher) Catch(err error, c echo.Context) {
	c.Logger().Error(err)
	if err := c.NoContent(cat.code); err != nil {
		c.Logger().Error(err)
	}
}
