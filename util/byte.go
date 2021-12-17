package util

import (
	"fmt"
	"math"
	"strconv"
)

// GetPosByteNum 获取指定位置全为0或者为1的数字
// in: 1 out: 1 11111111111111111110
// in: 5 out: 10000 11111111111111101111
func GetPosByteNum(pos int) (int, int) {
	setNum := int(math.Pow(2, float64(pos)-1))
	unSetNum, _ := strconv.ParseInt(fmt.Sprintf("%d", 0xfffff^setNum), 10, 64)
	return setNum, int(unSetNum)
}
