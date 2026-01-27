package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm 二次确认函数
// 返回true表示确认，false表示取消
// defaultValue为true表示默认y，false表示默认n
func Confirm(message string, force bool, defaultValue bool) bool {
	if force {
		return true
	}

	for {
		// 根据默认值生成提示格式
		promptFormat := "y/N"
		if defaultValue {
			promptFormat = "Y/n"
		}

		fmt.Printf("%s (%s): ", message, promptFormat)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')

		// 去除前后空白字符
		response = strings.TrimSpace(response)

		// 处理空输入，使用默认值
		if response == "" {
			defaultText := "n"
			if defaultValue {
				defaultText = "y"
			}
			fmt.Printf("Using default value: %s\n", defaultText)
			return defaultValue
		}

		// 处理有效输入
		lowerResponse := strings.ToLower(response)
		switch lowerResponse {
		case "y":
			return true
		case "n":
			return false
		}

		// 无效输入，重复询问
		fmt.Println("Invalid input, please enter y or n")
	}
}
