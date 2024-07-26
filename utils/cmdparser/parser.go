package cmdparser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Command interface {
	CheckCommand() error
}

type t struct {
	from  string
	value string
}

// 命令解析，命令形式类似Linux命令，对于解析到bool字段的可选项可以省略value，但不要将该省略项置于可选项最后
// e.g. "draw -bool_test -r 16:9 -out png girl" ✅ "draw -r 16:9 -out png -bool_test girl"❌
// TODO 解析bug修复，具体见test
func Parse(cmd string, target Command) (err error) {
	cmdTags := make(map[string]*t)
	rt := reflect.TypeOf(target).Elem()
	rv := reflect.ValueOf(target).Elem()
	requiredTagMaxIdx := -1
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag, ok := field.Tag.Lookup("cmd")
		if !ok {
			return fmt.Errorf("cmdparser can't find cmd tag in %v.%v", rt.Name(), field.Name)
		}
		cmdTags[tag] = nil
		if strings.HasPrefix(tag, "required") {
			requiredTagMaxIdx++
		}
	}
	// 命令按空格分段
	fields := strings.Fields(cmd)
	ri := 0
	// 从起手式下一个开始
	for i := 1; i < len(fields); {
		if strings.HasPrefix(fields[i], "-") {
			key := strings.TrimPrefix(fields[i], "-")
			// 有对应tag
			if _, ok := cmdTags[key]; ok {
				if i+1 < len(fields) && !strings.HasPrefix(fields[i+1], "-") {
					cmdTags[key] = &t{
						from:  "user",
						value: fields[i+1],
					}
					i += 2
				} else {
					// 在正常输入中只有bool类型的可选项能省略值
					cmdTags[key] = &t{
						from:  "cmdparser",
						value: "true",
					}
					i += 1
				}
			}
		} else {
			key := fmt.Sprintf("required%v", ri)
			if cmdTags[key] == nil {
				cmdTags[key] = &t{
					from:  "user",
					value: fields[i],
				}
			} else {
				oldT := cmdTags[key]
				cmdTags[key] = &t{
					from:  "user",
					value: oldT.value + " " + fields[i],
				}
			}
			i += 1
			// 判断后面还有没有required tag，有的话下次就填下一个
			if ri < requiredTagMaxIdx {
				ri++
			}
		}
	}
	// 回填
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("cmd")
		switch field.Type.Kind() {
		case reflect.String:
			if cmdTags[tag] != nil && cmdTags[tag].from == "user" {
				rv.Field(i).SetString(cmdTags[tag].value)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if cmdTags[tag] != nil && cmdTags[tag].from == "user" {
				iv, err := strconv.Atoi(cmdTags[tag].value)
				if err != nil {
					err = fmt.Errorf("cmdparser parse %v.%v err: %v", rt.Name(), field.Name, err.Error())
					return err
				}
				rv.Field(i).SetInt(int64(iv))
			}
		case reflect.Bool:
			if cmdTags[tag] != nil {
				rv.Field(i).SetBool(cmdTags[tag].value != "")
			}
		}
	}
	return
}
