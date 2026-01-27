package commands

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/i18n"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/logger"
	"github.com/spf13/cobra"
)

// stopCmd 停止命令
var stopCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// 获取标志值
		allServices, _ := cmd.Flags().GetBool("all")
		force, _ := cmd.Flags().GetBool("force")

		// 初始化服务管理器
		serviceManager := internal.NewServiceManager()
		serviceManager.InitServices()

		// 二次确认
		var confirmMessage string
		if allServices || len(args) == 0 {
			confirmMessage = "Are you sure you want to stop all services?"
		} else {
			confirmMessage = "Are you sure you want to stop service " + args[0] + "?"
		}

		if !Confirm(confirmMessage, force, false) {
			logger.Info("Operation canceled.")
			return
		}

		// 停止服务
		if allServices || len(args) == 0 {
			// 停止所有服务
			logger.Info("Stopping all services...")
			serviceManager.StopAllServices()
		} else {
			// 停止指定服务
			serviceName := args[0]
			logger.Info("Stopping service %s...", serviceName)
			if err := serviceManager.StopService(serviceName); err != nil {
				logger.Error("Failed to stop service %s: %v", serviceName, err)
				return
			}
		}
	},
}

func init() {
	// 添加停止命令标志
	stopCmd.Flags().Bool("all", false, i18n.T("flag.all"))
}
