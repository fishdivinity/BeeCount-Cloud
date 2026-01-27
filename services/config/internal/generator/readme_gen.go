package generator

import (
	"os"
	"path/filepath"
)

// GenerateREADMEFiles 生成README文件
// 在指定目录生成README.txt和readme_zh.txt文件
func GenerateREADMEFiles(configDir string) error {
	// README.txt内容
	readmeEN := "# BeeCount Cloud Configuration Service\n\n" +
		"## Overview\n" +
		"This service manages configuration for the BeeCount Cloud application. It provides a centralized way to handle configuration files and environment variables.\n\n" +
		"## Configuration Files\n\n" +
		"The configuration service generates and manages the following YAML files:\n\n" +
		"- server.yaml: Server-related configuration (port, mode, timeout, etc.)\n" +
		"- database.yaml: Database configuration (type, connection details, pool settings)\n" +
		"- storage.yaml: Storage configuration (local or S3-compatible storage)\n" +
		"- jwt.yaml: JWT authentication configuration\n" +
		"- log.yaml: Logging configuration (level, format, output)\n" +
		"- cors.yaml: CORS configuration\n\n" +
		"## Usage\n\n" +
		"### Starting the Service\n" +
		"# Start the configuration service\n" +
		"./beecount-config\n\n" +
		"### Configuration Management\n\n" +
		"1. File-based Configuration: Edit the YAML files directly in the config directory\n" +
		"2. Environment Variables: Set environment variables to override configuration files\n" +
		"3. gRPC API: Use the gRPC API to manage configuration dynamically\n\n" +
		"## Environment Variables\n\n" +
		"The following environment variables can be used to override configuration:\n\n" +
		"- CONFIG_PATH: Path to the configuration directory\n" +
		"- DATABASE_TYPE: Active database type (sqlite, mysql, postgres)\n" +
		"- STORAGE_TYPE: Active storage type (local, s3)\n" +
		"- ADMIN_PASSWORD: Admin account password\n" +
		"- SERVER_PORT: Server listening port\n\n" +
		"## License\n\n" +
		"MIT\n"

	// readme_zh.txt内容
	readmeZH := "# BeeCount Cloud 配置服务\n\n" +
		"## 概述\n" +
		"本服务用于管理 BeeCount Cloud 应用的配置，提供了集中式的方式来处理配置文件和环境变量。\n\n" +
		"## 配置文件\n\n" +
		"配置服务生成和管理以下 YAML 文件：\n\n" +
		"- server.yaml: 服务器相关配置（端口、运行模式、超时等）\n" +
		"- database.yaml: 数据库配置（类型、连接详情、连接池设置）\n" +
		"- storage.yaml: 存储配置（本地存储或 S3 兼容存储）\n" +
		"- jwt.yaml: JWT 认证配置\n" +
		"- log.yaml: 日志配置（级别、格式、输出方式）\n" +
		"- cors.yaml: CORS 配置\n\n" +
		"## 使用方法\n\n" +
		"### 启动服务\n" +
		"# 启动配置服务\n" +
		"./beecount-config\n\n" +
		"### 配置管理\n\n" +
		"1. 基于文件的配置：直接编辑 config 目录下的 YAML 文件\n" +
		"2. 环境变量：设置环境变量来覆盖配置文件\n" +
		"3. gRPC API：使用 gRPC API 动态管理配置\n\n" +
		"## 环境变量\n\n" +
		"以下环境变量可用于覆盖配置：\n\n" +
		"- CONFIG_PATH: 配置目录路径\n" +
		"- DATABASE_TYPE: 活动数据库类型（sqlite, mysql, postgres）\n" +
		"- STORAGE_TYPE: 活动存储类型（local, s3）\n" +
		"- ADMIN_PASSWORD: 管理员账户密码\n" +
		"- SERVER_PORT: 服务器监听端口\n\n" +
		"## 许可证\n\n" +
		"MIT\n"

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 生成 README.txt
	readmePathEN := filepath.Join(configDir, "README.txt")
	if _, err := os.Stat(readmePathEN); os.IsNotExist(err) {
		if err := os.WriteFile(readmePathEN, []byte(readmeEN), 0644); err != nil {
			return err
		}
	}

	// 生成 readme_zh.txt
	readmePathZH := filepath.Join(configDir, "readme_zh.txt")
	if _, err := os.Stat(readmePathZH); os.IsNotExist(err) {
		if err := os.WriteFile(readmePathZH, []byte(readmeZH), 0644); err != nil {
			return err
		}
	}

	return nil
}