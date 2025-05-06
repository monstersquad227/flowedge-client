package utils

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestGetAddress(t *testing.T) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Println(ipNet.IP.String())
				return
			}
		}
	}

	fmt.Println("127.0.0.1")
}

func TestGetHostname(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
		return
	}
	fmt.Println("Hostname:", hostname)
}
