package user

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 用户数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

func (r *Repo) FindByUsername(ctx context.Context, username string) (*SysUser, error) {
	var u SysUser
	err := r.withCtx(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) FindById(ctx context.Context, id string) (*SysUser, error) {
	var u SysUser
	err := r.withCtx(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) ListByIds(ctx context.Context, ids []string) ([]SysUser, error) {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	var out []SysUser
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		var rows []SysUser
		if err := r.withCtx(ctx).Where("id IN ?", part).Find(&rows).Error; err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (r *Repo) CountByOrgIds(ctx context.Context, orgIds []string) (int64, error) {
	orgIds = batchutil.UniqueStrings(orgIds)
	if len(orgIds) == 0 {
		return 0, nil
	}
	var total int64
	for _, part := range batchutil.Chunk(orgIds, batchutil.InSize) {
		var n int64
		if err := r.withCtx(ctx).Model(&SysUser{}).Where("org_id IN ?", part).Count(&n).Error; err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func (r *Repo) Page(ctx context.Context, username, orgId string, status int64, hasStatus bool, scopeOrgIds []string, selfUserId string, offset, limit int) ([]SysUser, int64, error) {
	q := r.withCtx(ctx).Model(&SysUser{})
	if username != "" {
		q = q.Where("username LIKE ?", "%"+username+"%")
	}
	if orgId != "" {
		q = q.Where("org_id = ?", orgId)
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	// 数据范围：本人 / 组织 IN / 全局
	if selfUserId != "" {
		q = q.Where("id = ?", selfUserId)
	} else if scopeOrgIds != nil {
		q = applyUserOrgInFilter(q, scopeOrgIds)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysUser
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func applyUserOrgInFilter(q *gorm.DB, ids []string) *gorm.DB {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return q.Where("1 = 0")
	}
	parts := batchutil.Chunk(ids, batchutil.InSize)
	if len(parts) == 1 {
		return q.Where("org_id IN ?", parts[0])
	}
	return q.Where(func(db *gorm.DB) *gorm.DB {
		for i, part := range parts {
			if i == 0 {
				db = db.Where("org_id IN ?", part)
			} else {
				db = db.Or("org_id IN ?", part)
			}
		}
		return db
	})
}

func (r *Repo) Create(ctx context.Context, u *SysUser) error {
	return r.withCtx(ctx).Create(u).Error
}

func (r *Repo) Save(ctx context.Context, u *SysUser) error {
	return r.withCtx(ctx).Where("id = ?", u.Id).Save(u).Error
}

func (r *Repo) Delete(ctx context.Context, ids []string) error {
	ids = batchutil.UniqueStrings(ids)
	if len(ids) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(ids, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("id IN ?", part).Delete(&SysUser{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) ListUserIdsByRoles(ctx context.Context, roleIds []string) ([]string, error) {
	roleIds = batchutil.UniqueStrings(roleIds)
	if len(roleIds) == 0 {
		return nil, nil
	}
	var ids []string
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		var chunk []string
		err := r.withCtx(ctx).Model(&RelUserRole{}).Where("role_id IN ?", part).Distinct("user_id").Pluck("user_id", &chunk).Error
		if err != nil {
			return nil, err
		}
		ids = append(ids, chunk...)
	}
	return batchutil.UniqueStrings(ids), nil
}

// ListUserIdsByResources 查出绑定过指定资源的用户（删资源后刷缓存用）。
func (r *Repo) ListUserIdsByResources(ctx context.Context, resourceIds []string) ([]string, error) {
	resourceIds = batchutil.UniqueStrings(resourceIds)
	if len(resourceIds) == 0 {
		return nil, nil
	}
	var ids []string
	for _, part := range batchutil.Chunk(resourceIds, batchutil.InSize) {
		var chunk []string
		// 跨表 JOIN：按资源反查用户，ORM 单表 API 表达不清
		err := r.withCtx(ctx).Raw(`
select distinct ur.user_id
from rel_user_role ur
inner join rel_role_resources rr on rr.role_id = ur.role_id
where rr.resource_id in ?`, part).Scan(&chunk).Error
		if err != nil {
			return nil, err
		}
		ids = append(ids, chunk...)
	}
	return batchutil.UniqueStrings(ids), nil
}

func (r *Repo) BindRoles(ctx context.Context, rows []RelUserRole) error {
	if len(rows) == 0 {
		return nil
	}
	return r.withCtx(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(rows, batchutil.WriteSize).Error
}

func (r *Repo) UnbindRoles(ctx context.Context, userId string, roleIds []string) error {
	roleIds = batchutil.UniqueStrings(roleIds)
	if userId == "" || len(roleIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("user_id = ? AND role_id IN ?", userId, part).Delete(&RelUserRole{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteRelationsByUserIds 删除用户的角色关联。
func (r *Repo) DeleteRelationsByUserIds(ctx context.Context, userIds []string) error {
	userIds = batchutil.UniqueStrings(userIds)
	if len(userIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(userIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("user_id IN ?", part).Delete(&RelUserRole{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteRelationsByRoleIds 删除角色侧的用户关联。
func (r *Repo) DeleteRelationsByRoleIds(ctx context.Context, roleIds []string) error {
	roleIds = batchutil.UniqueStrings(roleIds)
	if len(roleIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(roleIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("role_id IN ?", part).Delete(&RelUserRole{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// AddPasswordHistory 追加密码历史。
func (r *Repo) AddPasswordHistory(ctx context.Context, row *PasswordHistory) error {
	return r.withCtx(ctx).Create(row).Error
}

// ListRecentPasswordHashes 最近 N 条历史哈希（含当前可另传）。
func (r *Repo) ListRecentPasswordHashes(ctx context.Context, userId string, limit int) ([]string, error) {
	if userId == "" || limit <= 0 {
		return nil, nil
	}
	var hashes []string
	err := r.withCtx(ctx).Model(&PasswordHistory{}).
		Where("user_id = ?", userId).
		Order("created_at desc").
		Limit(limit).
		Pluck("password_hash", &hashes).Error
	return hashes, err
}

// TrimPasswordHistory 仅保留最近 keep 条。
func (r *Repo) TrimPasswordHistory(ctx context.Context, userId string, keep int) error {
	if userId == "" || keep < 0 {
		return nil
	}
	// 1. 取出应保留的最近 keep 条 id
	var keepIds []string
	err := r.withCtx(ctx).Model(&PasswordHistory{}).
		Where("user_id = ?", userId).
		Order("created_at desc").
		Limit(keep).
		Pluck("id", &keepIds).Error
	if err != nil {
		return err
	}
	// 2. 删除不在保留集中的历史（空保留集则全删）
	q := r.withCtx(ctx).Where("user_id = ?", userId)
	if len(keepIds) > 0 {
		q = q.Where("id NOT IN ?", keepIds)
	}
	return q.Delete(&PasswordHistory{}).Error
}

// DeletePasswordHistoryByUserIds 删用户时清历史。
func (r *Repo) DeletePasswordHistoryByUserIds(ctx context.Context, userIds []string) error {
	userIds = batchutil.UniqueStrings(userIds)
	if len(userIds) == 0 {
		return nil
	}
	for _, part := range batchutil.Chunk(userIds, batchutil.InSize) {
		if err := r.withCtx(ctx).Where("user_id IN ?", part).Delete(&PasswordHistory{}).Error; err != nil {
			return err
		}
	}
	return nil
}
