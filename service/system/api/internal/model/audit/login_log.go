package audit

import "time"

// LoginLog 登录审计。
type LoginLog struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	UserId    string    `gorm:"column:user_id;size:36;not null;default:''" json:"user_id"`
	Username  string    `gorm:"column:username;size:64;not null;default:''" json:"username"`
	Ip        string    `gorm:"column:ip;size:64;not null;default:''" json:"ip"`
	UserAgent string    `gorm:"column:user_agent;size:255;not null;default:''" json:"user_agent"`
	Result    string    `gorm:"column:result;size:16;not null;default:''" json:"result"`
	Message   string    `gorm:"column:message;size:255;not null;default:''" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (LoginLog) TableName() string { return "sys_login_log" }
