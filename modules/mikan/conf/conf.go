package conf

import (
	"errors"
	"time"
)

type config struct {
	MikanConf *MikanConfig
}

type MikanConfig struct {
	UseProxy    bool
	ProxyAddr   string
	Timeout     time.Duration
	RSSInterval time.Duration
	LinkPerPage int
}

func (c *config) Check() (err error) {
	if c == nil || c.MikanConf == nil {
		err = errors.New("mikan module conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		MikanConf: &MikanConfig{
			UseProxy:    true,
			ProxyAddr:   "http://127.0.0.1:7890",
			Timeout:     5 * time.Second,
			RSSInterval: 15 * time.Minute,
			LinkPerPage: 5,
		},
	}
}
