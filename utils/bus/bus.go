package bus

import (
	"a1in-bot-v3/infrastructure/qq"
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
	"a1in-bot-v3/utils/bus/conf"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var bus *Bus

func GetBus() *Bus {
	return bus
}

type Bus struct {
	conf     *conf.BusConfig
	qq       *qq.QQInfra
	subMap   map[event.EventId][]*BusChan
	subMu    sync.Mutex
	busChArr []*BusChan
}

func (b *Bus) InitUtil(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init bus util, an error happen: %v", err.Error())
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
	b.conf = c.BusConf
	b.qq = qq.GetQQInfra()
	b.subMap = make(map[event.EventId][]*BusChan)
	go b.loop()
	bus = b
	zap.L().Info("[util][bus] init successfully")
	return
}

func (b *Bus) loop() {
	zap.L().Info("[util][bus] start loop")
	for {
		_, msgData, err := b.qq.ReadMessage()
		if err != nil {
			zap.L().Error("[util][bus] read message from qq infra fail", zap.Error(err))
			time.Sleep(b.conf.RetryInterval)
			continue
		}
		e := &event.QQEvent{}
		err = json.Unmarshal(msgData, e)
		if err != nil {
			zap.L().Error("bus util unmarshal qq event fail", zap.Error(err), zap.String("msgData", string(msgData)))
			continue
		}
		b.pub(e)
	}
}

func (b *Bus) pub(e *event.QQEvent) {
	se, err := e.Adapt()
	if err != nil {
		zap.L().Error("[util][bus] build standard event fail", zap.Error(err))
		return
	}
	b.subMu.Lock()
	for _, ch := range b.subMap[se.GetEventId()] {
		if !ch.IsClose() {
			ch.write(se)
		}
	}
	b.subMu.Unlock()
}

func (b *Bus) GenBusChan(subList ...event.EventId) *BusChan {
	ch := newBusChan(b)
	b.subMu.Lock()
	for _, id := range subList {
		switch id {
		case event.EventId_MessageEventAll:
			b.subMap[event.EventId_MessageEventPrivateMessage] = append(b.subMap[event.EventId_MessageEventPrivateMessage], ch)
			b.subMap[event.EventId_MessageEventGroupMessage] = append(b.subMap[event.EventId_MessageEventGroupMessage], ch)
		default:
			b.subMap[id] = append(b.subMap[id], ch)
		}
	}
	b.subMu.Unlock()
	b.busChArr = append(b.busChArr, ch)
	return ch
}

// TODO
func (b *Bus) send(msg *api.APIRequest) {
	data, _ := json.Marshal(msg)
	err := b.qq.WriteMessage(data)
	if err != nil {
		zap.L().Error("[util][bus] write message fail", zap.Error(err))
	}
}
