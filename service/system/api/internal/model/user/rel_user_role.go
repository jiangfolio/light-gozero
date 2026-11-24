package user

import "time"

// RelUserRole 用户-角色关联。
type RelUserRole struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	UserId    string    `gorm:"column:user_id;size:36;not null;uniqueIndex:uk_rel_user_role" json:"user_id"`
	RoleId    string    `gorm:"column:role_id;size:36;not null;uniqueIndex:uk_rel_user_role" json:"role_id"`
	CreatedBy string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RelUserRole) TableName() string { return "rel_user_role" }
