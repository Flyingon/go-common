package util

import (
	"net"
	"strings"
)

var localIp = ""

func GetLocalIP() string {
	if localIp == "" {
		conn, err := net.Dial("udp", "www.oa.com:80")
		if err == nil {
			localIp = strings.Split(conn.LocalAddr().String(), ":")[0]
		}
		defer conn.Close()
	}
	return localIp
}
