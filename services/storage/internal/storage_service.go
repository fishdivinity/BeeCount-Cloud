package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	Path      string
	URLPrefix string
}

// FileInfo 文件信息模型
type FileInfo struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	UserID      string            `json:"user_id"`
	StoragePath string            `json:"storage_path"`
	CreatedAt   string            `json:"created_at"`
	Metadata    map[string]string `json:"metadata"`
}

// StorageService 存储服务实现
type StorageService struct {
	storage.UnimplementedStorageServiceServer
	common.UnimplementedHealthCheckServiceServer

	config LocalStorageConfig
	mu     sync.RWMutex
}

// NewStorageService 创建存储服务实例
func NewStorageService() *StorageService {
	return &StorageService{}
}

// ConfigureLocalStorage 配置本地存储
func (s *StorageService) ConfigureLocalStorage(config LocalStorageConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// 确保存储目录存在
	return os.MkdirAll(config.Path, 0755)
}

// UploadFile 上传文件（支持流式上传）
func (s *StorageService) UploadFile(stream storage.StorageService_UploadFileServer) error {
	// 初始化文件信息
	var fileInfo FileInfo
	var fileID string
	var file *os.File
	var totalSize int64
	var isFirstChunk = true

	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	// 接收文件流
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// 文件接收完成
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to receive file chunk: %v", err)
		}

		if isFirstChunk {
			// 第一次接收，初始化文件信息
			fileID = uuid.New().String()
			fileInfo = FileInfo{
				ID:          fileID,
				Filename:    req.Filename,
				ContentType: req.ContentType,
				UserID:      req.UserId,
				Metadata:    req.Metadata,
				CreatedAt:   time.Now().Format(time.RFC3339),
			}

			// 创建存储目录结构（按用户ID和日期）
			userDir := filepath.Join(s.config.Path, req.UserId)
			if err := os.MkdirAll(userDir, 0755); err != nil {
				return status.Errorf(codes.Internal, "Failed to create user directory: %v", err)
			}

			// 生成文件存储路径
			fileExt := filepath.Ext(req.Filename)
			storagePath := filepath.Join(userDir, fmt.Sprintf("%s%s", fileID, fileExt))
			fileInfo.StoragePath = storagePath

			// 创建文件
			file, err = os.Create(storagePath)
			if err != nil {
				return status.Errorf(codes.Internal, "Failed to create file: %v", err)
			}

			isFirstChunk = false
		}

		// 写入文件内容
		if _, err := file.Write(req.Chunk); err != nil {
			return status.Errorf(codes.Internal, "Failed to write file chunk: %v", err)
		}

		totalSize += int64(len(req.Chunk))
	}

	// 更新文件大小
	fileInfo.Size = totalSize

	// 关闭文件
	if err := file.Close(); err != nil {
		return status.Errorf(codes.Internal, "Failed to close file: %v", err)
	}
	file = nil

	// 返回上传结果
	return stream.SendAndClose(&storage.UploadFileResponse{
		FileInfo: &storage.FileInfo{
			Id:          fileInfo.ID,
			Filename:    fileInfo.Filename,
			ContentType: fileInfo.ContentType,
			Size:        fileInfo.Size,
			UserId:      fileInfo.UserID,
			StoragePath: fileInfo.StoragePath,
			CreatedAt:   fileInfo.CreatedAt,
			Metadata:    fileInfo.Metadata,
		},
		UploadedBytes: totalSize,
		Completed:     true,
	})
}

// DownloadFile 下载文件（支持流式下载）
func (s *StorageService) DownloadFile(req *storage.DownloadFileRequest, stream storage.StorageService_DownloadFileServer) error {
	// 获取文件信息
	fileInfo, err := s.GetFileInfo(context.Background(), &storage.GetFileInfoRequest{
		FileId: req.FileId,
		UserId: req.UserId,
	})
	if err != nil {
		return err
	}

	// 打开文件
	file, err := os.Open(fileInfo.StoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return status.Errorf(codes.NotFound, "File not found")
		}
		return status.Errorf(codes.Internal, "Failed to open file: %v", err)
	}
	defer file.Close()

	// 分块读取文件并发送
	chunkSize := 1024 * 1024 // 1MB
	buffer := make([]byte, chunkSize)
	isLastChunk := false

	for {
		// 读取文件块
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				isLastChunk = true
				break
			}
			return status.Errorf(codes.Internal, "Failed to read file: %v", err)
		}

		// 发送文件块
		if err := stream.Send(&storage.DownloadFileResponse{
			Chunk:       buffer[:n],
			IsLastChunk: n < chunkSize,
			FileInfo:    fileInfo,
		}); err != nil {
			return status.Errorf(codes.Internal, "Failed to send file chunk: %v", err)
		}
	}

	// 确保最后一块被发送
	if !isLastChunk {
		if err := stream.Send(&storage.DownloadFileResponse{
			Chunk:       []byte{},
			IsLastChunk: true,
			FileInfo:    fileInfo,
		}); err != nil {
			return status.Errorf(codes.Internal, "Failed to send final file chunk: %v", err)
		}
	}

	return nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(ctx context.Context, req *storage.DeleteFileRequest) (*common.Response, error) {
	// 获取文件信息
	fileInfo, err := s.GetFileInfo(ctx, &storage.GetFileInfoRequest{
		FileId: req.FileId,
		UserId: req.UserId,
	})
	if err != nil {
		return nil, err
	}

	// 删除文件
	if err := os.Remove(fileInfo.StoragePath); err != nil {
		if os.IsNotExist(err) {
			return &common.Response{
				Success: true,
				Message: "File not found, considered deleted",
				Code:    200,
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "Failed to delete file: %v", err)
	}

	return &common.Response{
		Success: true,
		Message: "File deleted successfully",
		Code:    200,
	}, nil
}

// GetFileInfo 获取文件信息
func (s *StorageService) GetFileInfo(ctx context.Context, req *storage.GetFileInfoRequest) (*storage.FileInfo, error) {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	// 搜索文件
	userDir := filepath.Join(config.Path, req.UserId)

	// 遍历用户目录下的所有文件，查找匹配的文件ID
	files, err := os.ReadDir(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, "User directory not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to read user directory: %v", err)
	}

	var foundFile string
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			fileID := strings.TrimSuffix(filename, filepath.Ext(filename))
			if fileID == req.FileId {
				foundFile = filename
				break
			}
		}
	}

	if foundFile == "" {
		return nil, status.Errorf(codes.NotFound, "File not found")
	}

	// 获取文件信息
	filePath := filepath.Join(userDir, foundFile)
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get file stat: %v", err)
	}

	// 解析文件名获取原始文件名（简单实现，实际可能需要存储在数据库中）
	// 这里假设文件名格式为：{fileID}.{original_extension}
	originalFilename := fmt.Sprintf("file_%s%s", req.FileId, filepath.Ext(foundFile))

	return &storage.FileInfo{
		Id:          req.FileId,
		Filename:    originalFilename,
		ContentType: "application/octet-stream", // 简单实现，实际应根据文件类型设置
		Size:        fileStat.Size(),
		UserId:      req.UserId,
		StoragePath: filePath,
		CreatedAt:   fileStat.ModTime().Format(time.RFC3339),
		Metadata:    map[string]string{},
	}, nil
}

// Check 健康检查
func (s *StorageService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (s *StorageService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 实现健康检查监听逻辑
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
