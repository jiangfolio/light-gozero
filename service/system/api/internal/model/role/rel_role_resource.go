package role

import "time"

// RelRoleResource 角色-资源关联。
type RelRoleResource struct {
	Id           string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	RoleId       string    `gorm:"column:role_id;size:36;not null;uniqueIndex:uk_rel_role_resources" json:"role_id"`
	ResourceId   string    `gorm:"column:resource_id;size:36;not null;uniqueIndex:uk_rel_role_resources" json:"resource_id"`
	ResourceType string    `gorm:"column:resource_type;size:16;not null;default:''" json:"resource_type"`
	CreatedBy    string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RelRoleResource) TableName() string { return "rel_role_resources" }
