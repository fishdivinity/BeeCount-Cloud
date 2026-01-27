package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// I18n 国际化管理器
type I18n struct {
	// 当前语言
	language string
	// 翻译映射
	translations map[string]map[string]string
	// 互斥锁，保证并发安全
	mu sync.RWMutex
}

// 单例实例
var (
	instance *I18n
	once     sync.Once
)

// GetInstance 获取i18n实例
func GetInstance() *I18n {
	once.Do(func() {
		// 从环境变量获取语言设置
		language := os.Getenv("BEECOUNT_LANG")
		if language == "" {
			language = "en-US" // 默认英文
		}

		// 初始化翻译映射，包含默认的英文翻译
		translations := make(map[string]map[string]string)
		translations["en-US"] = defaultEnTranslations

		instance = &I18n{
			language:     language,
			translations: translations,
		}
	})
	return instance
}

// LoadTranslations 加载翻译文件
func (i *I18n) LoadTranslations(basePath string) error {
	// 确保basePath存在
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil // 文件夹不存在，返回nil，使用默认英文
	}

	// 读取文件夹中的所有JSON文件
	files, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read i18n directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue
		}

		// 提取语言代码（文件名不带扩展名）
		langCode := strings.TrimSuffix(filename, ".json")

		// 读取文件内容
		filePath := filepath.Join(basePath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %v", filename, err)
		}

		// 解析JSON
		var trans map[string]string
		if err := json.Unmarshal(content, &trans); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %v", filename, err)
		}

		// 存储翻译
		i.mu.Lock()
		i.translations[langCode] = trans
		i.mu.Unlock()
	}

	return nil
}

// SetLanguage 设置当前语言
func (i *I18n) SetLanguage(lang string) {
	i.mu.Lock()
	i.language = lang
	i.mu.Unlock()
}

// GetLanguage 获取当前语言
func (i *I18n) GetLanguage() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.language
}

// T 获取翻译后的文本
// 如果指定语言不存在对应的翻译，则尝试使用默认语言（en-US）
// 如果默认语言也没有，则返回原始键
func (i *I18n) T(key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 尝试获取当前语言的翻译
	if trans, ok := i.translations[i.language]; ok {
		if value, ok := trans[key]; ok {
			return value
		}
	}

	// 尝试获取默认语言的翻译
	if trans, ok := i.translations["en-US"]; ok {
		if value, ok := trans[key]; ok {
			return value
		}
	}

	// 都没有找到，返回原始键
	return key
}

// MustT 获取翻译后的文本，如果没有找到则panic
func (i *I18n) MustT(key string) string {
	result := i.T(key)
	if result == key {
		panic(fmt.Sprintf("translation not found for key: %s", key))
	}
	return result
}

// 全局函数，方便使用

// LoadTranslations 加载翻译文件
func LoadTranslations(basePath string) error {
	return GetInstance().LoadTranslations(basePath)
}

// SetLanguage 设置当前语言
func SetLanguage(lang string) {
	GetInstance().SetLanguage(lang)
}

// GetLanguage 获取当前语言
func GetLanguage() string {
	return GetInstance().GetLanguage()
}

// T 获取翻译后的文本
func T(key string) string {
	return GetInstance().T(key)
}

// MustT 获取翻译后的文本，如果没有找到则panic
func MustT(key string) string {
	return GetInstance().MustT(key)
}
