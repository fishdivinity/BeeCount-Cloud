package transport

import (
	"fmt"
	"net"
	"strconv"
)

// TCPTransport TCP 通信实现
type TCPTransport struct{}

// NewTCPTransport 创建 TCP 通信实现
func NewTCPTransport() Transport {
	return &TCPTransport{}
}

// NewListener 创建 TCP 监听器
func (t *TCPTransport) NewListener(address string) (Listener, error) {
	// 创建 TCP 监听器
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create tcp listener: %v", err)
	}
	
	return listener, nil
}

// NewDialer 创建 TCP 拨号器
func (t *TCPTransport) NewDialer() Dialer {
	return &net.Dialer{}
}

// DefaultAddress 获取默认 TCP 地址
func (t *TCPTransport) DefaultAddress(serviceName string) string {
	// 为每个服务分配默认端口
	port := 50050
	switch serviceName {
	case "config":
		port = 50051
	case "log":
		port = 50052
	case "auth":
		port = 50053
	case "business":
		port = 50054
	case "storage":
		port = 50055
	case "firewall":
		port = 50056
	case "gateway":
		port = 50057
	}
	return fmt.Sprintf(":%d", port)
}

// ValidateAddress 检查 TCP 地址是否有效
func (t *TCPTransport) ValidateAddress(address string) bool {
	if address == "" {
		return false
	}
	
	// 检查地址格式是否为 :port
	if address[0] == ':' {
		portStr := address[1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return false
		}
		return port > 0 && port < 65536
	}
	
	// 检查地址格式是否为 host:port
	_, err := net.ResolveTCPAddr("tcp", address)
	return err == nil
}
