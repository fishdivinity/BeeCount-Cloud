package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/gateway/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	_ = flag.String("socket", "", "Unix domain socket path")
	flag.Parse()

	// 初始化API网关
	gateway := internal.NewAPIGateway()

	// 配置gRPC客户端，优先使用 Unix 域套接字
	grpcConfig := internal.GRPCClientConfig{
		AuthServiceAddr:     "localhost:50053",
		BusinessServiceAddr: "localhost:50054",
		StorageServiceAddr:  "localhost:50055",
		ConfigServiceAddr:   "localhost:50051",
		LogServiceAddr:      "localhost:50052",
	}

	// 使用默认的 Unix 域套接字路径
	grpcConfig.AuthServiceAddr = getSocketPath("auth")
	grpcConfig.BusinessServiceAddr = getSocketPath("business")
	grpcConfig.StorageServiceAddr = getSocketPath("storage")
	grpcConfig.ConfigServiceAddr = getSocketPath("config")
	grpcConfig.LogServiceAddr = getSocketPath("log")

	if err := gateway.ConfigureGRPCClients(grpcConfig); err != nil {
		log.Fatalf("Failed to configure gRPC clients: %v", err)
	}

	// 初始化路由
	router := gin.Default()
	gateway.SetupRoutes(router)

	// 配置服务器
	port := 8080
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("API Gateway is running on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API Gateway: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("API Gateway forced to shutdown: %v", err)
	}

	log.Println("API Gateway exited")
}

// getSocketPath 生成跨平台的 Unix 域套接字路径
func getSocketPath(serviceName string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("\\\\.\\pipe\\beecount_%s", serviceName)
	} else {
		return fmt.Sprintf("/tmp/beecount_%s.sock", serviceName)
	}
}
