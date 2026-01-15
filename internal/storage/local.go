package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
)

// LocalStorage 本地存储实现
// 使用本地文件系统存储文件
type LocalStorage struct {
	basePath  string
	urlPrefix string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(cfg *config.LocalConfig) Storage {
	return &LocalStorage{
		basePath:  cfg.Path,
		urlPrefix: cfg.URLPrefix,
	}
}

// Upload 上传文件到本地存储
func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, metadata *Metadata) error {
	fullPath := filepath.Join(s.basePath, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download 从本地存储下载文件
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// DownloadRange 下载文件的指定范围
func (s *LocalStorage) DownloadRange(ctx context.Context, key string, offset int64, length int64) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	if length > 0 {
		return &limitedReadCloser{ReadCloser: file, limit: length}, nil
	}

	return file, nil
}

// limitedReadCloser 限制读取大小的ReadCloser
type limitedReadCloser struct {
	io.ReadCloser
	limit int64
	read  int64
}

func (l *limitedReadCloser) Read(p []byte) (n int, err error) {
	if l.read >= l.limit {
		return 0, io.EOF
	}

	remaining := l.limit - l.read
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err = l.ReadCloser.Read(p)
	l.read += int64(n)
	return n, err
}

// Delete 从本地存储删除文件
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL 获取本地文件的访问URL
func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.urlPrefix, "/"), key), nil
}

// Exists 检查本地文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *LocalStorage) GetSize(ctx context.Context, key string) (int64, error) {
	fullPath := filepath.Join(s.basePath, key)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

// TestConnection 测试存储连接
func (s *LocalStorage) TestConnection(ctx context.Context) error {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	testFile := filepath.Join(s.basePath, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}

	return nil
}

// GetFileContentType 获取文件内容类型
// 根据文件扩展名或HTTP头判断文件类型
func GetFileContentType(file *multipart.FileHeader) string {
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		ext := filepath.Ext(file.Filename)
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		default:
			return "application/octet-stream"
		}
	}
	return contentType
}

