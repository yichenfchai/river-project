package errors

import (
	"regexp"
	"strings"
)

// 泄漏模式 → 替换
var leakPatterns = []struct {
	Pattern *regexp.Regexp
	Replace string
}{
	// PostgreSQL 约束名
	{regexp.MustCompile(`"([a-z_]+)_([a-z_]+)_key"`), `"***_key"`},
	// PostgreSQL 表名+列名
	{regexp.MustCompile(`relation "([a-z_]+)"`), `relation "***"`},
	{regexp.MustCompile(`column "([a-z_]+)"`), `column "***"`},
	// 重复键值
	{regexp.MustCompile(`Key \(([^)]+)\)=\((.*?)\)`), `Key ($1)=(***)`},
	// MySQL 表名
	{regexp.MustCompile(`Duplicate entry '([^']+)' for key`), `Duplicate entry '***' for key`},
	// 连接信息
	{regexp.MustCompile(`(host=[^\s]+)`), `host=***`},
	{regexp.MustCompile(`(dbname=[^\s]+)`), `dbname=***`},
	// Panic 信息
	{regexp.MustCompile(`index out of range \[\d+\]`), `index out of range [N]`},
	{regexp.MustCompile(`invalid memory address (0x[0-9a-f]+)`), `invalid memory address ***`},
	// UUID
	{regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), `<uuid>`},
}

// sanitizeDetail 脱敏错误详情
func sanitizeDetail(code ErrorCode, cause error) string {
	if cause == nil {
		return code.GetMessage()
	}
	raw := cause.Error()

	// 内部错误和数据库错误: 不返回任何原始消息
	if code == ErrInternal || code == ErrDatabaseError {
		return code.GetMessage()
	}

	// 执行脱敏
	for _, rule := range leakPatterns {
		raw = rule.Pattern.ReplaceAllString(raw, rule.Replace)
	}

	// 截断长度
	if len(raw) > 200 {
		raw = raw[:200] + "..."
	}
	return raw
}

// classifyError 生产环境安全分类
func classifyError(code ErrorCode, cause error) *SafeDetail {
	if cause == nil {
		return nil
	}
	errStr := cause.Error()
	sd := &SafeDetail{}

	switch {
	case strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "unique constraint"):
		sd.Type = "unique_violation"
		sd.Summary = "数据重复"

	case strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "violates foreign"):
		sd.Type = "foreign_key_violation"
		sd.Summary = "外键约束"

	case strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "i/o timeout"):
		sd.Type = "connection_error"
		sd.Summary = "数据库连接失败"

	case strings.Contains(errStr, "deadlock") ||
		strings.Contains(errStr, "could not serialize"):
		sd.Type = "deadlock"
		sd.Summary = "死锁"

	case strings.Contains(errStr, "no rows"):
		sd.Type = "not_found"
		sd.Summary = "记录不存在"

	default:
		sd.Type = "database_error"
		sd.Summary = "数据库错误"
	}

	return sd
}
