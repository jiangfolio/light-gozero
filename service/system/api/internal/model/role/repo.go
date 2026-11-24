package role

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 角色数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindById(ctx context.Context, id string) (*SysRole, error) {
	var role SysRole
	err := r.withCtx(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repo) ListByIds(ctx context.Context, ids []string) ([]SysRole, error) {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var list []SysRole
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		var chunk []SysRole
		if err := r.withCtx(ctx).Where("id IN ?", part).Find(&chunk).Error; err != nil {
			return nil, err
		}
		list = append(list, chunk...)
	}
	return list, nil
}

func (r *Repo) Page(ctx context.Context, name, code string, status int64, hasStatus bool, offset, limit int) ([]SysRole, int64, error) {
	q := r.withCtx(ctx).Model(&SysRole{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		q = q.Where("code = ?", code)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysRole
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *Repo) Create(ctx context.Context, role *SysRole) error {
	return r.withCtx(ctx).Create(role).Error
}

func (r *Repo) Save(ctx context.Context, role *SysRole) error {
	return r.withCtx(ctx).Where("id = ?", role.Id).Save(role).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysRole{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) BindResources(ctx context.Context, rows []RelRoleResource) error {
	if len(rows) == 0 {
		return nil
	}
	return r.withCtx(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(rows, batchutil.WriteSize).Error
}

func (r *Repo) UnbindResources(ctx context.Context, roleId string, resourceIds []string) error {
	resourceIds = batchutil.UniqueStrings(resourceIds)
	if roleId == "" || len(resourceIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(resourceIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("role_id = ? AND resource_id IN ?", roleId, part).Delete(&RelRoleResource{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteRelationsByRoleIds 删除角色的资源关联。
func (r *Repo) DeleteRelationsByRoleIds(ctx context.Context, roleIds []string) error {
	roleIds = batchutil.UniqueStrings(roleIds)
	if len(roleIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("role_id IN ?", part).Delete(&RelRoleResource{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteRelationsByResourceIds 删除资源侧的角色关联。
func (r *Repo) DeleteRelationsByResourceIds(ctx context.Context, resourceIds []string) error {
	resourceIds = batchutil.UniqueStrings(resourceIds)
	if len(resourceIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(resourceIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("resource_id IN ?", part).Delete(&RelRoleResource{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ReplaceOrgs 全量替换角色指定部门（先删后插）。
func (r *Repo) ReplaceOrgs(ctx context.Context, roleId string, rows []RelRoleOrg) error {
	if roleId == "" {
		return nil
	}
	if err := r.withCtx(ctx).Where("role_id = ?", roleId).Delete(&RelRoleOrg{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return r.withCtx(ctx).CreateInBatches(rows, batchutil.WriteSize).Error
}

// DeleteOrgsByRoleIds 删除角色的指定部门关联。
func (r *Repo) DeleteOrgsByRoleIds(ctx context.Context, roleIds []string) error {
	roleIds = batchutil.UniqueStrings(roleIds)
	if len(roleIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("role_id IN ?", part).Delete(&RelRoleOrg{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ListOrgIdsByRoleId 列出角色指定部门。
func (r *Repo) ListOrgIdsByRoleId(ctx context.Context, roleId string) ([]string, error) {
	if roleId == "" {
		return nil, nil
	}
	var ids []string
	err := r.withCtx(ctx).Model(&RelRoleOrg{}).Where("role_id = ?", roleId).Pluck("org_id", &ids).Error
	return ids, err
}

// ListOrgIdsByRoleIds 批量列出角色指定部门，返回 roleId -> orgIds。
func (r *Repo) ListOrgIdsByRoleIds(ctx context.Context, roleIds []string) (map[string][]string, error) {
	roleIds = batchutil.UniqueStrings(roleIds)
	out := make(map[string][]string)
	if len(roleIds) == 0 {
		return out, nil
	}
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		var rows []RelRoleOrg
		if err := r.withCtx(ctx).Select("role_id", "org_id").Where("role_id IN ?", part).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			out[row.RoleId] = append(out[row.RoleId], row.OrgId)
		}
	}
	return out, nil
}

// ListEnabledByUserId 查出用户已启用角色（含 data_scope），禁止 N+1。
func (r *Repo) ListEnabledByUserId(ctx context.Context, userId string) ([]SysRole, error) {
	if userId == "" {
		return nil, nil
	}
	var list []SysRole
	// 手写 JOIN：一次取出用户角色，避免先查关联再循环查角色
	err := r.withCtx(ctx).Raw(`
select r.id, r.name, r.code, r.data_scope, r.remark, r.status, r.created_by, r.updated_by, r.created_at, r.updated_at
from sys_role r
inner join rel_user_role ur on ur.role_id = r.id
where ur.user_id = ? and r.status = 1
`, userId).Scan(&list).Error
	return list, err
}
