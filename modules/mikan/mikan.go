package mikan

import (
	"a1in-bot-v3/infrastructure/env"
	"a1in-bot-v3/infrastructure/redis"
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
	"a1in-bot-v3/model/segment"
	"a1in-bot-v3/modules/mikan/conf"
	"a1in-bot-v3/utils/bus"
	"a1in-bot-v3/utils/cmdparser"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type Mikan struct {
	conf    *conf.MikanConfig
	bus     *bus.BusChan
	rds     *redis.RedisInfra
	httpCli *http.Client
}

type mikanSubscription struct {
	GroupId int64
	UserId  int64
	RssUrl  string
	// 已读列表，时间戳与MikanURL
	Read map[string]int64
}

func (mikan *Mikan) InitModule(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init mikan module, an error happen: %v", err.Error())
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
	mikan.conf = c.MikanConf
	mikan.bus = bus.GetBus().GenBusChan(event.EventId_MessageEventGroupMessage)
	mikan.rds = redis.GetRedisInfra()
	if mikan.conf.UseProxy {
		pu, err := url.Parse(mikan.conf.ProxyAddr)
		if err != nil {
			return err
		}
		mikan.httpCli = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(pu),
			},
			Timeout: mikan.conf.Timeout,
		}
	} else {
		mikan.httpCli = &http.Client{
			Timeout: mikan.conf.Timeout,
		}
	}
	zap.L().Info("[module][mikan] init successfully")
	return
}

func (mikan *Mikan) Run() {
	go mikan.loopRead()
	go mikan.loopSub()
}

func (mikan *Mikan) Cleanup() (err error) {
	return
}

func (mikan *Mikan) loopRead() {
	for {
		e := mikan.bus.Read()
		if ok, cmd := mikan.match(e); ok {
			mikan.handle(e, cmd)
		}
	}
}

func (mikan *Mikan) loopSub() {
	for {
		userList, err := mikan.getMikanUserList()
		if err != nil {
			zap.L().Error("[module][mikan] get mikan user list fail", zap.Error(err))
			continue
		}
		for _, userId := range userList {
			go mikan.userSub(userId)
		}
		time.Sleep(mikan.conf.RSSInterval)
	}
}

type Command struct {
	R    string
	Out  string
	Size int
}

func (mikan *Mikan) match(e *event.Event) (isMatch bool, cmd *MikanCommand) {
	if e.GetPostType() != event.PostTypeMessage {
		return
	}
	eventData, ok := e.EventData.(*event.Event_GroupMsg)
	groupId := eventData.GroupMsg.GetGroupId()
	if !ok {
		return
	}
	// 文本
	if len(eventData.GroupMsg.GetMessage()) != 1 {
		return
	}
	if eventData.GroupMsg.GetMessage()[0].Type != segment.SegmentTypeText {
		return
	}
	text := strings.TrimSpace(eventData.GroupMsg.GetMessage()[0].Data.Text)
	// 文本开头为指定字符串
	if !strings.HasPrefix(text, "#mikan") {
		return
	}
	cmd = &MikanCommand{}
	err := cmdparser.Parse(text, cmd)
	if err != nil {
		zap.L().Error("[module][mikan] parse file command fail", zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment(fmt.Sprintf("解析命令 %v 时出错: %v", text, err.Error())))
		mikan.bus.Send(msg)
		return
	}
	isMatch = true
	return
}

func (mikan *Mikan) handle(e *event.Event, cmd *MikanCommand) {
	eventData := e.EventData.(*event.Event_GroupMsg)
	userId := eventData.GroupMsg.GetUserId()
	groupId := eventData.GroupMsg.GetGroupId()
	zap.L().Info("[module][mikan] handle", zap.Any("event", e))
	if cmd.Help {
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("mikan命令格式: mikan [options] url\n"+
			"支持的可选项有:\n"+
			"-help 查看帮助\n"+
			"-bind 设置绑定的Mikan RSS源\n"+
			"-unbind 取消之前绑定的Mikan RSS源，不需要携带参数（还没做）"))
		mikan.bus.Send(msg)
		return
	} else {
		err := cmd.CheckCommand()
		if err != nil {
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 命令参数不合法: %v", err.Error())))
			mikan.bus.Send(msg)
			return
		}
		if cmd.Operation == "bind" {
			isExist, err := mikan.isMikanUserExist(userId)
			if err != nil {
				zap.L().Error("[module][mikan] check exist mikan user fail", zap.Int64("userId", userId), zap.Error(err))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("检查历史绑定记录时出错"))
				mikan.bus.Send(msg)
				return
			}
			var sub *mikanSubscription
			var replyText string
			if isExist {
				sub, err = mikan.getMikanUserSub(userId)
				if err != nil {
					zap.L().Error("[module][mikan] get mikan user sub fail", zap.Int64("userId", userId), zap.Error(err))
					msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("检查历史绑定记录时出错"))
					mikan.bus.Send(msg)
					return
				}
				sub.GroupId = groupId
				sub.RssUrl = cmd.Param
				replyText = fmt.Sprintf(" 之前已绑定过 MikanRSS 源: %v，已更换为指定源", sub.RssUrl)
			} else {
				sub = &mikanSubscription{
					GroupId: groupId,
					UserId:  userId,
					RssUrl:  cmd.Param,
					Read: map[string]int64{
						"init": time.Now().Unix(),
					},
				}
				replyText = " 成功绑定 MikanRSS 源"
			}
			eg := &errgroup.Group{}
			eg.Go(func() error {
				return mikan.setMikanUserSub(userId, sub)
			})
			eg.Go(func() error {
				return mikan.addMikanUserList(userId)
			})
			err = eg.Wait()
			if err != nil {
				zap.L().Error("[module][mikan] set mikan user sub fail", zap.Int64("userId", userId), zap.Error(err))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("绑定 MikanRSS 源时出错"))
				mikan.bus.Send(msg)
				return
			}
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(replyText))
			mikan.bus.Send(msg)
			go func() {
				time.Sleep(3 * time.Second)
				mikan.userSub(userId)
			}()
		} else if cmd.Operation == "unbind" {
			// TODO
			zap.L().Error("[module][mikan] unbind")
		}
	}
}

