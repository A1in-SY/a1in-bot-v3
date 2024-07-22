package conf

import "errors"

type config struct {
	LogConf *LogConfig
}

type LogConfig struct {
	Level    string
	FileName string
	// 单位MB
	MaxSize      int
	MaxAge       int
	MaxBackups   int
	IsStdout     bool
	IsStackTrace bool
}

func (c *config) Check() (err error) {
	if c == nil || c.LogConf == nil {
		err = errors.New("log util conf check error")
	}
	if c.LogConf.FileName == "" {
		err = errors.New("log util conf FileName is empty")
	}
	return
}

func DefaultConf() *config {
	return &config{
		LogConf: &LogConfig{
			Level:        "debug",
			FileName:     "bot.log",
			MaxSize:      200,
			MaxAge:       0,
			MaxBackups:   0,
			IsStdout:     true,
			IsStackTrace: true,
		},
	}
}
