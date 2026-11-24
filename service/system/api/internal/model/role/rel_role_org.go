package role

import "time"

// RelRoleOrg 角色-指定部门关联。
type RelRoleOrg struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	RoleId    string    `gorm:"column:role_id;size:36;not null;uniqueIndex:uk_rel_role_orgs" json:"role_id"`
	OrgId     string    `gorm:"column:org_id;size:36;not null;uniqueIndex:uk_rel_role_orgs" json:"org_id"`
	CreatedBy string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RelRoleOrg) TableName() string { return "rel_role_orgs" }
