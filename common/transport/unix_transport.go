package transport

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// UnixTransport Unix 域套接字通信实现
type UnixTransport struct{}

// NewUnixTransport 创建 Unix 域套接字通信实现
func NewUnixTransport() Transport {
	return &UnixTransport{}
}

// NewListener 创建 Unix 域套接字监听器
func (t *UnixTransport) NewListener(address string) (Listener, error) {
	// 确保套接字文件所在目录存在
	if address != "" {
		dir := filepath.Dir(address)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create socket directory: %v", err)
			}
		}

		// 移除已存在的套接字文件
		if _, err := os.Stat(address); err == nil {
			if err := os.Remove(address); err != nil {
				return nil, fmt.Errorf("failed to remove existing socket file: %v", err)
			}
		}
	}

	// 创建 Unix 域套接字监听器
	listener, err := net.Listen("unix", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create unix listener: %v", err)
	}

	return listener, nil
}

// NewDialer 创建 Unix 域套接字拨号器
func (t *UnixTransport) NewDialer() Dialer {
	return &net.Dialer{}
}

// DefaultAddress 获取默认 Unix 域套接字地址
func (t *UnixTransport) DefaultAddress(serviceName string) string {
	return fmt.Sprintf("/tmp/beecount_%s.sock", serviceName)
}

// ValidateAddress 检查 Unix 域套接字地址是否有效
func (t *UnixTransport) ValidateAddress(address string) bool {
	return address != ""
}
