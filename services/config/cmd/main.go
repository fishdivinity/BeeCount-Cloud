package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	configproto "github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	"github.com/fishdivinity/BeeCount-Cloud/common/transport"
	configmgr "github.com/fishdivinity/BeeCount-Cloud/services/config/internal/config"
	"google.golang.org/grpc"
)

func main() {
	// 解析命令行参数
	socketPath := flag.String("socket", "", "Unix domain socket path")
	flag.Parse()

	// 配置文件路径
	configPath := "d:/Work/code/BeeCount-Cloud/config/config.yaml"

	// 输出启动信息
	log.Println("Starting ConfigService...")

	// 初始化配置管理器
	configManager := configmgr.NewConfigManager(configPath)

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册配置服务
	configproto.RegisterConfigServiceServer(grpcServer, configManager)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, configManager)

	// 创建通信抽象层实例
	trans := transport.NewTransportWithFallback()

	// 确定服务地址
	address := *socketPath
	if address == "" {
		// 使用默认地址
		address = trans.DefaultAddress("config")
	}

	// 创建监听器
	listener, err := trans.NewListener(address)
	if err != nil {
		log.Printf("Failed to create listener: %v", err)
		log.Printf("Falling back to TCP port...")
		// 降级到使用网络端口
		tcpAddr := ":50051" // 使用固定的TCP端口
		listener, err = net.Listen("tcp", tcpAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Printf("ConfigService is running on %s", tcpAddr)
	} else {
		log.Printf("ConfigService is running on %s", address)
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

	log.Println("Shutting down ConfigService standby server...")
	grpcServer.GracefulStop()
	log.Println("ConfigService standby server exited")
}
