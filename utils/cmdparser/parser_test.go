package cmdparser_test

import (
	"a1in-bot-v3/modules/chat"
	"a1in-bot-v3/modules/filewatcher"
	"a1in-bot-v3/modules/magicpen"
	"a1in-bot-v3/utils/cmdparser"
	"fmt"
	"testing"
)

func Test_CommandParser(t *testing.T) {
	cmd0 := &magicpen.DrawCommand{}
	err := cmdparser.Parse("#draw -help white hair vtuber", cmd0)
	if err != nil || cmd0.CheckCommand() != nil {
		t.Error(err)
	}
	fmt.Printf("cmd0: %+v", cmd0)
	cmd1 := &filewatcher.FileCommand{}
	err = cmdparser.Parse("#file ls /home", cmd1)
	if err != nil || cmd1.CheckCommand() != nil {
		t.Error(err)
	}
	fmt.Printf("cmd1: %+v", cmd1)
	cmd2 := &chat.ChatCommand{}
	err = cmdparser.Parse("#chat who are you", cmd2)
	if err != nil || cmd2.CheckCommand() != nil {
		t.Error(err)
	}
	fmt.Printf("cmd2: %+v", cmd2)
}
