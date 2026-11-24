package config

import "time"

// SysConfig 系统动态配置。
type SysConfig struct {
	Id          string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;size:128;not null;uniqueIndex:uk_sys_config_key" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;size:512;not null;default:''" json:"config_value"`
	ValueType   string    `gorm:"column:value_type;size:16;not null;default:'string'" json:"value_type"`
	GroupName   string    `gorm:"column:group_name;size:32;not null;default:''" json:"group_name"`
	Remark      string    `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	Status      int64     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedBy   string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy   string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysConfig) TableName() string { return "sys_config" }
