package conf

import (
	"errors"
)

type config struct {
	RedisConf *RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func (c *config) Check() (err error) {
	if c == nil || c.RedisConf == nil {
		err = errors.New("redis infra conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		RedisConf: &RedisConfig{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		},
	}
}
