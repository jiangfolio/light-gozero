package user

import "time"

// PasswordHistory 密码历史。
type PasswordHistory struct {
	Id           string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	UserId       string    `gorm:"column:user_id;size:36;not null" json:"user_id"`
	PasswordHash string    `gorm:"column:password_hash;size:100;not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PasswordHistory) TableName() string { return "sys_password_history" }
