package util

import (
	"crypto/sha256"
)

/**
 * 使用sha256进行哈希
 */
func HashSHA256(data []byte)[]byte{
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}