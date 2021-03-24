package util

import (
	"reflect"
	"unsafe"
)

// StringToBytesFast convert string to []byte quickly by avoiding underlying (indirect part) memory allocation
func StringToBytesFast(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: 0}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
