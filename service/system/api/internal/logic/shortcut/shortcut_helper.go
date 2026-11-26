package shortcut

import (
	"path"
	"strings"

	"github.com/jiangbyte/light-gozero/common/convert"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	shortcutmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/shortcut"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

func toShortcutItem(row *shortcutmodel.SysUserShortcut, res *resmodel.SysResource, fullPath string) types.ShortcutItem {
	item := types.ShortcutItem{
		Id:         row.Id,
		ResourceId: row.ResourceId,
		Sort:       convert.Int64ToString(row.Sort),
		CreatedAt:  convert.TimeToString(row.CreatedAt),
		UpdatedAt:  convert.TimeToString(row.UpdatedAt),
	}
	if res != nil {
		item.Title = res.Name
		item.Icon = res.Icon
		item.Path = fullPath
	}
	return item
}

/** 拼接路由 path：绝对段直接用，相对段挂到父路径下 */
func joinRoutePath(parent, segment string) string {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return parent
	}
	if strings.HasPrefix(segment, "/") {
		return path.Clean(segment)
	}
	if parent == "" || parent == "/" {
		return path.Clean("/" + segment)
	}
	return path.Clean(parent + "/" + segment)
}

/** 自叶到根收集 path 段后反转拼出完整路由 */
func resolveFullPath(resMap map[string]*resmodel.SysResource, id string) string {
	cur, ok := resMap[id]
	if !ok || cur == nil {
		return ""
	}
	segments := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for cur != nil {
		if _, hit := seen[cur.Id]; hit {
			break
		}
		seen[cur.Id] = struct{}{}
		if cur.Path != "" {
			segments = append(segments, cur.Path)
			if strings.HasPrefix(strings.TrimSpace(cur.Path), "/") {
				break
			}
		}
		if cur.ParentId == "" || cur.ParentId == "0" {
			break
		}
		cur = resMap[cur.ParentId]
	}
	full := ""
	for i := len(segments) - 1; i >= 0; i-- {
		full = joinRoutePath(full, segments[i])
	}
	return full
}
