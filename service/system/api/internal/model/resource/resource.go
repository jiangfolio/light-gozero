package resource

import "time"

// SysResource 系统资源（目录/菜单/按钮）。
type SysResource struct {
	Id            string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	ParentId      string    `gorm:"column:parent_id;size:36;not null;default:''" json:"parent_id"`
	Name          string    `gorm:"column:name;size:64;not null" json:"name"`
	Type          string    `gorm:"column:type;size:16;not null" json:"type"`
	PermCode      *string   `gorm:"column:perm_code;size:128;uniqueIndex:uk_sys_resources_perm_code" json:"perm_code"`
	Path          string    `gorm:"column:path;size:255;not null;default:''" json:"path"`
	ComponentPath string    `gorm:"column:component_path;size:255;not null;default:''" json:"component_path"`
	Redirect      string    `gorm:"column:redirect;size:255;not null;default:''" json:"redirect"`
	RouteName     string    `gorm:"column:route_name;size:64;not null;default:''" json:"route_name"`
	Icon          string    `gorm:"column:icon;size:64;not null;default:''" json:"icon"`
	Hidden        int64     `gorm:"column:hidden;not null;default:0" json:"hidden"`
	KeepAlive     int64     `gorm:"column:keep_alive;not null;default:0" json:"keep_alive"`
	Affix         int64     `gorm:"column:affix;not null;default:0" json:"affix"`
	Sort          int64     `gorm:"column:sort;not null;default:0" json:"sort"`
	Status        int64     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedBy     string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy     string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysResource) TableName() string { return "sys_resources" }
