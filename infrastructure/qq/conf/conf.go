package conf

import (
	"errors"
	"time"
)

type config struct {
	QQConf *QQConfig
}

type QQConfig struct {
	WsAddr        string
	RetryInterval time.Duration
}

func (c *config) Check() (err error) {
	if c == nil || c.QQConf == nil {
		err = errors.New("qq infra conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		QQConf: &QQConfig{
			WsAddr:        "127.0.0.1:3001",
			RetryInterval: 1 * time.Second,
		},
	}
}
