package avalon

import (
	"a1in-bot-v3/infrastructure/redis"
	"a1in-bot-v3/utils/avalon/conf"
	"fmt"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var avalon *Avalon

func GetAvalon() *Avalon {
	return avalon
}

type Avalon struct {
	conf  *conf.AvalonConfig
	redis *redis.RedisInfra
}

func (a *Avalon) InitUtil(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init avalon util, an error happen: %v", err.Error())
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
	a.conf = c.AvalonConf
	a.redis = redis.GetRedisInfra()
	avalon = a
	zap.L().Info("[util][avalon] init successfully")
	return
}

func (a *Avalon) ReportRiskIP(ip string) {

}

func (a *Avalon) IsRiskBan(ip string) (ban bool) {
	return
}

func (a *Avalon) genRiskIPKey(ip string) (key string) {
	return fmt.Sprintf("risk_ip_%v", ip)
}
