package port

import (
	"fmt"
	"net"
)

func tryBind(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	listener.Close()
	return nil
}

// 找到一个可用的端口
func FindAvailablePort(startPort int) int {
	port := startPort
	for {
		if err := tryBind(port); err == nil {
			return port
		}
		port++
	}
}
