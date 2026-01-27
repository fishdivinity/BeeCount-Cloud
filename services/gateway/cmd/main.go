package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/gateway/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化API网关
	gateway := internal.NewAPIGateway()

	// 配置gRPC客户端
	if err := gateway.ConfigureGRPCClients(internal.GRPCClientConfig{
		AuthServiceAddr:     "localhost:50053",
		BusinessServiceAddr: "localhost:50054",
		StorageServiceAddr:  "localhost:50055",
		ConfigServiceAddr:   "localhost:50051",
		LogServiceAddr:      "localhost:50052",
	}); err != nil {
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
