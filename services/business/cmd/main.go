package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/business"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/services/business/internal"
	"google.golang.org/grpc"
)

func main() {
	// 初始化业务服务
	businessService := internal.NewBusinessService()

	// 配置数据库（SQLite3）
	if err := businessService.ConfigureDatabase(internal.DatabaseConfig{
		Type: "sqlite3",
		SQLiteConfig: internal.SQLiteConfig{
			Path: "./data/beecount.db",
		},
	}); err != nil {
		log.Fatalf("Failed to configure database: %v", err)
	}

	// 初始化数据库
	if err := businessService.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册业务服务
	business.RegisterBusinessServiceServer(grpcServer, businessService)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, businessService)

	// 监听端口
	port := 50054
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("BusinessService is running on port %d", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down BusinessService...")
	grpcServer.GracefulStop()
	log.Println("BusinessService exited")
}
