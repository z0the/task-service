package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var cfg *Config

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	cfg = &Config{}
	err = env.Parse(cfg)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	WorkerLimitPerQueue int `env:"WORKER_LIMIT_PER_QUEUE"`
}

func GetConfig() Config {
	return *cfg
}
