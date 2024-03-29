package util

import (
	"bytes"
	"fmt"
	json "github.com/json-iterator/go"
	"strings"
)

// JSONMarshal json序列化，escape设置为false
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// JSONUnMarshal 设置UseNumber,防止uint64位精度缺失
func JSONUnMarshal(jsonStream []byte, ret interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(string(jsonStream)))
	decoder.UseNumber()
	if err := decoder.Decode(&ret); err != nil {
		fmt.Println("error:", err)
		return err
	}
	return nil
}

// ValueToStr JSONUnMarshal到map[string]interface{}后interface{} to string
func ValueToStr(v interface{}) (ret string) {
	switch v.(type) {
	case string:
		ret = v.(string)
	case json.Number:
		ret = v.(json.Number).String()
	default:
		val, _ := JSONMarshal(v)
		ret = string(val)
	}
	return
}

// JSONMarshal json序列化，escape设置为false
func JSONMarshalToString(t interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return string(buffer.Bytes()), err
}
