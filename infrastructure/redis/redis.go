package redis

import (
	"a1in-bot-v3/infrastructure/redis/conf"
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var redisInfra *RedisInfra

func GetRedisInfra() *RedisInfra {
	return redisInfra
}

type RedisInfra struct {
	conf *conf.RedisConfig
	rdb  *redis.Client
}

func (r *RedisInfra) InitInfra(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init redis infra, an error happen: %v", err.Error())
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
	r.conf = c.RedisConf
	zap.L().Debug("[infra][redis] init with conf", zap.Any("", *r.conf))
	r.rdb = redis.NewClient(&redis.Options{
		Addr:     r.conf.Addr,
		Password: r.conf.Password,
		DB:       r.conf.DB,
	})
	redisInfra = r
	zap.L().Info("[infra][redis] init successfully")
	return
}

func (r *RedisInfra) Incr(key string) (res int64, err error) {
	return r.rdb.Incr(context.Background(), key).Result()
}

func (r *RedisInfra) Get(key string) (res string, err error) {
	return r.rdb.Get(context.Background(), key).Result()
}

func (r *RedisInfra) Set(key string, value interface{}, expiration time.Duration) (res string, err error) {
	return r.rdb.Set(context.Background(), key, value, expiration).Result()
}

func (r *RedisInfra) IsExist(key string) (res bool, err error) {
	ires, err := r.rdb.Exists(context.Background(), key).Result()
	if ires == 1 {
		return true, err
	}
	return false, err
}

func (r *RedisInfra) SAdd(key string, member interface{}) (res int64, err error) {
	return r.rdb.SAdd(context.Background(), key, member).Result()
}

func (r *RedisInfra) SMembers(key string) (res []string, err error) {
	return r.rdb.SMembers(context.Background(), key).Result()
}
