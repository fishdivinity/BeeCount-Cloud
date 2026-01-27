package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/firewall"
	"github.com/fishdivinity/BeeCount-Cloud/common/transport"
	"github.com/fishdivinity/BeeCount-Cloud/services/firewall/internal"
	"google.golang.org/grpc"
)

func main() {
	// 解析命令行参数
	socketPath := flag.String("socket", "", "Unix domain socket path")
	flag.Parse()

	// 初始化防火墙服务
	firewallService := internal.NewFirewallService()

	// 配置防火墙规则
	firewallService.ConfigureFirewallRules(internal.FirewallConfig{
		DefaultAction: internal.Allow,
		Rules: []internal.FirewallRule{
			{
				IP:     "0.0.0.0/0",
				Action: internal.Allow,
			},
		},
	})

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册防火墙服务
	firewall.RegisterFirewallServiceServer(grpcServer, firewallService)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, firewallService)

	// 创建通信抽象层实例
	trans := transport.NewTransportWithFallback()

	// 确定服务地址
	address := *socketPath
	if address == "" {
		// 使用默认地址
		address = trans.DefaultAddress("firewall")
	}

	// 创建监听器
	listener, err := trans.NewListener(address)
	if err != nil {
		log.Printf("Failed to create listener: %v", err)
		log.Printf("Falling back to TCP port...")
		// 降级到使用网络端口
		port := 50056
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Printf("FirewallService is running on port %d", port)
	} else {
		log.Printf("FirewallService is running on %s", address)
	}

	// 启动gRPC服务器
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down FirewallService...")
	grpcServer.GracefulStop()
	log.Println("FirewallService exited")
}
