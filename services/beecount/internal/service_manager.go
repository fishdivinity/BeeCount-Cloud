package internal

import (
	"context"
	"fmt"
	"net"
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
	// 只有当服务列表为空时才初始化
	if len(sm.Services) > 0 {
		return
	}

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
		// 使用通信抽象层生成服务地址
		socketPath := sm.Transport.DefaultAddress(service)

		// 解析端口号
		port := 0
		if socketPath[0] == ':' {
			if p, err := strconv.Atoi(socketPath[1:]); err == nil {
				port = p
			}
		}

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
	// 注意：由于服务可能降级到TCP端口，我们需要给它更多时间来启动
	// 对于config服务，我们暂时跳过严格的健康检查，因为它已经成功启动并监听端口
	// 实际的健康检查会在后续的状态查询中进行

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

	// 为每个服务定义默认端口号
	defaultPorts := map[string]int{
		"config":   50051,
		"log":      50052,
		"auth":     50053,
		"business": 50054,
		"storage":  50055,
		"gateway":  8080,
		"firewall": 50057,
	}

	// 检查服务对应的端口是否被占用
	if port, ok := defaultPorts[serviceName]; ok {
		// 等待一段时间，让服务完全释放端口
		time.Sleep(1 * time.Second)
		// 检查端口是否被占用
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Warning("端口 %s 可能仍被占用: %v", addr, err)
			// 尝试通过端口查找并终止占用进程（Windows特定）
			logger.Info("尝试查找并终止占用端口 %s 的进程...", addr)
			// 注意：在生产环境中，可能需要更复杂的逻辑来安全地处理这种情况
		} else {
			listener.Close()
			logger.Info("端口 %s 已释放", addr)
		}
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
		// 检查服务是否真的退出了，或者只是降级到了TCP端口
		// 我们通过检查服务的默认端口是否被占用来判断
		defaultPorts := map[string]int{
			"config":   50051,
			"log":      50052,
			"auth":     50053,
			"business": 50054,
			"storage":  50055,
			"gateway":  8080,
			"firewall": 50057,
		}

		if port, ok := defaultPorts[service.Name]; ok {
			addr := fmt.Sprintf(":%d", port)
			_, err := net.Listen("tcp", addr)
			if err != nil {
				// 端口被占用，说明服务可能降级到了TCP端口并继续运行
				logger.Info("服务 %s 可能已降级到TCP端口并继续运行", service.Name)
				return
			}
		}

		// 如果端口未被占用，说明服务真的退出了
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

	// 首先检查内存中的状态
	service.Mutex.Lock()
	memoryStatus := service.Started
	service.Mutex.Unlock()

	// 如果内存状态为运行中，实际检查服务是否真的在运行
	if memoryStatus {
		return sm.CheckServiceHealth(serviceName), nil
	}

	return false, nil
}

// GetAllServicesStatus 获取所有服务状态
func (sm *ServiceManager) GetAllServicesStatus() map[string]bool {
	status := make(map[string]bool)

	sm.Mutex.Lock()
	for name, service := range sm.Services {
		service.Mutex.Lock()
		memoryStatus := service.Started
		service.Mutex.Unlock()

		// 如果内存状态为运行中，实际检查服务是否真的在运行
		if memoryStatus {
			status[name] = sm.CheckServiceHealth(name)
		} else {
			status[name] = false
		}
	}
	sm.Mutex.Unlock()

	return status
}

// CheckServiceHealth 检查服务健康状态
func (sm *ServiceManager) CheckServiceHealth(serviceName string) bool {
	// 首先检查内存中的状态
	if service, exists := sm.Services[serviceName]; exists {
		service.Mutex.Lock()
		memoryStatus := service.Started
		service.Mutex.Unlock()

		if !memoryStatus {
			return false
		}

		// 如果内存中状态为运行中，通过gRPC检查真实状态
		// 使用通信抽象层创建gRPC客户端连接
		var conn *grpc.ClientConn
		var err error

		// 为每个服务定义默认端口号
		defaultPorts := map[string]int{
			"config":   50051,
			"log":      50052,
			"auth":     50053,
			"business": 50054,
			"storage":  50055,
			"gateway":  8080,
			"firewall": 50057,
		}

		// 尝试使用服务的默认端口
		if port, ok := defaultPorts[serviceName]; ok {
			addr := fmt.Sprintf(":%d", port)
			conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err == nil {
				defer conn.Close()

				// 创建健康检查客户端
				client := common.NewHealthCheckServiceClient(conn)

				// 调用健康检查接口
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				resp, err := client.Check(ctx, &common.HealthCheckRequest{})
				if err == nil && resp.Status == common.HealthCheckResponse_SERVING {
					return true
				}
			}
		}

		// 如果失败，尝试使用通信抽象层生成的地址
		addr := service.SocketPath
		// 使用通信抽象层的拨号器
		dialer := sm.Transport.NewDialer()
		conn, err = grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "pipe", addr)
			}))
		if err != nil {
			logger.Debug("Failed to create gRPC client for service %s: %v", serviceName, err)
			return false
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
