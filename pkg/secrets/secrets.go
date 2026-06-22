package secrets

import (
	"fmt"
	"os"
	"strings"
)

// Store 密钥获取的统一抽象 — 按优先级链读取：文件 > 环境变量 > 默认值
type Store struct {
	baseDir string // Docker/K8s Secrets mount 路径
}

// New 创建密钥存取器。baseDir 为空时默认 /run/secrets
func New(baseDir string) *Store {
	if baseDir == "" {
		baseDir = "/run/secrets"
	}
	return &Store{baseDir: baseDir}
}

// Get 按优先级读取密钥
//   key 示例: "LLM_API_KEY"
//   文件优先: {baseDir}/llm_api_key
//   其次环境变量: LLM_API_KEY
//   最后默认值: fallback
func (s *Store) Get(key string, fallback string) string {
	filePath := fmt.Sprintf("%s/%s", s.baseDir, strings.ToLower(key))
	if data, err := os.ReadFile(filePath); err == nil {
		return strings.TrimSpace(string(data))
	}
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Masked 返回脱敏摘要，安全用于日志输出
func (s *Store) Masked(key string) string {
	v := s.Get(key, "")
	if v == "" {
		return "未设置"
	}
	if len(v) <= 8 {
		return "***（已设置）"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// Summary 返回所有敏感密钥的脱敏状态列表（启动时打印）
func (s *Store) Summary() []string {
	keys := []string{"DB_PASSWORD", "JWT_SECRET", "LLM_API_KEY"}
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("[secrets] %s = %s", k, s.Masked(k)))
	}
	return lines
}
