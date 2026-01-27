package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	configproto "github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	configmgr "github.com/fishdivinity/BeeCount-Cloud/services/config/internal/config"
	"google.golang.org/grpc"
)

func main() {
	// 输出提示信息，说明服务不能直接使用
	log.Println("ConfigService is running in standby mode.")
	log.Println("Please use BeeCount-Cloud service to start ConfigService via gRPC.")
	log.Println("Direct usage of this executable is not supported.")

	// 配置文件路径
	configPath := "d:/Work/code/BeeCount-Cloud/config/config.yaml"

	// 初始化配置管理器
	configManager := configmgr.NewConfigManager(configPath)

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册配置服务
	configproto.RegisterConfigServiceServer(grpcServer, configManager)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, configManager)

	// 监听端口
	port := 50051
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("ConfigService standby server is running on port %d", port)
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
