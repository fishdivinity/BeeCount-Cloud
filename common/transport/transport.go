package transport

import (
	"runtime"
)

// NewTransport 创建适合当前平台的通信实现
func NewTransport() Transport {
	switch runtime.GOOS {
	case "windows":
		// 尝试使用 Windows 命名管道
		pipeTransport := NewWindowsPipeTransport()
		if pipeTransport != nil {
			return pipeTransport
		}
		// 如果命名管道不可用，使用 TCP
		return NewTCPTransport()
	case "linux", "darwin":
		// 使用 Unix 域套接字
		return NewUnixTransport()
	default:
		// 其他平台使用 TCP
		return NewTCPTransport()
	}
}

// NewTransportWithFallback 创建通信实现，支持降级
func NewTransportWithFallback() Transport {
	// 首先尝试使用平台首选的通信方式
	primaryTransport := NewTransport()

	// 如果首选方式不可用，使用 TCP 作为备选
	if primaryTransport == nil {
		return NewTCPTransport()
	}

	return primaryTransport
}
