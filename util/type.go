package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// 将interface{} 转换成 []byte, 支持string, float64和[]byte
func GetBytesByInterface(key interface{}) []byte {
	var ret string
	switch key.(type) {
	case string:
		ret = key.(string)
	case int:
		temp := key.(int)
		ret = strconv.Itoa(temp)
	case float64:
		ret = strconv.FormatFloat(key.(float64), 'f', -1, 64)
	case []byte:
		ret = string(key.([]byte))
	}
	return []byte(ret)
}

// InterfaceToString 将interface{} 转换成 string
func InterfaceToString(key interface{}) string {
	if key == nil { // nil返回空字符串，根据需求添加
		return ""
	}
	var ret string
	switch key.(type) {
	case string:
		ret = key.(string)
	case int:
		ret = strconv.FormatInt(int64(key.(int)), 10)
	case int8:
		ret = strconv.FormatInt(int64(key.(int8)), 10)
	case int16:
		ret = strconv.FormatInt(int64(key.(int16)), 10)
	case int32:
		ret = strconv.FormatInt(int64(key.(int32)), 10)
	case int64:
		ret = strconv.FormatInt(key.(int64), 10)
	case uint:
		ret = strconv.FormatUint(uint64(key.(uint)), 10)
	case uint8:
		ret = strconv.FormatUint(uint64(key.(uint8)), 10)
	case uint16:
		ret = strconv.FormatUint(uint64(key.(uint16)), 10)
	case uint32:
		ret = strconv.FormatUint(uint64(key.(uint32)), 10)
	case uint64:
		ret = strconv.FormatUint(key.(uint64), 10)
	case float32:
		ret = strconv.FormatFloat(float64(key.(float32)), 'f', -1, 64)
	case float64:
		ret = strconv.FormatFloat(key.(float64), 'f', -1, 64)
	case bool:
		ret = strconv.FormatBool(key.(bool))
	case []byte:
		ret = string(key.([]byte))
	case json.Number:
		ret = key.(json.Number).String()
	default:
		retBytes, _ := json.Marshal(key)
		ret = string(retBytes)
	}
	return ret
}

// TranStringToType 转换string类型到指定类型
func TranStringToType(valStr, keyType string) (interface{}, error) {
	var valFmt interface{}
	var err error
	switch keyType {
	case "int":
		valFmt, err = strconv.ParseInt(valStr, 10, 0)
	case "int8":
		valFmt, err = strconv.ParseInt(valStr, 10, 8)
	case "int16":
		valFmt, err = strconv.ParseInt(valStr, 10, 16)
	case "int32":
		valFmt, err = strconv.ParseInt(valStr, 10, 32)
	case "int64":
		valFmt, err = strconv.ParseInt(valStr, 10, 64)
	case "uint":
		valFmt, err = strconv.ParseUint(valStr, 10, 0)
	case "uint8":
		valFmt, err = strconv.ParseUint(valStr, 10, 8)
	case "uint16":
		valFmt, err = strconv.ParseUint(valStr, 10, 16)
	case "uint32":
		valFmt, err = strconv.ParseUint(valStr, 10, 32)
	case "uint64":
		valFmt, err = strconv.ParseUint(valStr, 10, 64)
	case "float", "float32":
		valFmt, err = strconv.ParseFloat(valStr, 32)
	case "float64":
		valFmt, err = strconv.ParseFloat(valStr, 64)
	case "bool":
		valFmt, err = strconv.ParseBool(valStr)
	default:
		err = fmt.Errorf("invalid type: %v", keyType)
	}
	if err != nil {
		return nil, err
	}
	return valFmt, nil
}

// GetZeroValueByType 获得指定类型到0值
func GetZeroValueByType(keyType string) interface{} {
	var ret interface{}
	switch keyType {
	case "string":
		ret = ""
	case "int":
		ret = 0
	case "int8":
		ret = int8(0)
	case "int16":
		ret = int16(0)
	case "int32":
		ret = int32(0)
	case "int64":
		ret = int64(0)
	case "uint":
		ret = uint(0)
	case "uint8":
		ret = uint8(0)
	case "uint16":
		ret = uint16(0)
	case "uint32":
		ret = uint32(0)
	case "uint64":
		ret = uint64(0)
	case "float", "float32":
		ret = float32(0)
	case "float64":
		ret = float64(0)
	case "bool":
		ret = false
	default:
		ret = ""
	}
	return ret
}

