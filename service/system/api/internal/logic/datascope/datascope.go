package datascope

import (
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/rbac"
)

// ForUserPage 将 Resolve 结果转为用户分页约束。
// scopeOrgIds==nil 表示全局；selfUserId 非空表示仅本人。
func ForUserPage(scope *rbac.DataScopeResult, uid string, reqOrgId string) (scopeOrgIds []string, selfUserId string, empty bool) {
	if scope == nil {
		return []string{}, "", true
	}
	if scope.All {
		if reqOrgId != "" {
			return nil, "", false
		}
		return nil, "", false
	}
	if scope.SelfOnly {
		return nil, uid, false
	}
	ids := scope.OrgIds
	if reqOrgId != "" {
		found := false
		for _, id := range ids {
			if id == reqOrgId {
				found = true
				break
			}
		}
		if !found {
			return []string{}, "", true
		}
		return []string{reqOrgId}, "", false
	}
	if len(ids) == 0 {
		return []string{}, "", true
	}
	return ids, "", false
}

// ForOrgPage 将 Resolve 结果转为组织分页约束。
func ForOrgPage(scope *rbac.DataScopeResult, userOrgId string) (scopeOrgIds []string, selfOrgId string, empty bool) {
	if scope == nil {
		return []string{}, "", true
	}
	if scope.All {
		return nil, "", false
	}
	if scope.SelfOnly {
		if userOrgId == "" {
			return []string{}, "", true
		}
		return nil, userOrgId, false
	}
	if len(scope.OrgIds) == 0 {
		return []string{}, "", true
	}
	return scope.OrgIds, "", false
}

// AllowOrg 校验目标组织是否在数据范围内。
func AllowOrg(scope *rbac.DataScopeResult, userOrgId, targetOrgId string) error {
	if scope == nil || targetOrgId == "" {
		return errorx.New(403, "无数据权限")
	}
	if scope.All {
		return nil
	}
	if scope.SelfOnly {
		if targetOrgId == userOrgId {
			return nil
		}
		return errorx.New(403, "无数据权限")
	}
	for _, id := range scope.OrgIds {
		if id == targetOrgId {
			return nil
		}
	}
	return errorx.New(403, "无数据权限")
}

// AllowUser 校验目标用户是否在数据范围内（self 仅本人；否则看所属组织）。
func AllowUser(scope *rbac.DataScopeResult, selfUserId, targetUserId, targetOrgId string) error {
	if scope == nil || targetUserId == "" {
		return errorx.New(403, "无数据权限")
	}
	if scope.All {
		return nil
	}
	if scope.SelfOnly {
		if targetUserId == selfUserId {
			return nil
		}
		return errorx.New(403, "无数据权限")
	}
	return AllowOrg(scope, "", targetOrgId)
}
