package commands

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/i18n"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/logger"
	"github.com/spf13/cobra"
)

// restartCmd 重启命令
var restartCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// 获取标志值
		allServices, _ := cmd.Flags().GetBool("all")
		background, _ := cmd.Flags().GetBool("background")
		force, _ := cmd.Flags().GetBool("force")

		// 初始化服务管理器
		serviceManager := internal.NewServiceManager()
		serviceManager.InitServices()

		// 二次确认
		var confirmMessage string
		if allServices || len(args) == 0 {
			confirmMessage = "Are you sure you want to restart all services?"
		} else {
			confirmMessage = "Are you sure you want to restart service " + args[0] + "?"
		}

		if !Confirm(confirmMessage, force, false) {
			logger.Info("Operation canceled.")
			return
		}

		// 重启服务
		if allServices || len(args) == 0 {
			// 重启所有服务
			logger.Info("Restarting all services...")
			serviceManager.StopAllServices()
			serviceManager.StartAllServices(background)
		} else {
			// 重启指定服务
			serviceName := args[0]
			logger.Info("Restarting service %s...", serviceName)

			// 停止服务
			if err := serviceManager.StopService(serviceName); err != nil {
				logger.Error("Failed to stop service %s: %v", serviceName, err)
				return
			}

			// 启动服务
			if err := serviceManager.StartService(serviceName, background); err != nil {
				logger.Error("Failed to start service %s: %v", serviceName, err)
				return
			}
		}
	},
}

func init() {
	// 添加重启命令标志
	restartCmd.Flags().Bool("all", false, i18n.T("flag.all"))
	restartCmd.Flags().Bool("background", false, i18n.T("flag.background"))
}
