package notice

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 通知公告数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindById(ctx context.Context, id string) (*SysNotice, error) {
	var row SysNotice
	err := r.withCtx(ctx).
		Select("id", "title", "content", "type", "status", "sort", "publish_at", "remark", "created_by", "updated_by", "created_at", "updated_at").
		Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListByIds 按 ID 批量加载。
func (r *Repo) ListByIds(ctx context.Context, ids []string) ([]SysNotice, error) {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var list []SysNotice
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		var chunk []SysNotice
		if err := r.withCtx(ctx).
			Select("id", "title", "content", "type", "status", "sort", "publish_at", "remark", "created_by", "updated_by", "created_at", "updated_at").
			Where("id IN ?", part).Find(&chunk).Error; err != nil {
			return nil, err
		}
		list = append(list, chunk...)
	}
	return list, nil
}

// Page 管理端分页；status 仅在 hasStatus 时过滤。
func (r *Repo) Page(ctx context.Context, title string, typ, status int64, hasType, hasStatus bool, offset, limit int) ([]SysNotice, int64, error) {
	q := r.withCtx(ctx).Model(&SysNotice{})
	if title != "" {
		q = q.Where("title LIKE ?", "%"+title+"%")
	}
	if hasType {
		q = q.Where("type = ?", typ)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysNotice
	err := q.Select("id", "title", "content", "type", "status", "sort", "publish_at", "remark", "created_by", "updated_by", "created_at", "updated_at").
		Order("sort asc, publish_at desc, created_at desc").
		Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// PagePublished 仅已发布分页。
func (r *Repo) PagePublished(ctx context.Context, title string, typ int64, hasType bool, offset, limit int) ([]SysNotice, int64, error) {
	q := r.withCtx(ctx).Model(&SysNotice{}).Where("status = ?", StatusPublished)
	if title != "" {
		q = q.Where("title LIKE ?", "%"+title+"%")
	}
	if hasType {
		q = q.Where("type = ?", typ)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysNotice
	err := q.Select("id", "title", "content", "type", "status", "sort", "publish_at", "remark", "created_by", "updated_by", "created_at", "updated_at").
		Order("sort asc, publish_at desc, created_at desc").
		Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *Repo) Create(ctx context.Context, row *SysNotice) error {
	return r.withCtx(ctx).Create(row).Error
}

func (r *Repo) Save(ctx context.Context, row *SysNotice) error {
	return r.withCtx(ctx).Where("id = ?", row.Id).Save(row).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysNotice{}).Error; err != nil {
			return err
		}
	}
	return nil
}
