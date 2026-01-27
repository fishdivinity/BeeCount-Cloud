package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/transport"
	"github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service 服务定义
type Service struct {
	Name       string
	Path       string
	Port       int
	SocketPath string
	Cmd        *exec.Cmd
	Started    bool
	PidFile    string
	LogFile    string
	Mutex      sync.Mutex
}

// ServiceManager 服务管理器
type ServiceManager struct {
	Services  map[string]*Service
	Transport transport.Transport
	Mutex     sync.Mutex
}

// NewServiceManager 创建服务管理器实例
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		Services:  make(map[string]*Service),
		Transport: transport.NewTransportWithFallback(),
		Mutex:     sync.Mutex{},
	}
}

// InitServices 初始化服务列表
func (sm *ServiceManager) InitServices() {
	// 获取项目根目录
	var rootDir string

	// 尝试通过当前工作目录获取
	currentDir, _ := os.Getwd()

	// 检查当前目录结构，寻找项目根目录的特征
	// 项目根目录应该包含 services 目录
	if _, err := os.Stat(filepath.Join(currentDir, "services")); err == nil {
		// 当前目录就是项目根目录
		rootDir = currentDir
	} else if _, err := os.Stat(filepath.Join(currentDir, "..", "services")); err == nil {
		// 向上一级是项目根目录
		rootDir = filepath.Dir(currentDir)
	} else if _, err := os.Stat(filepath.Join(currentDir, "..", "..", "services")); err == nil {
		// 向上两级是项目根目录
		rootDir = filepath.Dir(filepath.Dir(currentDir))
	} else {
		// 如果上述方法都失败，尝试使用可执行文件路径
		execPath, err := os.Executable()
		if err != nil {
			// 如果获取可执行文件路径失败，使用默认值
			rootDir = filepath.Dir(filepath.Dir(currentDir))
			logger.Warning("Failed to get executable path, using default root dir: %s", rootDir)
		} else {
			// 从可执行文件路径向上查找项目根目录
			execDir := filepath.Dir(execPath)
			if _, err := os.Stat(filepath.Join(execDir, "services")); err == nil {
				rootDir = execDir
			} else if _, err := os.Stat(filepath.Join(execDir, "..", "services")); err == nil {
				rootDir = filepath.Dir(execDir)
			} else if _, err := os.Stat(filepath.Join(execDir, "..", "..", "services")); err == nil {
				rootDir = filepath.Dir(filepath.Dir(execDir))
			} else {
				// 如果还是找不到，使用默认值
				rootDir = filepath.Dir(filepath.Dir(currentDir))
				logger.Warning("Failed to find project root from executable path, using default: %s", rootDir)
			}
		}
	}

	logger.Info("Using project root directory: %s", rootDir)

	// 创建日志和PID目录
	logsDir := filepath.Join(rootDir, "logs")
	pidsDir := filepath.Join(rootDir, "pids")
	os.MkdirAll(logsDir, 0755)
	os.MkdirAll(pidsDir, 0755)

	// 服务列表
	services := []string{
		"gateway",
		"config",
		"auth",
		"business",
		"storage",
		"log",
		"firewall",
	}

	// 初始化每个服务
	for _, service := range services {
		// 为每个服务分配端口
		port := 50050
		switch service {
		case "config":
			port = 50051
		case "auth":
			port = 50052
		case "business":
			port = 50053
		case "storage":
			port = 50054
		case "log":
			port = 50055
		case "firewall":
			port = 50056
		case "gateway":
			port = 50057
		}

		// 使用通信抽象层生成服务地址
		socketPath := sm.Transport.DefaultAddress(service)

		servicePath := filepath.Join(rootDir, "services", service, "cmd")
		sm.Services[service] = &Service{
			Name:       service,
			Path:       servicePath,
			Port:       port,
			SocketPath: socketPath,
			Started:    false,
			PidFile:    filepath.Join(pidsDir, service+".pid"),
			LogFile:    filepath.Join(logsDir, service+".log"),
			Mutex:      sync.Mutex{},
		}
	}
}

