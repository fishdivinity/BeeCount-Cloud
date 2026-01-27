package watcher

import (
	"log"
	"os"
	"time"
)

// EnvWatcher 环境变量监听器
type EnvWatcher struct {
	interval   time.Duration
	lastEnv    map[string]string
	onChange   func(map[string]string)
	stopCh     chan struct{}
	stopCalled bool
}

// NewEnvWatcher 创建环境变量监听器
func NewEnvWatcher(interval time.Duration, onChange func(map[string]string)) *EnvWatcher {
	return &EnvWatcher{
		interval: interval,
		lastEnv:  getCurrentEnv(),
		onChange: onChange,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动环境变量监听器
func (ew *EnvWatcher) Start() {
	go func() {
		ticker := time.NewTicker(ew.interval)
		defer func() {
			ticker.Stop()
			close(ew.stopCh)
		}()

		for {
			select {
			case <-ticker.C:
				currentEnv := getCurrentEnv()
				// 检查环境变量是否有变化
				if !compareEnvMaps(ew.lastEnv, currentEnv) {
					log.Println("Environment variables changed")
					ew.lastEnv = currentEnv
					if ew.onChange != nil {
						ew.onChange(currentEnv)
					}
				}
			case <-ew.stopCh:
				return
			}
		}
	}()
}

// Stop 停止环境变量监听器
func (ew *EnvWatcher) Stop() {
	if !ew.stopCalled {
		ew.stopCalled = true
		ew.stopCh <- struct{}{}
	}
}

// getCurrentEnv 获取当前环境变量
func getCurrentEnv() map[string]string {
	envMap := make(map[string]string)
	// 只检查关键配置相关的环境变量
	for _, key := range ConfigEnvVars {
		envMap[key] = os.Getenv(key)
	}

	return envMap
}

// compareEnvMaps 比较两个环境变量映射是否相同
func compareEnvMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}
