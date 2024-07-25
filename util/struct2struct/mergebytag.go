package struct2struct

import (
	"fmt"
	"reflect"
	"strings"
)

// MergeByTag 多个 sources 根据 tag 融合到一个 结构体 target
func MergeByTag(tagName string, target interface{}, sources ...interface{}) error {
	if len(tagName) == 0 {
		return fmt.Errorf("tag name must be valid")
	}
	targetPtrValue := reflect.ValueOf(target)
	if targetPtrValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}
	destValue := targetPtrValue.Elem()
	if !destValue.CanAddr() {
		return fmt.Errorf("result must be addressable (a pointer)")
	}
	tagToKey := make(map[string]string)
	for i := 0; i < destValue.NumField(); i++ {
		tagStr := destValue.Type().Field(i).Tag.Get(tagName)
		if len(tagStr) == 0 {
			continue
		}
		tag, _ := cutSuffix(tagStr, ",omitempty")
		tagToKey[tag] = destValue.Type().Field(i).Name
	}

	for index, source := range sources {
		sourcePtrValue := reflect.ValueOf(source)
		if sourcePtrValue.Kind() != reflect.Ptr {
			return fmt.Errorf("sources[%d] must be a pointer", index)
		}
		sourceValue := sourcePtrValue.Elem()
		if !sourceValue.CanAddr() {
			return fmt.Errorf("sources[%d] must be addressable (a pointer)", index)
		}

		for i := 0; i < sourceValue.NumField(); i++ {
			tagStr := sourceValue.Type().Field(i).Tag.Get(tagName)
			if len(tagStr) == 0 {
				continue
			}
			tag, _ := cutSuffix(tagStr, ",omitempty")
			if tag == "" {
				continue
			}
			key, exist := tagToKey[tag]
			if !exist {
				continue
			}
			destFieldValue := destValue.FieldByName(key)
			if destFieldValue.IsValid() && destFieldValue.CanSet() {
				destFieldValue.Set(sourceValue.Field(i))
			}
		}
	}
	return nil
}

func cutSuffix(s, suffix string) (before string, found bool) {
	if !strings.HasSuffix(s, suffix) {
		return s, false
	}
	return s[:len(s)-len(suffix)], true
}
