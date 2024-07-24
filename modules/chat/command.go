package chat

import "fmt"

type ChatCommand struct {
	Help    bool   `cmd:"help"`
	Content string `cmd:"required0"`
}

func (c *ChatCommand) CheckCommand() (err error) {
	if c.Content == "" {
		err = fmt.Errorf("empty content")
		return
	}
	return
}