// StartService 启动指定服务
func (sm *ServiceManager) StartService(serviceName string, background bool) error {
	// 初始化服务管理器，确保服务列表已加载
	sm.InitServices()

	// 首先检查服务是否已启动
	status, _ := sm.GetServiceStatus(serviceName)
	if status {
		logger.Info("服务 %s 已启动", serviceName)
		return nil
	}

	// 对于所有服务，使用统一的启动逻辑
	// 启动服务的可执行文件
	if err := sm.startServiceExecutable(serviceName, background); err != nil {
		logger.Error("启动服务可执行文件失败: %v", err)
		return err
	}

	// 等待一段时间，让服务启动
	time.Sleep(2 * time.Second)

	// 检查服务是否正常运行
	if serviceName == "config" {
		// 对于config服务，通过健康检查验证服务是否正常运行
		if !sm.CheckServiceHealth(serviceName) {
			logger.Error("config服务启动失败，请检查日志")
			return fmt.Errorf("config服务启动失败")
		}
	}

	logger.Info("服务 %s 已启动", serviceName)
	return nil
}

// startServiceExecutable 启动服务的可执行文件
func (sm *ServiceManager) startServiceExecutable(serviceName string, background bool) error {
	sm.Mutex.Lock()
	service, exists := sm.Services[serviceName]
	sm.Mutex.Unlock()

	if !exists {
		return fmt.Errorf("服务 %s 不存在", serviceName)
	}

	// 检查服务是否已启动
	service.Mutex.Lock()
	if service.Started {
		service.Mutex.Unlock()
		logger.Info("服务 %s 已启动", serviceName)
		return nil
	}

	// 使用服务管理器初始化时设置的服务路径
	servicePath := service.Path

	// 启动服务，先尝试使用 Unix 域套接字
	cmd := exec.Command("go", "run", ".", "--socket", service.SocketPath)
	cmd.Dir = servicePath

	if background {
		// 后台运行模式
		// 创建日志文件
		logFile, err := os.OpenFile(service.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			service.Mutex.Unlock()
			return fmt.Errorf("创建日志文件失败: %v", err)
		}
		defer logFile.Close()

		// 重定向输出
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Stdin = nil
	} else {
		// 前台运行模式
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		service.Mutex.Unlock()
		return fmt.Errorf("启动服务 %s 失败: %v", serviceName, err)
	}

	// 更新服务状态
	service.Cmd = cmd
	service.Started = true

	if background {
		// 写入PID文件
		pidFile, err := os.Create(service.PidFile)
		if err != nil {
			service.Mutex.Unlock()
			cmd.Process.Kill()
			return fmt.Errorf("创建PID文件失败: %v", err)
		}
		pidFile.WriteString(strconv.Itoa(cmd.Process.Pid))
		pidFile.Close()
	}

	service.Mutex.Unlock()

	logger.Info("服务 %s 可执行文件已启动 (PID: %d)", serviceName, cmd.Process.Pid)

	// 启动监控协程
	go sm.monitorService(service)

	return nil
}

// StopService 停止指定服务
func (sm *ServiceManager) StopService(serviceName string) error {
	sm.Mutex.Lock()
	service, exists := sm.Services[serviceName]
	sm.Mutex.Unlock()

	if !exists {
		return fmt.Errorf("服务 %s 不存在", serviceName)
	}

	// 检查服务是否已停止
	service.Mutex.Lock()
	if !service.Started {
		service.Mutex.Unlock()
		logger.Info("服务 %s 已停止", serviceName)
		return nil
	}

	// 停止服务
	if err := service.Cmd.Process.Kill(); err != nil {
		service.Mutex.Unlock()
		return fmt.Errorf("停止服务 %s 失败: %v", serviceName, err)
	}

	// 等待服务退出
	if err := service.Cmd.Wait(); err != nil {
		// 忽略退出错误，因为我们是强制终止的
	}

	// 更新服务状态
	service.Started = false
	service.Cmd = nil
	// 清理PID文件
	os.Remove(service.PidFile)
	service.Mutex.Unlock()

	logger.Info("服务 %s 已停止", serviceName)

	return nil
}

