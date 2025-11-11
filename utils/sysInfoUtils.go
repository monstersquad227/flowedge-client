package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/mem"
	"net"
	"os"
)

var (
	JVMOPTIONS = initEnv()
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

func initEnv() string {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return ""
	}

	// 先转换为 MB
	totalMB := int(memory.Total / 1024 / 1024)

	xmx := int(float64(totalMB) * 0.80) // 80%
	xms := int(float64(xmx) * 0.5)      // 启动时分配 50%
	xmn := int(float64(xmx) * 0.4)      // 新生代 40%

	JvmOpts := fmt.Sprintf("-Xms%dm -Xmx%dm -Xmn%dm -Xss256k", xms, xmx, xmn)
	scriptPath := "/etc/profile.d/jvm_opts.sh"
	content := fmt.Sprintf("export JAVA_OPTS=\"%s\"\n", JvmOpts)
	err = os.WriteFile(scriptPath, []byte(content), 0644)
	if err != nil {
		return ""
	}
	return JvmOpts
}