// IsEmptyValue 零值判断
func IsEmptyValue(key interface{}) bool {
	switch key.(type) {
	case string:
		return len(key.(string)) == 0
	case int:
		return key.(int) == 0
	case int8:
		return key.(int8) == 0
	case int16:
		return key.(int16) == 0
	case int32:
		return key.(int32) == 0
	case int64:
		return key.(int64) == 0
	case uint:
		return key.(uint) == 0
	case uint8:
		return key.(uint8) == 0
	case uint16:
		return key.(uint16) == 0
	case uint32:
		return key.(uint32) == 0
	case uint64:
		return key.(uint64) == 0
	case float32:
		return key.(float32) == 0
	case float64:
		return key.(float64) == 0
	case bool:
		return key.(bool) == false
	case []byte:
		return len(key.([]byte)) < 2
	case json.Number:
		return key.(json.Number).String() == "0"
	}
	return false
}

// InterfaceCopy 结构体拷贝-注意: 传入结构体对象, 传指针只能复制地址，没有意义
func InterfaceCopy(d interface{}) interface{} {
	n := reflect.New(reflect.TypeOf(d))
	n.Elem().Set(reflect.ValueOf(d))
	return n.Interface()
}

// InterfaceToInt 将interface{} 转换成 int
// 强制转换，忽略精度丢失; 忽略忽略错误，错误返回0
func InterfaceToInt(key interface{}) int {
	if key == nil { // nil返回零值，根据需求添加
		return 0
	}
	var ret int
	switch key.(type) {
	case string:
		tmp, _ := strconv.ParseFloat(key.(string), 64)
		ret = int(tmp)
	case int:
		ret = key.(int)
	case int8:
		ret = int(key.(int8))
	case int16:
		ret = int(key.(int16))
	case int32:
		ret = int(key.(int32))
	case int64:
		ret = int(key.(int64))
	case uint:
		ret = int(key.(uint))
	case uint8:
		ret = int(key.(uint8))
	case uint16:
		ret = int(key.(uint16))
	case uint32:
		ret = int(key.(uint32))
	case uint64:
		ret = int(key.(uint64))
	case float32:
		ret = int(key.(float32))
	case float64:
		ret = int(key.(float64))
	default:
	}
	return ret
}

// InterfaceToInt64 将interface{} 转换成 int64
// 强制转换，忽略精度丢失; 忽略忽略错误，错误返回0
func InterfaceToInt64(key interface{}) int64 {
	if key == nil { // nil返回零值，根据需求添加
		return 0
	}
	var ret int64
	switch key.(type) {
	case string:
		tmp, _ := strconv.ParseFloat(key.(string), 64)
		ret = int64(tmp)
	case int:
		ret = int64(key.(int))
	case int8:
		ret = int64(key.(int8))
	case int16:
		ret = int64(key.(int16))
	case int32:
		ret = int64(key.(int32))
	case int64:
		ret = key.(int64)
	case uint:
		ret = int64(key.(uint))
	case uint8:
		ret = int64(key.(uint8))
	case uint16:
		ret = int64(key.(uint16))
	case uint32:
		ret = int64(key.(uint32))
	case uint64:
		ret = int64(key.(uint64))
	case float32:
		ret = int64(key.(float32))
	case float64:
		ret = int64(key.(float64))
	default:
	}
	return ret
}

// InterfaceToFloat64 ...
func InterfaceToFloat64(key interface{}) float64 {
	if key == nil { // nil返回零值，根据需求添加
		return 0
	}
	var ret float64
	switch key.(type) {
	case string:
		tmp, _ := strconv.ParseFloat(key.(string), 64)
		ret = float64(tmp)
	case int:
		ret = float64(key.(int))
	case int8:
		ret = float64(key.(int8))
	case int16:
		ret = float64(key.(int16))
	case int32:
		ret = float64(key.(int32))
	case int64:
		ret = float64(key.(int64))
	case uint:
		ret = float64(key.(uint))
	case uint8:
		ret = float64(key.(uint8))
	case uint16:
		ret = float64(key.(uint16))
	case uint32:
		ret = float64(key.(uint32))
	case uint64:
		ret = float64(key.(uint64))
	case float32:
		ret = float64(key.(float32))
	case float64:
		ret = key.(float64)
	default:
	}
	return ret
}

func InterfaceToStrings(key interface{}) []string {
	var ret []string
	switch key.(type) {
	case []string:
		return key.([]string)
	case []interface{}:
		for _, i := range key.([]interface{}) {
			ret = append(ret, InterfaceToString(i))
		}
	default:
	}
	return ret
}

