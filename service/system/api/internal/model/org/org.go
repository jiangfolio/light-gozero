package org

import "time"

// SysOrg 系统组织。
type SysOrg struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	ParentId  string    `gorm:"column:parent_id;size:36;not null;default:''" json:"parent_id"`
	Name      string    `gorm:"column:name;size:64;not null;default:''" json:"name"`
	Code      string    `gorm:"column:code;size:64;not null;default:'';uniqueIndex:uk_sys_org_code" json:"code"`
	Sort      int64     `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int64     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedBy string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysOrg) TableName() string { return "sys_org" }
