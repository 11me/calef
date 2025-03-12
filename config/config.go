package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Nats   `envPrefix:"NATS_"`
	Server `envPrefix:"SERVER_"`
}

type Nats struct {
	URL string `env:"URL,notEmpty"`
}

type Server struct {
	Addr string `env:"ADDR" envDefault:":8080"`
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read .env: %w", err)
	}

	conf := &Config{}
	if err := env.Parse(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
