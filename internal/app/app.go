package app

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	eb "shortener/internal/handlers/echo"
	stb "shortener/internal/storage/base"
	db2 "shortener/internal/storage/db"

	"shortener/internal/handlers/echo/middlewares/auth"
	"shortener/internal/handlers/echo/middlewares/catcher"
	"shortener/internal/handlers/echo/middlewares/zippo"
)

type config struct {
	addr string
	base string
	// port  string
	store string
	dsn   string
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
	fmt.Println(c.base, base)
	if !ok {
		base = "http://localhost:8080"
	}
	flag.StringVar(&c.base, "b", base, "base part of url")

	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if !ok {
		dsn = ""
	}
	flag.StringVar(&c.dsn, "d", dsn, "db dsn")
	flag.Parse()

	var h *eb.Handler
	// "postgresql://postgres:postgres@localhost?sslmode=disable"
	db, err := sql.Open("postgres", c.dsn)
	if err != nil {
		log.Println(err)
	} else {
		defer func() {
			if err := db.Close(); err != nil {
				log.Fatalln(err)
			}
		}()
		if dbst, err := db2.New(db); err != nil {
			log.Println(err)
		} else {
			if h, err = eb.New(dbst, eb.WithBase(c.base)); err != nil {
				log.Println(err)
			}
		}
	}

	if h == nil {
		bst := stb.New(stb.WithFile(c.store))
		if h, err = eb.New(bst, eb.WithBase(c.base)); err != nil {
			log.Fatalln(err)
		}
		defer func() {
			if bst.File != nil {
				if err := bst.File.Close(); err != nil {
					log.Printf("unexpected error: %s", err.Error())
				}
			}
		}()
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(auth.Check())
	e.Use(zippo.ZippoReader())
	e.Use(zippo.ZippoWriter())
	e.HTTPErrorHandler = catcher.New().Catch

	e.GET("/:url", h.Get)
	e.POST("/", h.Post)
	e.GET("/api/shorten", h.Get)
	e.POST("/api/shorten", h.Post)
	e.POST("/api/shorten/batch", h.Batch)
	e.GET("/api/user/urls", h.User)
	e.GET("/ping", h.Ping)

	if err := e.Start(c.addr); err != http.ErrServerClosed {
		log.Fatalf("err> %s", err.Error())
	}
}
