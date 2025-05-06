package utils

import (
	"net"
	"os"
)

func GetAddress() string {
	address, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, addr := range address {
		// 检查地址是否是 IP 地址类型
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

func GetAgentID() string {
	return GetHostname() + "_" + GetAddress()
}
