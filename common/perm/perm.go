package perm

import "regexp"

// 权限码三段式：{模块}:{资源}:{动作}，段内多单词用 _
const Pattern = `^[a-z]+(_[a-z]+)*:[a-z]+(_[a-z]+)*:[a-z]+(_[a-z]+)*$`

var codeRegexp = regexp.MustCompile(Pattern)

const (
	SystemOrgCreate = "system:org:create"
	SystemOrgUpdate = "system:org:update"
	SystemOrgDelete = "system:org:delete"
	SystemOrgDetail = "system:org:detail"
	SystemOrgPage   = "system:org:page"

	SystemUserCreate      = "system:user:create"
	SystemUserUpdate      = "system:user:update"
	SystemUserDelete      = "system:user:delete"
	SystemUserDetail      = "system:user:detail"
	SystemUserPage        = "system:user:page"
	SystemUserBindRoles   = "system:user:bind_roles"
	SystemUserUnbindRoles = "system:user:unbind_roles"
	SystemUserResources   = "system:user:resources"

	SystemRoleCreate          = "system:role:create"
	SystemRoleUpdate          = "system:role:update"
	SystemRoleDelete          = "system:role:delete"
	SystemRoleDetail          = "system:role:detail"
	SystemRolePage            = "system:role:page"
	SystemRoleBindResources   = "system:role:bind_resources"
	SystemRoleUnbindResources = "system:role:unbind_resources"

	SystemResourceCreate = "system:resource:create"
	SystemResourceUpdate = "system:resource:update"
	SystemResourceDelete = "system:resource:delete"
	SystemResourceDetail = "system:resource:detail"
	SystemResourcePage   = "system:resource:page"

	SystemAuthLogout         = "system:auth:logout"
	SystemAuthChangePassword = "system:auth:change_password"

	SystemAuditLoginPage = "system:audit:login_page"
	SystemAuditOperPage  = "system:audit:oper_page"

	SystemConfigPage   = "system:config:page"
	SystemConfigDetail = "system:config:detail"
	SystemConfigUpdate = "system:config:update"

	SystemDictCreate = "system:dict:create"
	SystemDictUpdate = "system:dict:update"
	SystemDictDelete = "system:dict:delete"
	SystemDictDetail = "system:dict:detail"
	SystemDictPage   = "system:dict:page"
)

// Valid 校验权限码格式。
func Valid(code string) bool {
	return codeRegexp.MatchString(code)
}
