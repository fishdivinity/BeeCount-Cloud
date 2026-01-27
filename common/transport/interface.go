package transport

import (
	"context"
	"net"
)

// Listener 监听接口
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
}

// Dialer 拨号接口
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Transport 通信抽象接口
type Transport interface {
	// 创建监听器
	NewListener(address string) (Listener, error)
	// 创建拨号器
	NewDialer() Dialer
	// 获取默认地址格式
	DefaultAddress(serviceName string) string
	// 检查地址是否有效
	ValidateAddress(address string) bool
}
