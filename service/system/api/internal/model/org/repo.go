package org

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 组织数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindById(ctx context.Context, id string) (*SysOrg, error) {
	var o SysOrg
	err := r.withCtx(ctx).Where("id = ?", id).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repo) ListByIds(ctx context.Context, ids []string) ([]SysOrg, error) {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var out []SysOrg
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		var rows []SysOrg
		if err := r.withCtx(ctx).Where("id IN ?", part).Find(&rows).Error; err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (r *Repo) CountChildren(ctx context.Context, parentIds []string) (int64, error) {
	parentIds = batchutil.UniqueStrings(parentIds)
	if len(parentIds) == 0 {
		return 0, nil
	}
	var total int64
	for _, part := range batchutil.Chunk(parentIds, batchutil.InSize) {
		var n int64
		if err := r.withCtx(ctx).Model(&SysOrg{}).Where("parent_id IN ?", part).Count(&n).Error; err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

// ListDescendantIds 按层 BFS 取子孙组织 ID（不含根自身），IN 分批。
func (r *Repo) ListDescendantIds(ctx context.Context, rootIds []string) ([]string, error) {
	frontier := batchutil.UniqueStrings(rootIds)
	if len(frontier) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{})
	var out []string
	for len(frontier) > 0 {
		var next []string
		for _, part := range batchutil.Chunk(frontier, batchutil.InSize) {
			var children []string
			if err := r.withCtx(ctx).Model(&SysOrg{}).Where("parent_id IN ?", part).Pluck("id", &children).Error; err != nil {
				return nil, err
			}
			for _, id := range children {
				if id == "" {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				out = append(out, id)
				next = append(next, id)
			}
		}
		frontier = next
	}
	return out, nil
}

// applyIDInFilter 对列施加多分批 IN（OR 连接），避免单 IN 超 500。
func applyIDInFilter(q *gorm.DB, column string, ids []string) *gorm.DB {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return q.Where("1 = 0")
	}
	parts := batchutil.Chunk(ids, batchutil.InSize)
	if len(parts) == 1 {
		return q.Where(column+" IN ?", parts[0])
	}
	return q.Where(func(db *gorm.DB) *gorm.DB {
		for i, part := range parts {
			if i == 0 {
				db = db.Where(column+" IN ?", part)
			} else {
				db = db.Or(column+" IN ?", part)
			}
		}
		return db
	})
}

func (r *Repo) Page(ctx context.Context, name, code string, status int64, hasStatus bool, scopeOrgIds []string, selfOrgId string, offset, limit int) ([]SysOrg, int64, error) {
	q := r.withCtx(ctx).Model(&SysOrg{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		q = q.Where("code = ?", code)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	// 数据范围：self 仅本组织；scopeOrgIds!=nil 则 id IN；nil 表示全局不加约束
	if selfOrgId != "" {
		q = q.Where("id = ?", selfOrgId)
	} else if scopeOrgIds != nil {
		q = applyIDInFilter(q, "id", scopeOrgIds)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysOrg
	err := q.Order("sort asc, created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListForTree 按条件全量列出组织（含数据范围），供后端组树；无分页。
func (r *Repo) ListForTree(ctx context.Context, name, code string, status int64, hasStatus bool, scopeOrgIds []string, selfOrgId string) ([]SysOrg, error) {
	q := r.withCtx(ctx).Model(&SysOrg{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		q = q.Where("code = ?", code)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	if selfOrgId != "" {
		q = q.Where("id = ?", selfOrgId)
	} else if scopeOrgIds != nil {
		q = applyIDInFilter(q, "id", scopeOrgIds)
	}
	var list []SysOrg
	err := q.Order("sort asc, created_at desc").Find(&list).Error
	return list, err
}

func (r *Repo) Create(ctx context.Context, o *SysOrg) error {
	return r.withCtx(ctx).Create(o).Error
}

func (r *Repo) Save(ctx context.Context, o *SysOrg) error {
	// 按主键更新，避免无 WHERE 全表写
	return r.withCtx(ctx).Where("id = ?", o.Id).Save(o).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysOrg{}).Error; err != nil {
			return err
		}
	}
	return nil
}
