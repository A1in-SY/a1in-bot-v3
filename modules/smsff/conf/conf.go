package conf

import (
	"errors"
)

type config struct {
	SMSFFConf *SMSFFConfig
}

type SMSFFConfig struct {
	ListenAddr string
}

func (c *config) Check() (err error) {
	if c == nil || c.SMSFFConf == nil {
		err = errors.New("smsff module conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		SMSFFConf: &SMSFFConfig{
			ListenAddr: "0.0.0.0:12345",
		},
	}
}
