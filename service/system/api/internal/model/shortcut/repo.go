package shortcut

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 用户快捷入口数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

// ListByUser 按用户拉取快捷入口（有 MaxPerUser 上限）。
func (r *Repo) ListByUser(ctx context.Context, userId string) ([]SysUserShortcut, error) {
	var list []SysUserShortcut
	err := r.withCtx(ctx).
		Select("id", "user_id", "resource_id", "sort", "created_by", "updated_by", "created_at", "updated_at").
		Where("user_id = ?", userId).
		Order("sort ASC, created_at ASC").
		Limit(MaxPerUser).
		Find(&list).Error
	return list, err
}

func (r *Repo) CountByUser(ctx context.Context, userId string) (int64, error) {
	var n int64
	err := r.withCtx(ctx).Model(&SysUserShortcut{}).Where("user_id = ?", userId).Count(&n).Error
	return n, err
}

func (r *Repo) FindById(ctx context.Context, id string) (*SysUserShortcut, error) {
	var row SysUserShortcut
	err := r.withCtx(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) FindByUserAndResource(ctx context.Context, userId, resourceId string) (*SysUserShortcut, error) {
	var row SysUserShortcut
	err := r.withCtx(ctx).Where("user_id = ? AND resource_id = ?", userId, resourceId).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) Create(ctx context.Context, row *SysUserShortcut) error {
	return r.withCtx(ctx).Create(row).Error
}

func (r *Repo) Update(ctx context.Context, row *SysUserShortcut) error {
	return r.withCtx(ctx).Model(row).Select("resource_id", "sort", "updated_by", "updated_at").Updates(row).Error
}

// DeleteByUserAndIds 仅删除属于该用户的行，防越权。
func (r *Repo) DeleteByUserAndIds(ctx context.Context, userId string, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("user_id = ? AND id IN ?", userId, part).Delete(&SysUserShortcut{}).Error; err != nil {
			return err
		}
	}
	return nil
}
