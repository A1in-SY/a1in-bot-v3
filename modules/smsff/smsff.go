package smsff

import (
	"a1in-bot-v3/infrastructure/env"
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/segment"
	"a1in-bot-v3/modules/smsff/conf"
	"a1in-bot-v3/utils/bus"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SMSFF struct {
	conf *conf.SMSFFConfig
	srv  *http.Server
	bus  *bus.BusChan
	ls   net.Listener
}

func (sms *SMSFF) InitModule(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init smsff module, an error happen: %v", err.Error())
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
	listener, err := net.Listen("tcp", c.SMSFFConf.ListenAddr)
	if err != nil {
		return
	}
	router := gin.New()
	gin.DisableConsoleColor()
	router.Use(gin.Recovery(), GinLogger(zap.L()))
	router.POST("/sendsms", sms.handler)
	sms.conf = c.SMSFFConf
	sms.bus = bus.GetBus().GenBusChan()
	sms.srv = &http.Server{
		Addr:    sms.conf.ListenAddr,
		Handler: router,
	}
	sms.ls = listener
	zap.L().Info("[module][smsff] init successfully")
	return
}

func (sms *SMSFF) Run() {
	zap.L().Sugar().Infof("[module][smsff] start serve on %v", sms.conf.ListenAddr)
	err := sms.srv.Serve(sms.ls)
	if err != nil && err != http.ErrServerClosed {
		zap.L().Error("[module][smsff] serve fail", zap.Error(err))
	}
}

func (sms *SMSFF) Cleanup() (err error) {
	return sms.srv.Shutdown(context.Background())
}

type smsData struct {
	Text string `json:"text"`
}

func (sms *SMSFF) handler(c *gin.Context) {
	var data smsData
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		zap.L().Error("[module][smsff] decode post data fail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	msg := api.BuildSendGroupMsgRequest("", env.GetGroupId(), segment.BuildTextSegment(data.Text))
	sms.bus.Send(msg)
	zap.L().Info("[module][smsff] send msg success")
	c.JSON(http.StatusOK, nil)
}

func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info("[module][smsff] recv http request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}
