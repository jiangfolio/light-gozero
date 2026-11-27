package notice

import "time"

// SysNotice 通知公告。
type SysNotice struct {
	Id        string     `gorm:"column:id;primaryKey;size:36" json:"id"`
	Title     string     `gorm:"column:title;size:200;not null;default:''" json:"title"`
	Content   string     `gorm:"column:content;type:text;not null" json:"content"`
	Type      int64      `gorm:"column:type;not null;default:1" json:"type"`
	Status    int64      `gorm:"column:status;not null;default:0" json:"status"`
	Sort      int64      `gorm:"column:sort;not null;default:0" json:"sort"`
	PublishAt *time.Time `gorm:"column:publish_at" json:"publish_at"`
	Remark    string     `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	CreatedBy string     `gorm:"column:created_by;size:36;not null;default:''" json:"created_by"`
	UpdatedBy string     `gorm:"column:updated_by;size:36;not null;default:''" json:"updated_by"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (SysNotice) TableName() string { return "sys_notice" }

// 状态与类型约定。
const (
	StatusDraft     int64 = 0
	StatusPublished int64 = 1
	TypeNotice      int64 = 1
	TypeAnnounce    int64 = 2
)
