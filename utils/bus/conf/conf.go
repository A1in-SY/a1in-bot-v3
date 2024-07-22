package conf

import (
	"errors"
	"time"
)

type config struct {
	BusConf *BusConfig
}

type BusConfig struct {
	RetryInterval time.Duration
}

func (c *config) Check() (err error) {
	if c == nil || c.BusConf == nil {
		err = errors.New("bus util conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		BusConf: &BusConfig{
			RetryInterval: 1 * time.Second,
		},
	}
}
