package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/firewall"
	"github.com/fishdivinity/BeeCount-Cloud/services/firewall/internal"
	"google.golang.org/grpc"
)

func main() {
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

	// 监听端口
	port := 50056
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("FirewallService is running on port %d", port)
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
