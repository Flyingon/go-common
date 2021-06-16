package util

import (
	"fmt"
	"strings"
)

// CheckKeyExist 检查keys中的key是否都存在于data中
func CheckKeyExist(data map[string]interface{}, keys []string) error {
	lackKeys := make([]string, 0, len(data))
	for _, key := range keys {
		if _, exist := data[key]; !exist {
			lackKeys = append(lackKeys, key)
		}
	}
	if len(lackKeys) > 0 {
		return fmt.Errorf("keys[%s] is not exist", strings.Join(lackKeys, ","))
	}
	return nil
}

// ParamCheck 单个参数检查结构体
type ParamCheck struct {
	Condition bool
	ErrCode   int
	ErrMsg    string
}

// ParamsCheck 参数检查
func ParamsCheck(pcs []*ParamCheck) (int, error) {
	errCode := 1003
	errMsg := "参数检查失败"
	for _, c := range pcs {
		if c == nil {
			continue
		}
		if c.Condition {
			if c.ErrCode > 0 {
				errCode = c.ErrCode
			}
			if len(c.ErrMsg) > 0 {
				errMsg = c.ErrMsg
			}
			return errCode, fmt.Errorf(errMsg)
		}
	}
	return 0, nil
}
