package commands

import (
	"fmt"

	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal"
	"github.com/spf13/cobra"
)

// statusCmd 状态命令
var statusCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化服务管理器
		serviceManager := internal.NewServiceManager()
		serviceManager.InitServices()

		if len(args) == 0 {
			// 查看所有服务状态
			serviceManager.GetAllServicesStatus()
			fmt.Println("Service Status:")
			fmt.Println("----------------")
			for service := range serviceManager.Services {
				// 使用健康检查获取真实状态
				realStatus := serviceManager.CheckServiceHealth(service)
				statusText := "Stopped"
				if realStatus {
					statusText = "Running"
				}
				fmt.Printf("%10s: %s\n", service, statusText)
			}
		} else {
			// 查看指定服务状态
			serviceName := args[0]
			// 使用健康检查获取真实状态
			realStatus := serviceManager.CheckServiceHealth(serviceName)
			statusText := "Stopped"
			if realStatus {
				statusText = "Running"
			}
			fmt.Printf("Service %s status: %s\n", serviceName, statusText)
		}
	},
}
