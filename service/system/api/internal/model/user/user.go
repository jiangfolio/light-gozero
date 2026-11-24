package user

import "time"

// SysUser 系统用户。
type SysUser struct {
	Id                 string     `gorm:"column:id;primaryKey;size:36" json:"id"`
	OrgId              string     `gorm:"column:org_id;size:36;not null;default:''" json:"org_id"`
	Username           string     `gorm:"column:username;size:64;not null;uniqueIndex:uk_sys_user_username" json:"username"`
	PasswordHash       string     `gorm:"column:password_hash;size:100;not null" json:"-"`
	Nickname           string     `gorm:"column:nickname;size:64;not null;default:''" json:"nickname"`
	Status             int64      `gorm:"column:status;not null;default:1" json:"status"`
	FailCount          int64      `gorm:"column:fail_count;not null;default:0" json:"fail_count"`
	LockedUntil        *time.Time `gorm:"column:locked_until" json:"locked_until"`
	PasswordChangedAt  time.Time  `gorm:"column:password_changed_at" json:"password_changed_at"`
	MustChangePassword int64      `gorm:"column:must_change_password;not null;default:0" json:"must_change_password"`
	CreatedBy          string     `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy          string     `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (SysUser) TableName() string { return "sys_user" }
