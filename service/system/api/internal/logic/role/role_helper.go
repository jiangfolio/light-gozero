package role

import (
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	rolemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/role"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

func parseDataScope(raw string, def int64) (int64, error) {
	if raw == "" {
		return def, nil
	}
	v, err := convert.Int64FromString(raw)
	if err != nil {
		return 0, errorx.New(400, "数据范围无效")
	}
	switch v {
	case rolemodel.DataScopeSelf, rolemodel.DataScopeDept, rolemodel.DataScopeDeptAndBelow,
		rolemodel.DataScopeCustom, rolemodel.DataScopeAll:
		return v, nil
	default:
		return 0, errorx.New(400, "数据范围无效")
	}
}

func toRoleItem(role *rolemodel.SysRole, orgIds []string) *types.RoleItem {
	if orgIds == nil {
		orgIds = []string{}
	}
	return &types.RoleItem{
		Id: role.Id, Name: role.Name, Code: role.Code,
		DataScope: convert.Int64ToString(role.DataScope), OrgIds: orgIds,
		Remark: role.Remark, Status: convert.Int64ToString(role.Status),
		CreatedBy: role.CreatedBy, UpdatedBy: role.UpdatedBy,
		CreatedAt: convert.TimeToString(role.CreatedAt), UpdatedAt: convert.TimeToString(role.UpdatedAt),
	}
}

func buildRoleOrgRows(roleId, op string, orgIds []string, nextID func() (string, error)) ([]rolemodel.RelRoleOrg, error) {
	orgIds = batchutil.UniqueStrings(orgIds)
	rows := make([]rolemodel.RelRoleOrg, 0, len(orgIds))
	for _, oid := range orgIds {
		id, err := nextID()
		if err != nil {
			return nil, err
		}
		rows = append(rows, rolemodel.RelRoleOrg{Id: id, RoleId: roleId, OrgId: oid, CreatedBy: op})
	}
	return rows, nil
}
