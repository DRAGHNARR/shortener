package app

import (
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"shortener/internal/storage"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"shortener/internal/handler/catcher"
	"shortener/internal/handler/shorty"
)

type config struct {
	addr string
	base string
	// port  string
	store string
}

func App() {
	c := &config{}

	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		c.addr = addr
	} else {
		flag.StringVar(&c.addr, "a", "localhost:8080", "host")
	}
	fmt.Println(c.addr)

	/*
		if port, ok := os.LookupEnv("SERVER_PORT"); ok {
			c.port = port
		} else {
			flag.StringVar(&c.port, "p", "8080", "port")
		}
	*/

	if store, ok := os.LookupEnv("TEMP_FILE"); ok {
		c.store = store
	} else {
		flag.StringVar(&c.store, "f", "test.json", "data storage")
	}
	s := storage.New(
		storage.WithFile(c.store),
	)
	defer func() {
		if s.File != nil {
			if err := s.File.Close(); err != nil {
				log.Printf("unexpected error: %s", err.Error())
			}
		}
	}()

	if base, ok := os.LookupEnv("BASE_URL"); ok {
		c.base = base
	} else {
		flag.StringVar(&c.base, "b", "http://localhost:8080", "base part of url")
	}
	h := shorty.New(
		s,
		shorty.WithBase(c.base),
	)
	flag.Parse()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: gzip.BestCompression,
		Skipper: func(c echo.Context) bool {
			return !strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), "gzip")
		},
	}))
	e.HTTPErrorHandler = catcher.New().Catch

	e.GET("/:url", h.Get)
	e.POST("/", h.Post)
	e.GET("/api/shorten", h.Get)
	e.POST("/api/shorten", h.Post)

	e.Host(c.addr)
	if err := e.Start(c.addr); err != http.ErrServerClosed {
		log.Fatalf("err> %s", err.Error())
	}
}
