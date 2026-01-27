package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/storage"
	"github.com/fishdivinity/BeeCount-Cloud/services/storage/internal"
	"google.golang.org/grpc"
)

func main() {
	// 初始化存储服务
	storageService := internal.NewStorageService()

	// 配置本地存储
	if err := storageService.ConfigureLocalStorage(internal.LocalStorageConfig{
		Path:      "./data/uploads",
		URLPrefix: "/uploads",
	}); err != nil {
		log.Fatalf("Failed to configure local storage: %v", err)
	}

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册存储服务
	storage.RegisterStorageServiceServer(grpcServer, storageService)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, storageService)

	// 监听端口
	port := 50055
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("StorageService is running on port %d", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down StorageService...")
	grpcServer.GracefulStop()
	log.Println("StorageService exited")
}
