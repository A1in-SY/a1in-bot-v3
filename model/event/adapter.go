package event

import (
	"fmt"
)

func (e *QQEvent) Adapt() (se *Event, err error) {
	switch e.GetPostType() {
	case PostTypeMessage:
		return e.adaptMessageEvent()
	default:
		return nil, fmt.Errorf("unknown post type: %v", e.GetPostType())
	}
}

func (e *QQEvent) adaptMessageEvent() (se *Event, err error) {
	switch e.GetMessageType() {
	case MessageTypeGroup:
		return e.adaptGroupMessageEvent()
	default:
		return nil, fmt.Errorf("unknown message type: %v", e.GetMessageType())
	}
}

func (e *QQEvent) adaptGroupMessageEvent() (se *Event, err error) {
	data := &Event_GroupMsg{
		GroupMsg: &GroupMessageEvent{
			MessageType: e.GetMessageType(),
			SubType:     e.GetSubType(),
			MessageId:   e.GetMessageId(),
			GroupId:     e.GetGroupId(),
			UserId:      e.GetUserId(),
			Anonymous:   nil,
			Message:     e.GetMessage(),
			RawMessage:  e.GetRawMessage(),
			Font:        e.GetFont(),
			Sender: &GroupMessageEvent_Sender{
				UserId:   e.GetSender().GetUserId(),
				Nickname: e.GetSender().GetNickname(),
				Sex:      e.GetSender().GetSex(),
				Age:      e.GetSender().GetAge(),
				Card:     e.GetSender().GetCard(),
				Area:     e.GetSender().GetArea(),
				Level:    e.GetSender().GetLevel(),
				Role:     e.GetSender().GetRole(),
				Title:    e.GetSender().GetTitle(),
			},
		}}
	if e.GetAnonymous() != nil {
		data.GroupMsg.Anonymous = &GroupMessageEvent_Anonymous{
			Id:   e.GetAnonymous().GetId(),
			Name: e.GetAnonymous().GetName(),
			Flag: e.GetAnonymous().GetFlag(),
		}
	}
	se = &Event{
		EventId:   EventId_MessageEventGroupMessage,
		Time:      e.GetTime(),
		SelfId:    e.GetSelfId(),
		PostType:  e.GetPostType(),
		EventData: data,
	}
	return
}
