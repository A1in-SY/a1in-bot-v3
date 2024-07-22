package magicpen

import (
	"a1in-bot-v3/utils/cmdparser"
	"fmt"
)

type DrawCommand struct {
	cmdparser.Command
	Help     bool   `cmd:"help"`
	Ratio    string `cmd:"ratio"`
	Model    string `cmd:"model"`
	Output   string `cmd:"output"`
	Negative string `cmd:"negative"`
	Prompt   string `cmd:"required0"`
}

func (c *DrawCommand) check() (err error) {
	if c.Ratio == "" {
		c.Ratio = "16:9"
	} else {
		if c.Ratio != "16:9" && c.Ratio != "1:1" && c.Ratio != "21:9" && c.Ratio != "2:3" && c.Ratio != "3:2" && c.Ratio != "4:5" && c.Ratio != "5:4" && c.Ratio != "9:16" && c.Ratio != "9:21" {
			err = fmt.Errorf("ratio option illegal")
			return
		}
	}
	if c.Model == "" {
		c.Model = "sd3-large"
	} else {
		if c.Model != "sd3-large" && c.Model != "sd3-large-turbo" && c.Model != "sd3-medium" {
			err = fmt.Errorf("model option illegal")
			return
		}
	}
	if c.Output == "" {
		c.Output = "png"
	} else {
		if c.Output != "png" && c.Output != "jpeg" {
			err = fmt.Errorf("output option illegal")
			return
		}
	}
	return
}
