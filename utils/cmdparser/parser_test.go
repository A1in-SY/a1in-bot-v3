package cmdparser_test

import (
	"a1in-bot-v3/modules/filewatcher"
	"a1in-bot-v3/modules/magicpen"
	"a1in-bot-v3/utils/cmdparser"
	"fmt"
	"testing"
)

func Test_CommandParser(t *testing.T) {
	cmd0 := &magicpen.DrawCommand{}
	err := cmdparser.Parse("#draw white hair vtuber", cmd0)
	if err != nil || cmd0.CheckCommand() != nil {
		t.Error(err)
	}
	// 已知bug: 这里只能解析到white
	fmt.Printf("cmd0: %+v", cmd0)
	cmd1 := &filewatcher.FileCommand{}
	err = cmdparser.Parse("#file ls /home", cmd1)
	if err != nil || cmd1.CheckCommand() != nil {
		t.Error(err)
	}
	fmt.Printf("cmd0: %+v", cmd1)
}
