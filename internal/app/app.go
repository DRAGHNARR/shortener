package app

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"shortener/internal/config"

	handlersInterfaces "shortener/internal/handlers"
	echoHandlers "shortener/internal/handlers/echo"
	"shortener/internal/handlers/echo/middlewares/auth"
	"shortener/internal/handlers/echo/middlewares/catcher"
	"shortener/internal/handlers/echo/middlewares/zippo"
	fileStorage "shortener/internal/storage/base"
	dbStorage "shortener/internal/storage/db"
)

func App() {
	c, err := config.New()
	fmt.Println(c.Addr)
	if err != nil {
		log.Fatalf("err:> unable to initialize config: %s\n", err.Error())
	}

	var st handlersInterfaces.Storage
	// "postgresql://postgres:postgres@localhost?sslmode=disable"

	if c.DSN != "" {
		db, err := sql.Open("postgres", c.DSN)
		if err != nil {
			log.Fatalf("err:> unable to connect with dsn %s: %s\n", c.DSN, err.Error())
		}
		st, err = dbStorage.New(db)
		if err != nil {
			log.Fatalf("err:> unable to initialize db structures: %s\n", err.Error())
		}
	} else if c.Store != "" {
		st = fileStorage.New(fileStorage.WithFile(c.Store))
	} else {
		st = fileStorage.New()
	}
	defer func() {
		if err := st.Close(); err != nil {
			log.Fatalf("err:> unexpected error on storage closing: %s\n", err.Error())
		}
	}()

	h, err := echoHandlers.New(st, echoHandlers.WithBase(c.BaseURI))
	if err != nil {
		log.Fatalf("err:> unable to initialize handlers: %s\n", err.Error())
	}
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(auth.Check())
	e.Use(zippo.ZippoReader())
	e.Use(zippo.ZippoWriter())
	e.HTTPErrorHandler = catcher.New().Catch

	e.DELETE("/api/user/urls", h.DeleteURIsByList)
	e.GET("/:url", h.Get)
	e.POST("/", h.Post)
	e.GET("/api/shorten", h.Get)
	e.POST("/api/shorten", h.Post)
	e.POST("/api/shorten/batch", h.Batch)
	e.GET("/api/user/urls", h.User)
	e.GET("/ping", h.Ping)

	if err := e.Start(c.Addr); err != http.ErrServerClosed {
		log.Fatalf("err> %s", err.Error())
	}
}