func (mikan *Mikan) userSub(userId int64) {
	now := time.Now()
	zap.L().Info("[module][mikan] start mikan user sub", zap.Int64("userId", userId))
	defer func() {
		cost := time.Since(now)
		zap.L().Info("[module][mikan] end mikan user sub", zap.Int64("userId", userId), zap.Duration("cost", cost))
	}()
	sub, err := mikan.getMikanUserSub(userId)
	if err != nil {
		zap.L().Error("[module][mikan] get mikan user sub fail", zap.Int64("userId", userId), zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", env.GetGroupId(), segment.BuildTextSegment("获取用户绑定信息时出错，查看日志获取详情"))
		mikan.bus.Send(msg)
		return
	}
	feed, err := mikan.getRSSFeed(sub.RssUrl)
	if err != nil {
		zap.L().Error("[module][mikan] get mikan rss feed fail", zap.Int64("userId", userId), zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", sub.GroupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf("从指定源获取更新失败， 错误: %v", err.Error())))
		mikan.bus.Send(msg)
		return
	}
	linkArr := []string{}
	for _, item := range feed.Channel.Items {
		if _, ok := sub.Read[item.Link]; !ok {
			linkArr = append(linkArr, fmt.Sprintf("标题: %v\nMikan 链接: %v\n种子地址: %v\n\n", item.Description, item.Link, item.Enclosure.URL))
			sub.Read[item.Link] = now.Unix()
		}
	}
	if len(linkArr) == 0 {
		zap.L().Debug("[module][mikan] get mikan rss feed no update", zap.Int64("userId", userId))
		return
	}
	totalPage := int(math.Ceil(float64(len(linkArr)) / float64(mikan.conf.LinkPerPage)))
	for nowPage := 1; nowPage <= totalPage; nowPage++ {
		text := fmt.Sprintf(" 检测到你的 Mikan RSS 源有以下更新【%v/%v】\n\n", nowPage, totalPage)
		for i := (nowPage - 1) * mikan.conf.LinkPerPage; i < len(linkArr) && i < nowPage*mikan.conf.LinkPerPage; i++ {
			text += linkArr[i]
		}
		msg := api.BuildSendGroupMsgRequest("", sub.GroupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(text))
		mikan.bus.Send(msg)
	}
	err = mikan.setMikanUserSub(userId, sub)
	if err != nil {
		zap.L().Error("[module][mikan] set mikan user sub fail", zap.Int64("userId", userId), zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", sub.GroupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment("⚠️ 本次获取更新后持久化失败，可能导致重复订阅更新提醒"))
		mikan.bus.Send(msg)
	}
}

func (mikan *Mikan) getMikanUserList() (userList []int64, err error) {
	key := "mikan_user_list"
	list, err := mikan.rds.SMembers(key)
	if err != nil {
		return
	}
	userList = make([]int64, 0)
	for _, u := range list {
		iu, err := strconv.ParseInt(u, 10, 64)
		if err != nil {
			continue
		}
		userList = append(userList, iu)
	}
	return
}

func (mikan *Mikan) addMikanUserList(userId int64) (err error) {
	key := "mikan_user_list"
	_, err = mikan.rds.SAdd(key, userId)
	return
}

func (mikan *Mikan) getMikanUserSub(userId int64) (sub *mikanSubscription, err error) {
	key := fmt.Sprintf("mikan_user_%v", userId)
	data, err := mikan.rds.Get(key)
	if err != nil {
		return
	}
	sub = &mikanSubscription{}
	err = json.Unmarshal([]byte(data), sub)
	return
}

func (mikan *Mikan) setMikanUserSub(userId int64, sub *mikanSubscription) (err error) {
	key := fmt.Sprintf("mikan_user_%v", userId)
	data, err := json.Marshal(sub)
	if err != nil {
		return
	}
	_, err = mikan.rds.Set(key, string(data), time.Duration(0))
	if err != nil {
		return
	}
	return
}

func (mikan *Mikan) isMikanUserExist(userId int64) (isExist bool, err error) {
	key := fmt.Sprintf("mikan_user_%v", userId)
	return mikan.rds.IsExist(key)
}
