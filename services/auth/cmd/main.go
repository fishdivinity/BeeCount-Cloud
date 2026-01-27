package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/auth"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/services/auth/internal"
	"google.golang.org/grpc"
)

func main() {
	// 初始化认证服务
	authService := internal.NewAuthService()

	// 配置JWT
	if err := authService.ConfigureJWT(internal.JWTConfig{
		Secret:               "your-secret-key",
		ExpireHours:          24,
		RotationIntervalDays: 7,
	}); err != nil {
		log.Fatalf("Failed to configure JWT: %v", err)
	}

	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册认证服务
	auth.RegisterAuthServiceServer(grpcServer, authService)

	// 注册健康检查服务
	common.RegisterHealthCheckServiceServer(grpcServer, authService)

	// 监听端口
	port := 50053
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 启动gRPC服务器
	go func() {
		log.Printf("AuthService is running on port %d", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down AuthService...")
	grpcServer.GracefulStop()
	log.Println("AuthService exited")
}
