package filewatcher

import (
	"fmt"
)

type FileCommand struct {
	Help      bool   `cmd:"help"`
	Operation string `cmd:"required0"`
	Param     string `cmd:"required1"`
}

func (c *FileCommand) CheckCommand() (err error) {
	if c.Operation == "" {
		err = fmt.Errorf("empty operation")
		return
	} else {
		if c.Operation != "ls" && c.Operation != "upload" {
			err = fmt.Errorf("operation illegal")
			return
		}
	}
	if c.Param == "" {
		err = fmt.Errorf("empty param")
		return
	}
	return
}
