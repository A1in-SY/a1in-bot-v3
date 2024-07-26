package chat

import (
	"a1in-bot-v3/infrastructure/env"
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
	"a1in-bot-v3/model/segment"
	"a1in-bot-v3/modules/chat/conf"
	"a1in-bot-v3/utils/bus"
	"a1in-bot-v3/utils/cmdparser"
	"fmt"
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

type Chat struct {
	conf    *conf.ChatConfig
	bus     *bus.BusChan
	httpCli *http.Client
}

func (chat *Chat) InitModule(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init chat module, an error happen: %v", err.Error())
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
	chat.conf = c.ChatConf
	chat.conf.APIKey = env.GetChatAPIKey()
	chat.bus = bus.GetBus().GenBusChan(event.EventId_MessageEventGroupMessage)
	if chat.conf.UseProxy {
		pu, err := url.Parse(chat.conf.ProxyAddr)
		if err != nil {
			return err
		}
		chat.httpCli = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(pu),
			},
			Timeout: chat.conf.Timeout,
		}
	} else {
		chat.httpCli = &http.Client{
			Timeout: chat.conf.Timeout,
		}
	}
	zap.L().Info("[module][chat] init successfully")
	return
}

func (chat *Chat) Run() {
	for {
		e := chat.bus.Read()
		if ok, cmd := chat.match(e); ok {
			chat.handle(e, cmd)
		}
	}
}

func (chat *Chat) Cleanup() (err error) {
	return
}

func (chat *Chat) match(e *event.Event) (isMatch bool, cmd *ChatCommand) {
	eventData, ok := e.EventData.(*event.Event_GroupMsg)
	groupId := eventData.GroupMsg.GetGroupId()
	if !ok {
		return
	}
	// at和文本
	// 文本
	if len(eventData.GroupMsg.GetMessage()) != 1 {
		return
	}
	if eventData.GroupMsg.GetMessage()[0].Type != segment.SegmentTypeText {
		return
	}
	text := strings.TrimSpace(eventData.GroupMsg.GetMessage()[0].Data.Text)
	// 文本开头为指定字符串
	if !strings.HasPrefix(text, "#chat") {
		return
	}
	cmd = &ChatCommand{}
	err := cmdparser.Parse(text, cmd)
	if err != nil {
		zap.L().Error("[module][magicpen] parse draw command fail", zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment(fmt.Sprintf("解析命令 %v 时出错: %v", text, err.Error())))
		chat.bus.Send(msg)
		return
	}
	isMatch = true
	return
}

func (chat *Chat) handle(e *event.Event, cmd *ChatCommand) {
	eventData := e.EventData.(*event.Event_GroupMsg)
	userId := eventData.GroupMsg.GetUserId()
	groupId := eventData.GroupMsg.GetGroupId()
	zap.L().Info("[module][chat] handle", zap.Int64("userId", userId), zap.Int64("groupId", groupId), zap.Any("cmd", cmd))
	if cmd.Help {
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("chat命令格式: chat content\n"+
			"支持的可选项有:\n"+
			"-help 查看帮助\n"+
			"感谢安总的支持"))
		chat.bus.Send(msg)
		return
	} else {
		err := cmd.CheckCommand()
		if err != nil {
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 命令参数不合法: %v", err.Error())))
			chat.bus.Send(msg)
			return
		}
		res, err := chat.chat(cmd)
		if err != nil {
			zap.L().Error("[module][chat] chat fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 调用API失败，错误: %v", err.Error())))
			chat.bus.Send(msg)
			return
		}
		if len(res.Choices) == 0 {
			zap.L().Error("[module][chat] empty result", zap.Int64("userId", userId), zap.Any("res", res))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 调用API返回结果为空"))
			chat.bus.Send(msg)
			return
		}
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(res.Choices[0].Message.Content))
		chat.bus.Send(msg)
		return
	}
}
