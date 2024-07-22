package bus

import (
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
)

type BusChan struct {
	ch      chan *event.Event
	isClose bool
	bus     *Bus
}

func newBusChan(bus *Bus) *BusChan {
	return &BusChan{
		ch:      make(chan *event.Event, 1024),
		isClose: false,
		bus:     bus,
	}
}

func (ch *BusChan) Close() {
	ch.isClose = true
}

func (ch *BusChan) IsClose() bool {
	return ch.isClose
}

func (ch *BusChan) Read() *event.Event {
	return <-ch.ch
}

func (ch *BusChan) write(event *event.Event) {
	ch.ch <- event
}

func (ch *BusChan) Send(msg *api.APIRequest) {
	ch.bus.send(msg)
}
