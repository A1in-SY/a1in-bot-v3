package conf

import (
	"errors"
	"time"
)

type config struct {
	MagicPenConf *MagicPenConfig
}

type MagicPenConfig struct {
	Sk             string
	UseProxy       bool
	ProxyAddr      string
	Timeout        time.Duration
	PicStoragePath string
}

func (c *config) Check() (err error) {
	if c == nil || c.MagicPenConf == nil {
		err = errors.New("magicpen module conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		MagicPenConf: &MagicPenConfig{
			Sk:             "",
			UseProxy:       true,
			ProxyAddr:      "http://127.0.0.1:7890",
			Timeout:        15 * time.Second,
			PicStoragePath: "/home/admin/bot/draw/",
		},
	}
}
