package audit

import (
	"context"

	"gorm.io/gorm"
)

// Repo 审计日志数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) CreateLoginLog(ctx context.Context, row *LoginLog) error {
	return r.withCtx(ctx).Create(row).Error
}

func (r *Repo) CreateOperLog(ctx context.Context, row *OperLog) error {
	return r.withCtx(ctx).Create(row).Error
}

func (r *Repo) PageLoginLogs(ctx context.Context, username, result string, offset, limit int) ([]LoginLog, int64, error) {
	q := r.withCtx(ctx).Model(&LoginLog{})
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if result != "" {
		q = q.Where("result = ?", result)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []LoginLog
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *Repo) PageOperLogs(ctx context.Context, username, module, result string, offset, limit int) ([]OperLog, int64, error) {
	q := r.withCtx(ctx).Model(&OperLog{})
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if result != "" {
		q = q.Where("result = ?", result)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []OperLog
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// PageLoginLogsByUserId 按用户 ID 精确查登录日志；个人中心专用，ID 不可变、不受改名影响。
func (r *Repo) PageLoginLogsByUserId(ctx context.Context, userId, result string, offset, limit int) ([]LoginLog, int64, error) {
	q := r.withCtx(ctx).Model(&LoginLog{}).Where("user_id = ?", userId)
	if result != "" {
		q = q.Where("result = ?", result)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []LoginLog
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// PageOperLogsByUserId 按用户 ID 精确查操作日志；个人中心专用，ID 不可变、不受改名影响。
func (r *Repo) PageOperLogsByUserId(ctx context.Context, userId, module, result string, offset, limit int) ([]OperLog, int64, error) {
	q := r.withCtx(ctx).Model(&OperLog{}).Where("user_id = ?", userId)
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if result != "" {
		q = q.Where("result = ?", result)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []OperLog
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}
