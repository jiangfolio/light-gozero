package svc

import (
	"context"

	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/rbac"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	rolemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/role"
)

// newScopeLoader 多角色数据范围并集解析。
func newScopeLoader(repo *model.Repo) rbac.ScopeLoader {
	return func(ctx context.Context, userId string) (*rbac.DataScopeResult, error) {
		// 1. 一次查出启用角色
		roles, err := repo.Role.ListEnabledByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		if len(roles) == 0 {
			return &rbac.DataScopeResult{}, nil
		}
		for _, role := range roles {
			if role.DataScope == rolemodel.DataScopeAll {
				return &rbac.DataScopeResult{All: true, OrgIds: []string{}}, nil
			}
		}

		// 2. 批量取指定部门；用户 org 只查一次
		customRoleIds := make([]string, 0)
		needDept, needBelow, needSelf := false, false, false
		for _, role := range roles {
			switch role.DataScope {
			case rolemodel.DataScopeSelf:
				needSelf = true
			case rolemodel.DataScopeDept:
				needDept = true
			case rolemodel.DataScopeDeptAndBelow:
				needBelow = true
			case rolemodel.DataScopeCustom:
				customRoleIds = append(customRoleIds, role.Id)
			}
		}
		customMap, err := repo.Role.ListOrgIdsByRoleIds(ctx, customRoleIds)
		if err != nil {
			return nil, err
		}

		orgSet := make(map[string]struct{})
		add := func(ids ...string) {
			for _, id := range ids {
				if id != "" {
					orgSet[id] = struct{}{}
				}
			}
		}
		for _, ids := range customMap {
			add(ids...)
		}

		var userOrgId string
		if needDept || needBelow || needSelf {
			u, err := repo.User.FindById(ctx, userId)
			if err != nil {
				return nil, err
			}
			userOrgId = u.OrgId
		}
		if needDept {
			add(userOrgId)
		}
		if needBelow {
			add(userOrgId)
			desc, err := repo.Org.ListDescendantIds(ctx, []string{userOrgId})
			if err != nil {
				return nil, err
			}
			add(desc...)
		}

		orgIds := make([]string, 0, len(orgSet))
		for id := range orgSet {
			orgIds = append(orgIds, id)
		}
		orgIds = batchutil.UniqueStrings(orgIds)

		// 3. 仅本人且无组织集合
		if len(orgIds) == 0 && needSelf {
			return &rbac.DataScopeResult{SelfOnly: true, OrgIds: []string{}}, nil
		}
		return &rbac.DataScopeResult{OrgIds: orgIds}, nil
	}
}
