package dict

import "time"

// SysDict 系统字典：parent_id 空为类型，否则为字典项。
type SysDict struct {
	Id        string    `gorm:"column:id;primaryKey;size:36" json:"id"`
	ParentId  string    `gorm:"column:parent_id;size:36;not null;default:''" json:"parent_id"`
	DictCode  string    `gorm:"column:dict_code;size:64;not null;default:''" json:"dict_code"`
	Label     string    `gorm:"column:label;size:128;not null;default:''" json:"label"`
	Value     string    `gorm:"column:value;size:128;not null;default:''" json:"value"`
	Sort      int64     `gorm:"column:sort;not null;default:0" json:"sort"`
	Status    int64     `gorm:"column:status;not null;default:1" json:"status"`
	Remark    string    `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	CreatedBy string    `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SysDict) TableName() string { return "sys_dict" }
