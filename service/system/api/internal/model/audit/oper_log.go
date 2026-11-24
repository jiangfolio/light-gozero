package audit

import "time"

// OperLog 操作审计。
type OperLog struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	UserId    string    `gorm:"column:user_id;size:36;not null;default:''" json:"user_id"`
	Username  string    `gorm:"column:username;size:64;not null;default:''" json:"username"`
	Module    string    `gorm:"column:module;size:64;not null;default:''" json:"module"`
	Action    string    `gorm:"column:action;size:64;not null;default:''" json:"action"`
	Method    string    `gorm:"column:method;size:16;not null;default:''" json:"method"`
	Path      string    `gorm:"column:path;size:255;not null;default:''" json:"path"`
	ReqBody   string    `gorm:"column:req_body;size:2000;not null;default:''" json:"req_body"`
	Ip        string    `gorm:"column:ip;size:64;not null;default:''" json:"ip"`
	LatencyMs int64     `gorm:"column:latency_ms;not null;default:0" json:"latency_ms"`
	Result    string    `gorm:"column:result;size:16;not null;default:''" json:"result"`
	Message   string    `gorm:"column:message;size:512;not null;default:''" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OperLog) TableName() string { return "sys_oper_log" }
