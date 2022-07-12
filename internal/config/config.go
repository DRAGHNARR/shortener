package config

import (
	"flag"
	"github.com/caarlos0/env"
)

type Config struct {
	Addr    string `env:"SERVER_ADDRESS" envDefault:"http://localhost:8080"`
	BaseURI string `env:"BASE_URL" envDefault:"localhost:8080"`
	Store   string `env:"FILE_STORAGE_PATH" envDefault:"test.json"`
	DSN     string `env:"DATABASE_DSN"`
}

func New() (*Config, error) {
	c := &Config{}
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	flag.StringVar(&c.Addr, "a", c.Addr, "host")
	flag.StringVar(&c.Store, "f", c.Store, "data storage")
	flag.StringVar(&c.BaseURI, "b", c.BaseURI, "base part of url")
	flag.StringVar(&c.DSN, "d", c.DSN, "db dsn")
	flag.Parse()

	return c, nil
}
