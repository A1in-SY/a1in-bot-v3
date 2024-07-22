package env

import (
	"a1in-bot-v3/infrastructure/env/conf"
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var env *Env

type Env struct {
	conf           *conf.EnvConfig
	GroupId        int64
	StabilityAPISk string
}

func (e *Env) InitInfra(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init env infra, an error happen: %v", err.Error())
		}
	}()
	c := conf.DefaultConf()
	err = toml.Unmarshal(cbs, c)
	if err != nil {
		return
	}
	err = c.Check()
	if err != nil {
		return
	}
	e.conf = c.EnvConf
	zap.L().Debug("[infra][env] init with conf", zap.Any("", *e.conf))
	gEnv := os.Getenv(e.conf.GroupIdEnv)
	if gEnv == "" {
		err = fmt.Errorf("can't find %v in system env, plz check", e.conf.GroupIdEnv)
		return
	}
	e.GroupId, err = strconv.ParseInt(gEnv, 10, 64)
	if err != nil {
		return
	}
	sk := os.Getenv(e.conf.StabilityAPISk)
	if sk == "" {
		err = fmt.Errorf("can't find %v in system env, plz check", e.conf.StabilityAPISk)
		return
	}
	e.StabilityAPISk = sk
	env = e
	zap.L().Info("[infra][env] init successfully")
	return
}

func GetGroupId() int64 {
	return env.GroupId
}

func GetStabilityAPISk() string {
	return env.StabilityAPISk
}
