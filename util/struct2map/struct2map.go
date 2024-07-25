package struct2map

import (
	"fmt"
	"reflect"
	"strings"
)

// DecodeByJsonTag 通过 json 解析结构体到 map
func DecodeByJsonTag(input interface{}, result map[string]interface{}, keys map[string]struct{}) error {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return fmt.Errorf("result must be addressable (a pointer)")
	}
	if input == nil {
		return nil
	}
	if result == nil {
		result = make(map[string]interface{})
	}

	r := reflect.Indirect(val)
	n := reflect.TypeOf(input)
	if n.Kind() == reflect.Ptr {
		n = n.Elem()
	}

	fieldNum := r.NumField()
	for i := 0; i < fieldNum; i++ {
		tag := n.Field(i).Tag.Get("json")
		if len(tag) == 0 {
			continue
		}
		key, _ := cutSuffix(tag, ",omitempty")
		if keys != nil {
			if _, ok := keys[key]; !ok {
				continue
			}
		}

		var field interface{}
		switch val.Field(i).Kind() {
		case reflect.Struct, reflect.Map: // 先不支持 struct, map，rule-engine 不支持
			continue
		case reflect.Int32: // 为了支持 proto 中的枚举值
			field = val.Field(i).Int()
		case reflect.Pointer: // 处理指针类型
			if val.Field(i).IsZero() || // 指针的零值是 nil，不要了
				reflect.Indirect(val.Field(i)).Kind() == reflect.Struct ||
				reflect.Indirect(val.Field(i)).Kind() == reflect.Map {
				continue
			}
			field = reflect.Indirect(val.Field(i)).Interface()
		default:
			field = val.Field(i).Interface()
		}
		result[key] = field
	}
	return nil
}

func cutSuffix(s, suffix string) (before string, found bool) {
	if !strings.HasSuffix(s, suffix) {
		return s, false
	}
	return s[:len(s)-len(suffix)], true
}
