package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	PgPassword string `env:"POSTGRES_PASSWORD,required,unset"`
	PgUser     string `env:"POSTGRES_USER" envDefault:"postgres"`
	PgDb       string `env:"POSTGRES_DB" envDefault:"postgres"`
	PgHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port       int    `env:"POSTGRES_PORT" envDefault:"5432"`
}

var Envs = initConfig()

func initConfig() Config {

	err := godotenv.Load("configs/.env")
	if err != nil {
		logrus.WithError(err).Fatal("Error loading .env file")
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("failed to parse env")
	}

	return cfg
}