func InterfacesToStrings(s []interface{}) []string {
	ss := make([]string, 0, len(s))
	for _, i := range s {
		ss = append(ss, InterfaceToString(i))
	}
	return ss
}

// IntsToStrings [1,2,3] => ["1","2","3"]
func IntsToStrings(s []int) []string {
	ss := make([]string, 0, len(s))
	for _, i := range s {
		ss = append(ss, strconv.Itoa(int(i)))
	}
	return ss
}

// IntsToString [1,2,3] => "1,2,3"
func IntsToString(s []int, sep string) string {
	return strings.Join(IntsToStrings(s), sep)
}

// StringToInts "1,2,3" => [1,2,3]
func StringToInts(s string, sep string) []int {
	arrStr := strings.Split(s, sep)
	ints := make([]int, 0, len(arrStr))
	for _, str := range arrStr {
		ints = append(ints, InterfaceToInt(str))
	}
	return ints
}

// BatchGetValues 遍历获取[]*Struct,[]Struct,[]map[string]interface 中指定字段
func BatchGetValues(t interface{}, key string) []interface{} {
	var res []interface{}
	switch reflect.TypeOf(t).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(t)
		for i := 0; i < s.Len(); i++ {
			switch s.Index(i).Kind() {
			case reflect.Ptr:
				v := s.Index(i).Elem()
				res = append(res, v.FieldByName(key).Interface())
			case reflect.Struct:
				v := s.Index(i)
				res = append(res, v.FieldByName(key).Interface())
			case reflect.Map:
				res = append(res, s.Index(i).MapIndex(reflect.ValueOf(key)).Interface())
			default:
				return res
			}
		}
	}
	return res
}

// SortedAppend 顺序插入
func SortedAppend(ids *[]int, id int) {
	_, ok := SortedFind(*ids, id)
	if ok {
		return
	}

	sorted := append(*ids, id)
	sort.Ints(sorted)
	*ids = sorted
}

// SortedRem 顺序
func SortedRem(ids *[]int, id int) {
	index, ok := SortedFind(*ids, id)
	if !ok {
		return
	}

	*ids = append((*ids)[:index], (*ids)[index+1:]...)
}

func SortedFind(ids []int, id int) (index int, find bool) {
	index = sort.SearchInts(ids, id)
	if index < len(ids) && ids[index] == id {
		return index, true
	}
	return -1, false
}

/*
铺平嵌套map
例如:
{
	"k1":{
		"k2":v
	}
}
铺平后为
{
	"k1.k2":v
}
*/
func MapSpread(m map[string]interface{}) {
	for k, v := range m {
		if mi, ok := v.(map[string]interface{}); ok {
			MapSpread(mi)
			for mik, miv := range mi {
				m[fmt.Sprintf("%s.%s", k, mik)] = miv
			}
			delete(m, k)
		}
	}
}

func MapDot(m map[string]interface{}, key string) interface{} {
	MapSpread(m)
	return m[key]
}

// InterfaceToUint64 将interface{} 转换成 uint64
// 强制转换，忽略精度丢失; 忽略忽略错误，错误返回0
func InterfaceToUint64(key interface{}) uint64 {
	if key == nil { // nil返回零值，根据需求添加
		return 0
	}
	var ret uint64
	switch key.(type) {
	case string:
		tmp, _ := strconv.ParseUint(key.(string), 10, 64)
		ret = tmp
	case int:
		ret = uint64(key.(int))
	case int8:
		ret = uint64(key.(int8))
	case int16:
		ret = uint64(key.(int16))
	case int32:
		ret = uint64(key.(int32))
	case int64:
		ret = uint64(key.(int64))
	case uint:
		ret = uint64(key.(uint))
	case uint8:
		ret = uint64(key.(uint8))
	case uint16:
		ret = uint64(key.(uint16))
	case uint32:
		ret = uint64(key.(uint32))
	case uint64:
		ret = key.(uint64)
	case float32:
		ret = uint64(key.(float32))
	case float64:
		ret = uint64(key.(float64))
	default:
	}
	return ret
}

// SetReqDefault 设置proto请求里面的默认值
func SetReqDefault(req interface{}, kv map[string]interface{}) {
	e := reflect.ValueOf(req).Elem()
	for k, v := range kv {
		if field := e.FieldByName(k); field.CanSet() {
			if field.IsZero() {
				field.Set(reflect.ValueOf(v))
			}
		}
	}
}
