package dict

import (
	"github.com/jiangbyte/light-gozero/common/convert"
	dictmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/dict"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

func toDictItem(row *dictmodel.SysDict) *types.DictItem {
	return &types.DictItem{
		Id: row.Id, ParentId: row.ParentId, DictCode: row.DictCode,
		Label: row.Label, Value: row.Value,
		Sort: convert.Int64ToString(row.Sort), Status: convert.Int64ToString(row.Status),
		Remark: row.Remark, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy,
		CreatedAt: convert.TimeToString(row.CreatedAt), UpdatedAt: convert.TimeToString(row.UpdatedAt),
	}
}
