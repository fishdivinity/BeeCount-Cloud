package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/i18n"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/version"
	"github.com/spf13/cobra"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use: "BeeCount-Cloud",
	Run: func(cmd *cobra.Command, args []string) {
		// 默认显示帮助信息
		cmd.Help()
	},
	// 禁用自动生成的命令（如 completion）
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Execute 执行根命令
func Execute() error {
	// 在执行命令前处理语言切换
	// 手动检查命令行参数中是否有--lang标志
	for i, arg := range os.Args[1:] {
		if arg == "--lang" && i+1 < len(os.Args[1:]) {
			// 找到--lang标志，获取下一个参数作为语言
			lang := os.Args[1:][i+1]
			i18n.SetLanguage(lang)
			// 重新设置命令翻译
			setCommandTranslations()
			break
		} else if strings.HasPrefix(arg, "--lang=") {
			// 处理--lang=zh-CN格式
			lang := strings.TrimPrefix(arg, "--lang=")
			i18n.SetLanguage(lang)
			// 重新设置命令翻译
			setCommandTranslations()
			break
		}
	}

	return rootCmd.Execute()
}

// 初始化根命令
func init() {
	// 初始化i18n
	initI18n()

	// 设置命令的翻译字段
	setCommandTranslations()

	// 设置版本信息
	rootCmd.Version = version.GetVersion()

	// 添加全局强制标志
	rootCmd.PersistentFlags().Bool("force", false, i18n.T("flag.force"))

	// 添加语言切换标志
	rootCmd.PersistentFlags().String("lang", i18n.GetLanguage(), i18n.T("flag.lang"))

	// 设置版本模板
	rootCmd.SetVersionTemplate("BeeCount Cloud {{.Version}}\n")

	// 自定义帮助信息
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    i18n.T("root.help.use"),
		Short:  i18n.T("root.help.short"),
		Long:   i18n.T("root.help.long"),
		Hidden: true,
	})

	// 自定义标志错误处理函数
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		// 提取未知标志名称
		errMsg := err.Error()
		var flagName string
		if strings.HasPrefix(errMsg, "unknown flag:") {
			flagName = strings.TrimSpace(strings.TrimPrefix(errMsg, "unknown flag:"))
		} else {
			flagName = errMsg
		}
		fmt.Printf("BeeCount-Cloud: unknown flag '%s'. See 'BeeCount-Cloud --help'.\n", flagName)
		return nil
	})

	// 解析语言标志前的钩子
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// 从命令行标志获取语言设置
		lang, _ := cmd.Flags().GetString("lang")
		if lang != "" {
			i18n.SetLanguage(lang)
			// 重新设置所有命令的翻译字段
			setCommandTranslations()
		}
	}

	// 添加子命令
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(healthCmd)
}

// 设置命令的翻译字段
func setCommandTranslations() {
	// 设置根命令的翻译字段
	rootCmd.Short = i18n.T("root.short")
	rootCmd.Long = i18n.T("root.long")

	// 设置start命令的翻译字段
	startCmd.Use = i18n.T("start.use")
	startCmd.Short = i18n.T("start.short")
	startCmd.Long = i18n.T("start.long")
	startCmd.Example = i18n.T("start.example")

	// 设置stop命令的翻译字段
	stopCmd.Use = i18n.T("stop.use")
	stopCmd.Short = i18n.T("stop.short")
	stopCmd.Long = i18n.T("stop.long")
	stopCmd.Example = i18n.T("stop.example")

	// 设置restart命令的翻译字段
	restartCmd.Use = i18n.T("restart.use")
	restartCmd.Short = i18n.T("restart.short")
	restartCmd.Long = i18n.T("restart.long")
	restartCmd.Example = i18n.T("restart.example")

	// 设置status命令的翻译字段
	statusCmd.Use = i18n.T("status.use")
	statusCmd.Short = i18n.T("status.short")
	statusCmd.Long = i18n.T("status.long")
	statusCmd.Example = i18n.T("status.example")

	// 设置health命令的翻译字段
	healthCmd.Use = i18n.T("health.use")
	healthCmd.Short = i18n.T("health.short")
	healthCmd.Long = i18n.T("health.long")
	healthCmd.Example = i18n.T("health.example")
}

// 初始化i18n
func initI18n() {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	// 获取可执行文件所在目录
	execDir := filepath.Dir(execPath)

	// 尝试在可执行文件所在目录下查找i18n文件夹
	i18nDir := filepath.Join(execDir, "i18n")
	if _, err := os.Stat(i18nDir); err == nil {
		// 目录存在，尝试加载翻译文件
		if err := i18n.LoadTranslations(i18nDir); err == nil {
			// 加载成功，直接返回
			return
		}
	}

	// 如果在可执行文件所在目录没有找到或加载失败，尝试从父目录查找
	// 获取可执行文件所在目录的父目录
	parentDir := filepath.Dir(execDir)
	i18nDir = filepath.Join(parentDir, "i18n")
	if _, err := os.Stat(i18nDir); err == nil {
		// 目录存在，尝试加载翻译文件
		i18n.LoadTranslations(i18nDir)
	}
}
