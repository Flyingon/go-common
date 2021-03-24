package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func CharsetConvert(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	utf8 := srcCoder.ConvertString(src)

	tagCoder := mahonia.NewEncoder(tagCode)
	return tagCoder.ConvertString(utf8)
}

var supportedCharSet = map[string]struct{}{
	"UTF-8":    {},
	"UTF-16BE": {},
	"GB-18030": {},
	"GBK":      {},
	"BIG-5":    {},
}

func IsCharsetSupported(charset string) bool {
	_, ok := supportedCharSet[charset]
	return ok
}

func DetectCharSet(bs []byte) (string, error) {
	rs, err := chardet.NewTextDetector().DetectAll(bs)
	if err != nil {
		return "", err
	}
	charsets := make([]string, 0, len(rs))
	for _, r := range rs {
		charsets = append(charsets, r.Charset)
		switch {
		case r.Confidence == 100:
			return r.Charset, nil
		case r.Charset == "UTF-8":
			return "UTF-8", nil
		case strings.HasPrefix(r.Charset, "UTF"):
			return r.Charset, nil
		case r.Language == "zh" && strings.HasPrefix(r.Charset, "GB"):
			return r.Charset, nil
		}
	}
	return "", fmt.Errorf("不支持的字符集，可能 %+v", charsets)
}

// transform GBK bytes to UTF-8 bytes
func GbkToUtf8(in []byte) ([]byte, error) {
	r := transform.NewReader(bytes.NewReader(in), simplifiedchinese.GBK.NewDecoder())
	return ioutil.ReadAll(r)
}

// transform UTF-8 bytes to GBK bytes
func Utf8ToGbk(in []byte) ([]byte, error) {
	r := transform.NewReader(bytes.NewReader(in), simplifiedchinese.GBK.NewEncoder())
	return ioutil.ReadAll(r)
}
