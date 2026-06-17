package auth

// 权限位掩码设计
// 每种操作占一个比特位，角色 = 多个权限的 OR 组合
//
// 示例:
//   user    = PermRead | PermPost | PermComment   = 0b0111
//   monitor = PermRead | PermPost | PermComment | PermUpload = 0b1111
//   admin   = 0xFFFFFFFF

const (
	PermRead     uint32 = 1 << iota // 0x01: 读
	PermPost                        // 0x02: 发帖
	PermComment                     // 0x04: 评论
	PermUpload                      // 0x08: 上传图片
	PermQuiz                        // 0x10: 答题
	PermModerate                    // 0x20: 审核
	PermManageUsers                 // 0x40: 管理用户
	PermSystem                      // 0x80: 系统配置
)

// RolePermissions 角色 → 权限位掩码
var RolePermissions = map[string]uint32{
	"user":    PermRead | PermPost | PermComment | PermQuiz,
	"monitor": PermRead | PermPost | PermComment | PermQuiz | PermUpload,
	"admin":   0xFFFFFFFF,
}

// GetPermissions 获取角色的权限掩码
func GetPermissions(role string) uint32 {
	if perm, ok := RolePermissions[role]; ok {
		return perm
	}
	return 0
}

// HasPermission 检查用户是否有某权限
func HasPermission(userRole string, required uint32) bool {
	userPerm := GetPermissions(userRole)
	return userPerm&required == required
}