// StartAllServices 启动所有服务
func (sm *ServiceManager) StartAllServices(background bool) {
	// 初始化服务列表
	sm.InitServices()

	// 服务启动顺序（依赖关系）
	startOrder := []string{
		"config",
		"log",
		"auth",
		"business",
		"storage",
		"firewall",
		"gateway",
	}

	// 启动所有服务
	for _, serviceName := range startOrder {
		if err := sm.StartService(serviceName, background); err != nil {
			logger.Error("启动服务 %s 失败: %v", serviceName, err)
			continue
		}
		// 等待服务启动
		time.Sleep(1 * time.Second)
	}
}

// StopAllServices 停止所有服务
func (sm *ServiceManager) StopAllServices() {
	// 服务停止顺序（与启动顺序相反）
	stopOrder := []string{
		"gateway",
		"firewall",
		"storage",
		"business",
		"auth",
		"log",
		"config",
	}

	// 停止所有服务
	for _, serviceName := range stopOrder {
		if err := sm.StopService(serviceName); err != nil {
			logger.Error("停止服务 %s 失败: %v", serviceName, err)
			continue
		}
	}
}

// monitorService 监控服务运行状态
func (sm *ServiceManager) monitorService(service *Service) {
	// 等待服务退出
	if err := service.Cmd.Wait(); err != nil {
		service.Mutex.Lock()
		service.Started = false
		service.Cmd = nil
		// 清理PID文件
		os.Remove(service.PidFile)
		service.Mutex.Unlock()
		logger.Error("服务 %s 意外退出: %v", service.Name, err)
	}
}

// GetServiceStatus 获取服务状态
func (sm *ServiceManager) GetServiceStatus(serviceName string) (bool, error) {
	sm.Mutex.Lock()
	service, exists := sm.Services[serviceName]
	sm.Mutex.Unlock()

	if !exists {
		return false, fmt.Errorf("服务 %s 不存在", serviceName)
	}

	service.Mutex.Lock()
	status := service.Started
	service.Mutex.Unlock()

	return status, nil
}

// GetAllServicesStatus 获取所有服务状态
func (sm *ServiceManager) GetAllServicesStatus() map[string]bool {
	status := make(map[string]bool)

	sm.Mutex.Lock()
	for name, service := range sm.Services {
		service.Mutex.Lock()
		status[name] = service.Started
		service.Mutex.Unlock()
	}
	sm.Mutex.Unlock()

	return status
}

// CheckServiceHealth 检查服务健康状态
func (sm *ServiceManager) CheckServiceHealth(serviceName string) bool {
	// 首先检查内存中的状态
	status, _ := sm.GetServiceStatus(serviceName)
	if !status {
		return false
	}

	// 如果内存中状态为运行中，通过gRPC检查真实状态
	if service, exists := sm.Services[serviceName]; exists {
		// 使用通信抽象层创建gRPC客户端连接
		var conn *grpc.ClientConn
		var err error

		// 优先使用通信抽象层生成的地址
		addr := service.SocketPath

		// 根据地址类型选择不同的连接方式
		if sm.Transport.ValidateAddress(addr) {
			// 尝试使用通信抽象层的连接方式
			conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logger.Debug("Failed to create gRPC client for service %s: %v", serviceName, err)
				// 如果失败，尝试使用网络端口
				addr := fmt.Sprintf(":%d", service.Port)
				conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					logger.Debug("Failed to create TCP gRPC client for service %s: %v", serviceName, err)
					return false
				}
			}
		} else {
			// 使用网络端口作为备选
			addr := fmt.Sprintf(":%d", service.Port)
			conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logger.Debug("Failed to create TCP gRPC client for service %s: %v", serviceName, err)
				return false
			}
		}
		defer conn.Close()

		// 创建健康检查客户端
		client := common.NewHealthCheckServiceClient(conn)

		// 调用健康检查接口
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := client.Check(ctx, &common.HealthCheckRequest{})
		if err != nil {
			logger.Debug("Health check failed for service %s: %v", serviceName, err)
			return false
		}

		// 返回健康状态
		return resp.Status == common.HealthCheckResponse_SERVING
	}

	return false
}
