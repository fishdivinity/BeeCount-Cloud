//go:build windows

package transport

import (
	"context"
	"fmt"
	"net"
	"runtime"

	"github.com/Microsoft/go-winio"
)

// WindowsPipeTransport Windows 命名管道通信实现
type WindowsPipeTransport struct{}

// NewWindowsPipeTransport 创建 Windows 命名管道通信实现
func NewWindowsPipeTransport() Transport {
	// 检查是否为 Windows 平台
	if runtime.GOOS != "windows" {
		return nil
	}
	return &WindowsPipeTransport{}
}

// NewListener 创建 Windows 命名管道监听器
func (t *WindowsPipeTransport) NewListener(address string) (Listener, error) {
	// Windows 命名管道使用特殊的地址格式：\\.\pipe\pipename
	if address == "" {
		return nil, fmt.Errorf("empty pipe address")
	}

	// 创建 Windows 命名管道监听器，添加安全描述符以解决权限问题
	listener, err := winio.ListenPipe(address, &winio.PipeConfig{
		// 安全描述符：允许系统、管理员和所有用户访问
		SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GA;;;WD)",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create windows pipe listener: %v", err)
	}

	return listener, nil
}

// NewDialer 创建 Windows 命名管道拨号器
func (t *WindowsPipeTransport) NewDialer() Dialer {
	return &windowsPipeDialer{}
}

// windowsPipeDialer Windows 命名管道拨号器
type windowsPipeDialer struct{}

// DialContext 拨号到 Windows 命名管道
func (d *windowsPipeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, address)
}

// DefaultAddress 获取默认 Windows 命名管道地址
func (t *WindowsPipeTransport) DefaultAddress(serviceName string) string {
	return fmt.Sprintf("\\\\.\\pipe\\beecount_%s", serviceName)
}

// ValidateAddress 检查 Windows 命名管道地址是否有效
func (t *WindowsPipeTransport) ValidateAddress(address string) bool {
	return address != ""
}
