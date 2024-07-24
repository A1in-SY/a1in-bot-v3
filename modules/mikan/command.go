package mikan

import "fmt"

type MikanCommand struct {
	Help      bool   `cmd:"help"`
	Operation string `cmd:"required0"`
	Param     string `cmd:"required1"`
}

func (c *MikanCommand) CheckCommand() (err error) {
	if c.Operation == "" {
		err = fmt.Errorf("empty operation")
		return
	} else {
		if c.Operation != "bind" && c.Operation != "unbind" {
			err = fmt.Errorf("operation illegal")
			return
		}
	}
	if c.Operation == "bind" && c.Param == "" {
		err = fmt.Errorf("empty param when operation is bind")
		return
	}
	return
}
