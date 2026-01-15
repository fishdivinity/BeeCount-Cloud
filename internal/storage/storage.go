package storage

import (
	"context"
	"io"
	"time"
)

// Storage 存储接口
// 定义对象存储的通用方法，支持多种存储后端
type Storage interface {
	// Upload 上传文件
	Upload(ctx context.Context, key string, reader io.Reader, metadata *Metadata) error
	// Download 下载文件
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// DownloadRange 下载文件的指定范围
	DownloadRange(ctx context.Context, key string, offset int64, length int64) (io.ReadCloser, error)
	// Delete 删除文件
	Delete(ctx context.Context, key string) error
	// GetURL 获取文件访问URL
	GetURL(ctx context.Context, key string) (string, error)
	// Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// GetSize 获取文件大小
	GetSize(ctx context.Context, key string) (int64, error)
	// TestConnection 测试存储连接
	TestConnection(ctx context.Context) error
}

// Metadata 文件元数据
type Metadata struct {
	ContentType string
	Size        int64
	Width       *int
	Height      *int
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}

// UploadSession 上传会话
type UploadSession struct {
	ID           string
	Key          string
	TotalSize    int64
	UploadedSize int64
	ChunkSize    int64
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

