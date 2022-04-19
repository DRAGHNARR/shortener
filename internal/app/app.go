package app

import (
	"flag"
	"log"
	"net/http"
	"os"
	
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"shortener/internal/handler/catcher"
	"shortener/internal/handler/shorty"
	"shortener/internal/handler/zippo"
	"shortener/internal/storage"
)

type config struct {
	addr string
	base string
	// port  string
	store string
}

func App() {
	c := &config{}

	addr, ok := os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		addr = "localhost:8080"
	}
	flag.StringVar(&c.addr, "a", addr, "host")

	store, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		store = "test.json"
	}
	flag.StringVar(&c.store, "f", store, "data storage")

	base, ok := os.LookupEnv("BASE_URL")
	if !ok {
		base = "http://localhost:8080"
	}
	flag.StringVar(&c.base, "b", base, "base part of url")
	flag.Parse()

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

	h := shorty.New(
		s,
		shorty.WithBase(c.base),
	)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(zippo.ZippoReader())
	e.Use(zippo.ZippoWriter())
	e.HTTPErrorHandler = catcher.New().Catch

	e.GET("/:url", h.Get)
	e.POST("/", h.Post)
	e.GET("/api/shorten", h.Get)
	e.POST("/api/shorten", h.Post)

	if err := e.Start(c.addr); err != http.ErrServerClosed {
		log.Fatalf("err> %s", err.Error())
	}
}
