package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// CalculateSHA256 计算SHA256哈希值
func CalculateSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateSHA256Bytes 计算字节数组的SHA256哈希值
func CalculateSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

