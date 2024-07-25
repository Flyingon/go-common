package struct2struct

import (
	"fmt"
	"testing"

	jsoniter "github.com/json-iterator/go"

	"github.com/stretchr/testify/assert"
)

type Additional struct {
	F string `json:"f"`
	G int64  `json:"g"`
}

type SrcStruct1 struct {
	A string                 `json:"a"`
	B int64                  `json:"b"`
	C *string                `json:"c"`
	D *int64                 `json:"d,omitempty"`
	E map[string]interface{} `json:"e"`
	F *Additional            `json:"f"`
	G *Additional            `json:"g"`
}

type SrcStruct2 struct {
	H string                 `json:"h"`
	I int64                  `json:"i"`
	J *string                `json:"j"`
	K *int64                 `json:"k,omitempty"`
	L map[string]interface{} `json:"l"`
	M *Additional            `json:"m"`
	N *Additional            `json:"n,omitempty"`
}

type Target struct {
	A string                 `json:"a"`
	B int64                  `json:"b"`
	C *string                `json:"c"`
	D *int64                 `json:"d,omitempty"`
	E map[string]interface{} `json:"e"`
	F *Additional            `json:"f"`
	G *Additional            `json:"g"`
	H string                 `json:"h"`
	I int64                  `json:"i"`
	J *string                `json:"j"`
	K *int64                 `json:"k,omitempty"`
	L map[string]interface{} `json:"l"`
	M *Additional            `json:"m"`
}

func Test_MergeByTag(t *testing.T) {
	testStr1 := []byte(`{"a": "a", "b": 2, "c": "c", "d": 3, "e": {"e": 1}, "f": {"f": "f", "x": 4}}`)
	testStr2 := []byte(`{"h": "h", "i": 2, "j": "j", "k": 3, "l": {"l": "l"}, "m": {"m": "m", "y": 4}}`)
	src1 := SrcStruct1{}
	src2 := SrcStruct2{}
	err := jsoniter.Unmarshal(testStr1, &src1)
	assert.Nil(t, err)
	err = jsoniter.Unmarshal(testStr2, &src2)
	assert.Nil(t, err)
	target := Target{}
	err = MergeByTag("json", &target, &src1, &src2)
	assert.Nil(t, err)

	targetJson, _ := jsoniter.MarshalToString(target)

	fmt.Println(targetJson)

}
