package shorty

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"shortener/internal/utils"
)

type shorty struct {
	box *sync.Map
}

func New() *shorty {
	box := new(sync.Map)

	return &shorty{
		box: box,
	}
}

func (s *shorty) Get(c echo.Context) error {
	if orig, ok := s.box.Load(c.Param("url")); ok {
		c.Response().Header().Set(echo.HeaderLocation, orig.(string))
		return c.NoContent(http.StatusTemporaryRedirect)
	}

	return c.NoContent(http.StatusUnauthorized)
}

func (s *shorty) Post(c echo.Context) error {
	defer func() {
		if err := c.Request().Body.Close(); err != nil {
			c.Logger().Error(err)
		}
	}()
	orig, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	short, err := utils.Shotifier(s.box, string(orig))
	if err != nil {
		return err
	}

	return c.String(http.StatusCreated, fmt.Sprintf("http://%s/%s", utils.Host, short))
}
