package qq

import (
	"a1in-bot-v3/infrastructure/qq/conf"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var qqInfra *QQInfra

func GetQQInfra() *QQInfra {
	return qqInfra
}

type QQInfra struct {
	conf            *conf.QQConfig
	conn            *websocket.Conn
	isConnAvailable bool
	isReconnecting  bool
	reconnectMu     sync.Mutex
}

func (qq *QQInfra) InitInfra(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init qq infra, an error happen: %v", err.Error())
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
	qq.conf = c.QQConf
	zap.L().Debug("[infra][qq] init with conf", zap.Any("", *qq.conf))
	err = qq.initQQConn()
	if err != nil {
		return
	}
	qqInfra = qq
	zap.L().Info("[infra][qq] init successfully")
	return
}

func (qq *QQInfra) initQQConn() (err error) {
	qqWsAddr := qq.conf.WsAddr
	qqWsURL := url.URL{Scheme: "ws", Host: qqWsAddr, Path: "/"}
	header := make(http.Header)
	header.Add("Origin", fmt.Sprintf("http://%v", qqWsAddr))
	qqWsConn, _, err := websocket.DefaultDialer.Dial(qqWsURL.String(), header)
	if err != nil {
		return
	}
	qq.conn = qqWsConn
	qq.isConnAvailable = true
	return
}

func (qq *QQInfra) ReadMessage() (msgType int, msgData []byte, err error) {
	if qq.isConnAvailable {
		msgType, msgData, err = qq.conn.ReadMessage()
		if err != nil {
			zap.L().Error("qq infra read message fail", zap.Error(err))
			qq.isConnAvailable = false
			go qq.reconnect()
			return -1, nil, errors.New("qq infra is unavailable, plz try later")
		}
		// zap.L().Debug("qq infra recv message", zap.String("msg", string(msgData)))
		return
	} else {
		return -1, nil, errors.New("qq infra is unavailable, plz try later")
	}
}

func (qq *QQInfra) WriteMessage(data []byte) (err error) {
	if qq.isConnAvailable {
		err = qq.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			zap.L().Error("qq infra write message fail", zap.Error(err))
			qq.isConnAvailable = false
			go qq.reconnect()
			return errors.New("qq infra is unavailable, plz try later")
		}
	} else {
		return errors.New("qq infra is unavailable, plz try later")
	}
	return
}

func (qq *QQInfra) reconnect() {
	qq.reconnectMu.Lock()
	if qq.isReconnecting {
		zap.L().Warn("qq infra is reconnecting now")
		qq.reconnectMu.Unlock()
		return
	} else {
		zap.L().Warn("qq infra start reconnecting")
		qq.reconnectMu.Unlock()
		qq.isReconnecting = true
	}
	for {
		err := qq.initQQConn()
		if err != nil {
			zap.L().Warn("qq infra reconnect fail", zap.Error(err))
			time.Sleep(qq.conf.RetryInterval)
		} else {
			break
		}
	}
	qq.isReconnecting = false
}
