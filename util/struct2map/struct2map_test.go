package struct2map

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

type TargetStruct struct {
	A         string                 `json:"a"`
	B         int64                  `json:"b"`
	C         *string                `json:"c"`
	D         *int64                 `json:"d,omitempty"`
	E         map[string]interface{} `json:"e"`
	Add       *Additional            `json:"add"`
	AddSecond *Additional            `json:"add_second"`
}

func TestTargetStruct(t *testing.T) {
	testStr1 := []byte(`{"a": "a", "b": 2, "c": "c", "d": 3, "add": {"f": "f", "g": 4}}`)
	t1 := TargetStruct{}
	err := jsoniter.Unmarshal(testStr1, &t1)
	assert.Nil(t, err)
	r1 := make(map[string]interface{})
	err = DecodeByJsonTag(&t1, r1, nil)
	assert.Nil(t, err)
	fmt.Println(r1)
}
