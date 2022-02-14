package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	WorkerLimitPerQueue int `env:"WORKER_LIMIT_PER_QUEUE"`
}

func (c *Config) init(envPath string) {
	err := godotenv.Load(envPath)
	if err != nil {
		panic(err)
	}

	err = env.Parse(c)
	if err != nil {
		panic(err)
	}
}

func GetConfig(envPath string) Config {
	cfg := &Config{}
	cfg.init(envPath)

	return *cfg
}
