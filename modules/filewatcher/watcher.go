package filewatcher

import (
	"a1in-bot-v3/model/api"
	"a1in-bot-v3/model/event"
	"a1in-bot-v3/model/segment"
	"a1in-bot-v3/utils/bus"
	"a1in-bot-v3/utils/cmdparser"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

type FileWatcher struct {
	bus *bus.BusChan
}

func (w *FileWatcher) InitModule(cbs []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("when init filewatcher module, an error happen: %v", err.Error())
		}
	}()
	//c := conf.DefaultConf()
	//err = toml.Unmarshal(cbs, c)
	//if err != nil {
	//	return
	//}
	//err = c.Check()
	//if err != nil {
	//	return
	//}
	w.bus = bus.GetBus().GenBusChan(event.EventId_MessageEventGroupMessage)
	zap.L().Info("[module][filewatcher] init successfully")
	return
}

func (w *FileWatcher) Run() {
	for {
		e := w.bus.Read()
		if w.match(e) {
			w.handle(e)
		}
	}
}

func (w *FileWatcher) Cleanup() (err error) {
	return
}

func (w *FileWatcher) match(e *event.Event) (isMatch bool) {
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
	if !strings.HasPrefix(text, "#file") {
		return
	}
	cmd := &FileCommand{}
	err := cmdparser.Parse(text, cmd)
	if err != nil {
		zap.L().Error("[module][filewatch] parse file command fail", zap.Error(err))
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildTextSegment(fmt.Sprintf("解析命令 %v 时出错: %v", text, err.Error())))
		w.bus.Send(msg)
		return
	}
	isMatch = true
	return
}

func (w *FileWatcher) handle(e *event.Event) {
	eventData := e.EventData.(*event.Event_GroupMsg)
	cmd := strings.Split(strings.TrimLeft(eventData.GroupMsg.GetMessage()[1].Data.Text, " "), " ")
	userId := eventData.GroupMsg.GetUserId()
	groupId := eventData.GroupMsg.GetGroupId()
	operation := cmd[1]
	path := cmd[2]
	zap.L().Info("[module][filewatch] handle", zap.Any("event", e))
	if operation == "ls" {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				zap.L().Error("[module][filewatch] dir not exist", zap.String("path", path), zap.Int64("userId", userId))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 指定的文件夹不存在"))
				w.bus.Send(msg)
			} else {
				zap.L().Error("[module][filewatch] get dir info fail", zap.String("path", path), zap.Int64("userId", userId), zap.Error(err))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 获取文件夹信息时出错: "+err.Error()))
				w.bus.Send(msg)
			}
			return
		}
		files, err := os.ReadDir(path)
		if err != nil {
			zap.L().Error("[module][filewatch] get dir info fail", zap.String("path", path), zap.Int64("userId", userId), zap.Error(err))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 获取文件夹信息时出错: "+err.Error()))
			w.bus.Send(msg)
			return
		}
		text := " 指定文件夹下有以下内容: \n"
		for _, file := range files {
			if file.IsDir() {
				text += fmt.Sprintf("目录  %v\n", file.Name())
			} else {
				text += fmt.Sprintf("文件  %v\n", file.Name())
			}
		}
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(text))
		w.bus.Send(msg)
		return
	} else if operation == "upload" {
		fileInfo, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				zap.L().Error("[module][filewatch] file not exist", zap.String("path", path), zap.Int64("userId", userId))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 指定的文件不存在"))
				w.bus.Send(msg)
			} else {
				zap.L().Error("[module][filewatch] get file info fail", zap.String("path", path), zap.Int64("userId", userId), zap.Error(err))
				msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 获取文件信息时出错: "+err.Error()))
				w.bus.Send(msg)
			}
			return
		}
		if fileInfo.IsDir() {
			zap.L().Error("[module][filewatch] path is dir", zap.String("path", path), zap.Int64("userId", userId))
			msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" 指定路径为文件夹"))
			w.bus.Send(msg)
		} else {
			msg := api.BuildUploadGroupFileRequest("", groupId, path, fileInfo.Name())
			w.bus.Send(msg)
		}
		return
	} else {
		msg := api.BuildSendGroupMsgRequest("", groupId, segment.BuildAtSegment(fmt.Sprint(userId)), segment.BuildTextSegment(" file 插件支持的命令有: \nls 列出目录下文件与文件夹\nupload 上传指定文件"))
		w.bus.Send(msg)
		return
	}
}
