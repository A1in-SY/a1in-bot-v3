package conf

import (
	"errors"
	"time"
)

type config struct {
	ChatConf *ChatConfig
}

type ChatConfig struct {
	APIKey    string
	UseProxy  bool
	ProxyAddr string
	Timeout   time.Duration
}

func (c *config) Check() (err error) {
	if c == nil || c.ChatConf == nil {
		err = errors.New("chat module conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		ChatConf: &ChatConfig{
			APIKey:    "",
			UseProxy:  false,
			ProxyAddr: "http://127.0.0.1:7890",
			Timeout:   15 * time.Second,
		},
	}
}
