package magicpen

import (
	"a1in-bot-v3/infrastructure/env"
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
	"a1in-bot-v3/model/segment"
	"a1in-bot-v3/modules/magicpen/conf"
	"a1in-bot-v3/utils/bus"
	"a1in-bot-v3/utils/cmdparser"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type MagicPen struct {
	conf    *conf.MagicPenConfig
	bus     *bus.BusChan
	httpCli *http.Client
}

func (mp *MagicPen) InitModule(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init mp module, an error happen: %v", err.Error())
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
	mp.conf = c.MagicPenConf
	mp.conf.Sk = env.GetStabilityAPISk()
	mp.bus = bus.GetBus().GenBusChan(event.EventId_MessageEventGroupMessage)
	if mp.conf.UseProxy {
		pu, err := url.Parse(mp.conf.ProxyAddr)
		if err != nil {
			return err
		}
		mp.httpCli = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(pu),
			},
			Timeout: mp.conf.Timeout,
		}
	} else {
		mp.httpCli = &http.Client{
			Timeout: mp.conf.Timeout,
		}
	}
	zap.L().Info("[module][magicpen] init successfully")
	return
}

func (mp *MagicPen) Run() {
	for {
		e := mp.bus.Read()
		if ok, cmd := mp.match(e); ok {
			mp.handle(e, cmd)
		}
	}
}

func (mp *MagicPen) Cleanup() (err error) {
	return
}

func (mp *MagicPen) match(e *event.Event) (isMatch bool, cmd *DrawCommand) {
	eventData, ok := e.EventData.(*event.Event_GroupMsg)
	if !ok {
		return
	}
	// at和文本
	if len(eventData.GroupMsg.GetMessage()) != 2 {
		return
	}
	if eventData.GroupMsg.GetMessage()[0].Type != segment.SegmentTypeAt {
		return
	}
	// at机器人
	if eventData.GroupMsg.GetMessage()[0].Data.Qq != strconv.FormatInt(e.SelfId, 10) {
		return
	}
	if eventData.GroupMsg.GetMessage()[1].Type != segment.SegmentTypeText {
		return
	}
	// 文本开头为指定字符串
	if strings.Split(strings.TrimLeft(eventData.GroupMsg.GetMessage()[1].Data.Text, " "), " ")[0] != "draw" {
		return
	}
	cmd = &DrawCommand{}
	err := cmdparser.Parse(strings.TrimLeft(eventData.GroupMsg.GetMessage()[1].Data.Text, " "), cmd)
	if err != nil {
		zap.L().Error("[module][magicpen] parse draw command fail", zap.Error(err))
		return
	}
	isMatch = true
	return
}

func (mp *MagicPen) handle(e *event.Event, cmd *DrawCommand) {
	eventData := e.EventData.(*event.Event_GroupMsg)
	userId := eventData.GroupMsg.GetUserId()
	groupId := eventData.GroupMsg.GetGroupId()
	zap.L().Info("[module][magicpen] handle", zap.Int64("userId", userId), zap.Int64("groupId", groupId), zap.Any("cmd", cmd))
	if cmd.Help {
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment("draw命令格式: draw [options] your_prompt_text\n"+
			"支持的可选项有:\n"+
			"-help 查看帮助\n"+
			"-ratio 设置长宽比，有16:9 1:1 21:9 2:3 3:2 4:5 5:4 9:16 9:21可选，默认为16:9\n"+
			"-model 选择模型，有sd3-large sd3-large-turbo sd3-medium可选，默认为sd3-large\n"+
			"-out 输出格式，有jpeg png，默认为png\n"+
			"-negative 负面提示词"))
		mp.bus.Send(msg)
		return
	} else {
		err := cmd.check()
		if err != nil {
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 命令参数不合法: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 正在画图中..."))
		mp.bus.Send(msg)
		res, err := mp.draw(cmd)
		if err != nil {
			zap.L().Error("[module][magicpen] draw fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 调用Stability API失败， 错误: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		picData, err := base64.StdEncoding.DecodeString(res.Image)
		if err != nil {
			zap.L().Error("[module][magicpen] decode png fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 解码图片失败， 错误: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		picName := fmt.Sprintf("%v_%v.png", userId, time.Now().Unix())
		picPath, err := filepath.Abs(filepath.Join(mp.conf.PicStoragePath, picName))
		if err != nil {
			zap.L().Error("[module][magicpen] abs png path fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 获取本地路径失败， 错误: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		if _, te := os.Stat(mp.conf.PicStoragePath); os.IsNotExist(te) {
			os.Mkdir(mp.conf.PicStoragePath, os.ModePerm)
		}
		picFile, err := os.Create(picPath)
		if err != nil {
			zap.L().Error("[module][magicpen] create png file fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 创建本地图片失败， 错误: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		defer picFile.Close()
		_, err = picFile.Write(picData)
		if err != nil {
			zap.L().Error("[module][magicpen] write png file fail", zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(fmt.Sprintf(" 写入本地图片失败， 错误: %v", err.Error())))
			mp.bus.Send(msg)
			return
		}
		msg = api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildImageSegment(fmt.Sprintf("file://%v", picPath)))
		mp.bus.Send(msg)
		return
	}
}
