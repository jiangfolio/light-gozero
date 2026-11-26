package shortcut

import "time"

// SysUserShortcut 用户个人快捷入口：只存资源 ID，展示字段始终跟资源走。
type SysUserShortcut struct {
	Id         string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	UserId     string    `gorm:"column:user_id;size:36;not null;default:''" json:"user_id"`
	ResourceId string    `gorm:"column:resource_id;size:36;not null;default:''" json:"resource_id"`
	Sort       int64     `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatedBy  string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy  string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysUserShortcut) TableName() string { return "sys_user_shortcut" }

// MaxPerUser 单用户快捷入口硬上限，list 与 create 共用。
const MaxPerUser = 8
