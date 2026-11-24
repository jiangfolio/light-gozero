package dict

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 字典数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindById(ctx context.Context, id string) (*SysDict, error) {
	var row SysDict
	err := r.withCtx(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// FindTypeByCode 按编码查找字典类型（parent_id 为空且 value 为空）。
func (r *Repo) FindTypeByCode(ctx context.Context, dictCode string) (*SysDict, error) {
	var row SysDict
	err := r.withCtx(ctx).Where("parent_id = '' AND dict_code = ? AND value = ''", dictCode).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) CountChildren(ctx context.Context, parentIds []string) (int64, error) {
	parentIds = batchutil.UniqueStrings(parentIds)
	if len(parentIds) == 0 {
		return 0, nil
	}
	var total int64
	for _, part := range batchutil.Chunk(parentIds, batchutil.InSize) {
		var n int64
		if err := r.withCtx(ctx).Model(&SysDict{}).Where("parent_id IN ?", part).Count(&n).Error; err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func (r *Repo) Page(ctx context.Context, parentId, dictCode, label string, status int64, hasStatus bool, offset, limit int) ([]SysDict, int64, error) {
	q := r.withCtx(ctx).Model(&SysDict{})
	if parentId != "" {
		q = q.Where("parent_id = ?", parentId)
	}
	if dictCode != "" {
		q = q.Where("dict_code = ?", dictCode)
	}
	if label != "" {
		q = q.Where("label LIKE ?", "%"+label+"%")
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysDict
	err := q.Order("sort asc, created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListEnabledItemsByCode 按编码取启用中的字典项（排除类型行）。
func (r *Repo) ListEnabledItemsByCode(ctx context.Context, dictCode string) ([]SysDict, error) {
	var list []SysDict
	err := r.withCtx(ctx).
		Where("dict_code = ? AND parent_id <> '' AND status = 1", dictCode).
		Order("sort asc, created_at asc").
		Find(&list).Error
	return list, err
}

func (r *Repo) Create(ctx context.Context, row *SysDict) error {
	return r.withCtx(ctx).Create(row).Error
}

func (r *Repo) Save(ctx context.Context, row *SysDict) error {
	return r.withCtx(ctx).Where("id = ?", row.Id).Save(row).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysDict{}).Error; err != nil {
			return err
		}
	}
	return nil
}
