package main

import (
	"a1in-bot-v3/infrastructure"
	"a1in-bot-v3/infrastructure/env"
	"a1in-bot-v3/infrastructure/qq"
	"a1in-bot-v3/infrastructure/redis"
	"a1in-bot-v3/log"
	"a1in-bot-v3/modules"
	"a1in-bot-v3/modules/chat"
	"a1in-bot-v3/modules/filewatcher"
	"a1in-bot-v3/modules/magicpen"
	"a1in-bot-v3/modules/mikan"
	"a1in-bot-v3/modules/smsff"
	"a1in-bot-v3/utils"
	"a1in-bot-v3/utils/avalon"
	"a1in-bot-v3/utils/bus"
	"flag"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "a1in-bot-v3"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "../configs/bot.toml", "config path, eg: -conf bot.toml")
	flag.Parse()
}

func main() {
	if _, err := os.Stat(flagconf); err != nil {
		panic(err)
	}
	cbs, err := os.ReadFile(flagconf)
	if err != nil {
		panic(err)
	}

	err = log.InitLogger(cbs)
	if err != nil {
		panic(err)
	}

	eg := &errgroup.Group{}

	// Init Infrastructure
	infras := []infrastructure.Infrastructure{&env.Env{}, &qq.QQInfra{}, &redis.RedisInfra{}}
	for _, infra := range infras {
		infra := infra
		eg.Go(func() error {
			return infra.InitInfra(cbs)
		})
	}
	if err = eg.Wait(); err != nil {
		panic(err)
	}
	zap.L().Info("infras init successfully.")

	// Init Util
	utils := []utils.Util{&bus.Bus{}, &avalon.Avalon{}}
	for _, util := range utils {
		util := util
		eg.Go(func() error {
			return util.InitUtil(cbs)
		})
	}
	if err = eg.Wait(); err != nil {
		panic(err)
	}
	zap.L().Info("utils init successfully.")

	// Init Module
	modules := []modules.Module{&smsff.SMSFF{}, &mikan.Mikan{}, &filewatcher.FileWatcher{}, &magicpen.MagicPen{}, &chat.Chat{}}
	for _, module := range modules {
		module := module
		eg.Go(func() error {
			if err := module.InitModule(cbs); err != nil {
				return err
			}
			go module.Run()
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		panic(err)
	}
	zap.L().Info("modules init successfully.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	zap.L().Info("Bot try closing...")
	for _, module := range modules {
		module := module
		eg.Go(func() error {
			return module.Cleanup()
		})
	}
	if err = eg.Wait(); err != nil {
		zap.L().Error("clean up bot modules fail", zap.Error(err))
	}
}
