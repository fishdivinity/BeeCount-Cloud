package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	logpb "github.com/fishdivinity/BeeCount-Cloud/common/proto/log"
	"github.com/fishdivinity/BeeCount-Cloud/services/log/internal"
	"google.golang.org/grpc"
)

func main() {
	// 初始化日志服务
	logService := internal.NewLogService()

	// 配置日志服务
	if err := logService.Configure(internal.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
		FileConfig: internal.FileConfig{
			Path:           "./logs/app.log",
			MaxSize:        10,
			MaxBackups:     100,
			MaxAge:         28,
			Compress:       true,
			MaxTotalSizeGB: 1,
		},
	}); err != nil {
		log.Fatalf("Failed to configure log service: %v", err)
	}

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册日志服务
	logpb.RegisterLogServiceServer(grpcServer, logService)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, logService)

	// 监听端口
	port := 50052
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("LogService is running on port %d", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down LogService...")
	grpcServer.GracefulStop()
	log.Println("LogService exited")
}
