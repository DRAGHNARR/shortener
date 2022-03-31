package app

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"shortener/internal/handler/catcher"
	"shortener/internal/handler/shorty"
)

func App() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = catcher.New().Catch

	h := shorty.New()

	e.GET("/:url", h.Get)
	e.POST("/", h.Post)

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatalf("err> %s", err.Error())
	}
}
