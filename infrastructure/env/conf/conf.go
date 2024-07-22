package conf

import (
	"errors"
)

type config struct {
	EnvConf *EnvConfig
}

type EnvConfig struct {
	GroupIdEnv     string
	StabilityAPISk string
}

func (c *config) Check() (err error) {
	if c == nil || c.EnvConf == nil {
		err = errors.New("env infra conf check error")
	}
	return
}

func DefaultConf() *config {
	return &config{
		EnvConf: &EnvConfig{
			GroupIdEnv:     "BOT_GROUP_ID",
			StabilityAPISk: "SB_API_SK",
		},
	}
}
