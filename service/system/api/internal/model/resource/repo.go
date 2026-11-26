package resource

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 资源数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindById(ctx context.Context, id string) (*SysResource, error) {
	var res SysResource
	err := r.withCtx(ctx).Where("id = ?", id).First(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repo) CountChildren(ctx context.Context, parentIds []string) (int64, error) {
	parentIds = batchutil.UniqueStrings(parentIds)
	if len(parentIds) == 0 {
		return 0, nil
	}
	var total int64
	for _, part := range batchutil.Chunk(parentIds, batchutil.InSize) {
		var n int64
		if err := r.withCtx(ctx).Model(&SysResource{}).Where("parent_id IN ?", part).Count(&n).Error; err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func (r *Repo) Page(ctx context.Context, name, typ, parentId string, status int64, hasStatus bool, offset, limit int) ([]SysResource, int64, error) {
	q := r.withCtx(ctx).Model(&SysResource{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	if parentId != "" {
		q = q.Where("parent_id = ?", parentId)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysResource
	err := q.Order("sort asc, created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListByParent 按父节点一次性列出子资源（如菜单下按钮），无分页。
func (r *Repo) ListByParent(ctx context.Context, parentId, name, typ string, status int64, hasStatus bool) ([]SysResource, error) {
	q := r.withCtx(ctx).Model(&SysResource{}).Where("parent_id = ?", parentId)
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var list []SysResource
	err := q.Order("sort asc, created_at desc").Find(&list).Error
	return list, err
}

// ListForTree 管理端组树：仅目录/菜单，不含按钮；按钮按 parent_id 另查。
func (r *Repo) ListForTree(ctx context.Context, name, typ string, status int64, hasStatus bool) ([]SysResource, error) {
	// 树接口不承载按钮；显式查 button 时直接空结果
	if typ == "button" {
		return nil, nil
	}
	q := r.withCtx(ctx).Model(&SysResource{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	} else {
		q = q.Where("type IN ?", []string{"directory", "menu"})
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var list []SysResource
	err := q.Order("sort asc, created_at desc").Find(&list).Error
	return list, err
}

func (r *Repo) ListByIds(ctx context.Context, ids []string) ([]SysResource, error) {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var list []SysResource
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		var chunk []SysResource
		if err := r.withCtx(ctx).Where("id IN ?", part).Find(&chunk).Error; err != nil {
			return nil, err
		}
		list = append(list, chunk...)
	}
	return list, nil
}

func (r *Repo) ListPermCodesByUser(ctx context.Context, userId string) ([]string, error) {
	var codes []string
	// 多表权限展开，GORM 关联不便表达且需 DISTINCT
	err := r.withCtx(ctx).Raw(`
select distinct r.perm_code
from sys_resources r
inner join rel_role_resources rr on rr.resource_id = r.id
inner join rel_user_role ur on ur.role_id = rr.role_id
where ur.user_id = ?
  and r.type = 'button'
  and r.status = 1
  and r.perm_code is not null
  and r.perm_code <> ''`, userId).Scan(&codes).Error
	return codes, err
}

func (r *Repo) ListByUser(ctx context.Context, userId string) ([]SysResource, error) {
	var list []SysResource
	// 多表 JOIN + DISTINCT 资源树，显式列避免 SELECT *
	err := r.withCtx(ctx).Raw(`
select distinct
  r.id, r.parent_id, r.name, r.type, r.perm_code, r.path, r.component_path,
  r.redirect, r.route_name, r.icon, r.hidden, r.keep_alive, r.affix, r.sort,
  r.status, r.created_by, r.updated_by, r.created_at, r.updated_at
from sys_resources r
inner join rel_role_resources rr on rr.resource_id = r.id
inner join rel_user_role ur on ur.role_id = rr.role_id
where ur.user_id = ?
  and r.status = 1
order by r.sort asc, r.created_at asc`, userId).Scan(&list).Error
	return list, err
}

func (r *Repo) Create(ctx context.Context, res *SysResource) error {
	return r.withCtx(ctx).Create(res).Error
}

func (r *Repo) Save(ctx context.Context, res *SysResource) error {
	return r.withCtx(ctx).Where("id = ?", res.Id).Save(res).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysResource{}).Error; err != nil {
			return err
		}
	}
	return nil
}
