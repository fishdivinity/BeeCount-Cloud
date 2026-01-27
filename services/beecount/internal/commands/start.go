package commands

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/i18n"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/logger"
	"github.com/spf13/cobra"
)

// startCmd 启动命令
var startCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// 获取标志值
		allServices, _ := cmd.Flags().GetBool("all")
		background, _ := cmd.Flags().GetBool("background")

		// 初始化服务管理器
		serviceManager := internal.NewServiceManager()
		serviceManager.InitServices()

		// 启动服务
		if allServices || len(args) == 0 {
			// 启动所有服务
			logger.Info("Starting all services...")
			serviceManager.StartAllServices(background)
		} else {
			// 启动指定服务
			serviceName := args[0]
			logger.Info("Starting service %s...", serviceName)
			if err := serviceManager.StartService(serviceName, background); err != nil {
				logger.Error("Failed to start service %s: %v", serviceName, err)
				return
			}
		}

		if background {
			logger.Info("All services started in background.")
		}
	},
}

func init() {
	// 添加启动命令标志
	startCmd.Flags().Bool("all", false, i18n.T("flag.all"))
	startCmd.Flags().Bool("background", false, i18n.T("flag.background"))
	startCmd.Flags().BoolP("gateway", "g", false, i18n.T("flag.gateway"))
	startCmd.Flags().BoolP("config", "c", false, i18n.T("flag.config"))
	startCmd.Flags().BoolP("auth", "a", false, i18n.T("flag.auth"))
	startCmd.Flags().BoolP("business", "b", false, i18n.T("flag.business"))
	startCmd.Flags().BoolP("storage", "s", false, i18n.T("flag.storage"))
	startCmd.Flags().BoolP("log", "l", false, i18n.T("flag.log"))
	startCmd.Flags().BoolP("firewall", "f", false, i18n.T("flag.firewall"))
}
