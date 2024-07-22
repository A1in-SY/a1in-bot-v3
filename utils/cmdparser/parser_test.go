package cmdparser_test

import (
	"a1in-bot-v3/modules/magicpen"
	"a1in-bot-v3/utils/cmdparser"
	"testing"
)

func Test_CommandParser(t *testing.T) {
	cmd0 := &magicpen.DrawCommand{}
	err := cmdparser.Parse("draw girl", cmd0)
	if err != nil {
		t.Error(err)
	}
}
