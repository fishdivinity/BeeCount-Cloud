package watcher

import (
	"log"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/loader"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/sync"
	"github.com/fsnotify/fsnotify"
)

// FileWatcher 文件监听器
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	configPath string
	onChange   func(*model.Config)
}

// NewFileWatcher 创建文件监听器
func NewFileWatcher(configPath string, onChange func(*model.Config)) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// 监听配置文件
	if err := watcher.Add(configPath); err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    watcher,
		configPath: configPath,
		onChange:   onChange,
	}, nil
}

// Start 启动文件监听器
func (fw *FileWatcher) Start() {
	go func() {
		defer fw.watcher.Close()

		var debounceTimer *time.Timer
		const debounceDuration = 500 * time.Millisecond

		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}

				// 防抖处理，避免短时间内多次触发
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(debounceDuration, func() {
					// 只处理写事件和创建事件
					if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
						log.Printf("Config file changed: %s", event.Name)
						// 重新加载配置
						cfg, _, err := loader.LoadConfig(fw.configPath)
						if err != nil {
							log.Printf("Failed to reload config: %v", err)
							return
						}

						// 检查配置完整性
						cfg = sync.CheckConfigIntegrity(cfg)

						// 同步配置到环境变量
						if err := sync.SyncConfig(cfg, model.ConfigSourceFile, fw.configPath); err != nil {
							log.Printf("Failed to sync config: %v", err)
							return
						}

						// 触发配置变更回调
						if fw.onChange != nil {
							fw.onChange(cfg)
						}
					}
				})
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("File watcher error: %v", err)
			}
		}
	}()
}

// Stop 停止文件监听器
func (fw *FileWatcher) Stop() error {
	return fw.watcher.Close()
}
