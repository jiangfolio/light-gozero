package role

import "time"

// 数据范围常量。
const (
	DataScopeSelf         int64 = 1 // 本人
	DataScopeDept         int64 = 2 // 本部门
	DataScopeDeptAndBelow int64 = 3 // 本部门及以下
	DataScopeCustom       int64 = 4 // 指定部门
	DataScopeAll          int64 = 5 // 全局
)

// SysRole 系统角色。
type SysRole struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	Name      string    `gorm:"column:name;size:64;not null" json:"name"`
	Code      string    `gorm:"column:code;size:64;not null;uniqueIndex:uk_sys_role_code" json:"code"`
	DataScope int64     `gorm:"column:data_scope;not null;default:5" json:"data_scope"`
	Remark    string    `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	Status    int64     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedBy string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysRole) TableName() string { return "sys_role" }
