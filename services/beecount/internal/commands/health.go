package commands

import (
	"fmt"

	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal"
	"github.com/spf13/cobra"
)

// healthCmd 健康检查命令
var healthCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化服务管理器
		serviceManager := internal.NewServiceManager()
		serviceManager.InitServices()

		if len(args) == 0 {
			// 检查所有服务健康状态
			fmt.Println("Service Health Status:")
			fmt.Println("----------------------")

			// 获取所有服务
			serviceManager.Mutex.Lock()
			services := serviceManager.Services
			serviceManager.Mutex.Unlock()

			for serviceName := range services {
				healthy := serviceManager.CheckServiceHealth(serviceName)
				statusText := "Unhealthy"
				if healthy {
					statusText = "Healthy"
				}
				fmt.Printf("%10s: %s\n", serviceName, statusText)
			}
		} else {
			// 检查指定服务健康状态
			serviceName := args[0]
			healthy := serviceManager.CheckServiceHealth(serviceName)
			statusText := "Unhealthy"
			if healthy {
				statusText = "Healthy"
			}
			fmt.Printf("Service %s health: %s\n", serviceName, statusText)
		}
	},
}
