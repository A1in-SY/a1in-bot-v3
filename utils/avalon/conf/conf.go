package conf

import (
	"errors"
)

type config struct {
	AvalonConf *AvalonConfig
}

type AvalonConfig struct {
}

func (c *config) Check() (err error) {
	if c == nil || c.AvalonConf == nil {
		err = errors.New("avalon util conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		AvalonConf: &AvalonConfig{},
	}
}
