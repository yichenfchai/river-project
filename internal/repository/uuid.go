package repository

import (
	"crypto/sha1"
	"fmt"
)

// UUID5 基于 namespace + name 生成确定性 UUID v5
func UUID5(namespace, name string) string {
	h := sha1.New()
	h.Write([]byte(namespace + ":" + name))
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}
