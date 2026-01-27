package main

import (
	"os"

	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/internal/commands"
)

func main() {
	// 执行命令，让 Cobra 直接处理错误
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
