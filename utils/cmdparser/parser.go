package cmdparser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Command interface {
	isCommand()
}

// 命令解析，命令形式类似Linux命令，对于解析到bool字段的可选项可以省略value，但不要将该省略项置于可选项最后
// e.g. "draw -bool_test -r 16:9 -out png girl" ✅ "draw -r 16:9 -out png -bool_test girl"❌
func Parse(cmd string, target Command) (err error) {
	fields := strings.Fields(cmd)
	fm := make(map[string]string)
	rpi := 0
	for i := 1; i < len(fields); {
		if strings.HasPrefix(fields[i], "-") {
			key := strings.TrimPrefix(fields[i], "-")
			if i+1 < len(fields) && !strings.HasPrefix(fields[i+1], "-") {
				fm[key] = fields[i+1]
				i += 2
			} else {
				fm[key] = ""
				i += 1
			}
		} else {
			key := fmt.Sprintf("required%v", rpi)
			fm[key] = fields[i]
			rpi++
			i++
		}
	}
	rt := reflect.TypeOf(target).Elem()
	rv := reflect.ValueOf(target).Elem()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("cmd")
		if v, ok := fm[tag]; ok {
			switch field.Type.Kind() {
			case reflect.String:
				rv.Field(i).SetString(v)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				iv, err := strconv.Atoi(v)
				if err != nil {
					err = fmt.Errorf("cmdparser parse %v.%v err: %v", rt.Name(), field.Name, err.Error())
					return err
				}
				rv.Field(i).SetInt(int64(iv))
			case reflect.Bool:
				rv.Field(i).SetBool(v == "true" || v == "")
			}
		}
	}
	return
}
