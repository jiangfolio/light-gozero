// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package shortcut

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/errorx"
	resmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ShortcutListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拉当前用户快捷入口；登录即可，结果集有上限
func NewShortcutListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortcutListLogic {
	return &ShortcutListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShortcutList 返回当前用户快捷入口；标题/路径/图标一律从资源现算。
func (l *ShortcutListLogic) ShortcutList() (resp *types.ShortcutListResp, err error) {
	// 1. 仅要求登录
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return nil, errorx.New(401, "未登录或登录已失效")
	}
	// 2. 拉本人快捷入口
	list, err := l.svcCtx.Repo.Shortcut.ListByUser(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return &types.ShortcutListResp{Items: []types.ShortcutItem{}}, nil
	}
	// 3. 批量加载资源及祖先，拼完整路由，禁止 N+1
	seedIds := make([]string, 0, len(list))
	for i := range list {
		seedIds = append(seedIds, list[i].ResourceId)
	}
	resMap, err := l.loadResourceClosure(seedIds)
	if err != nil {
		return nil, err
	}
	items := make([]types.ShortcutItem, 0, len(list))
	for i := range list {
		res := resMap[list[i].ResourceId]
		fullPath := ""
		if res != nil {
			fullPath = resolveFullPath(resMap, res.Id)
		}
		items = append(items, toShortcutItem(&list[i], res, fullPath))
	}
	return &types.ShortcutListResp{Items: items}, nil
}

/** 加载资源及其祖先闭包，用于拼 path */
func (l *ShortcutListLogic) loadResourceClosure(seedIds []string) (map[string]*resmodel.SysResource, error) {
	resMap := make(map[string]*resmodel.SysResource)
	pending := batchutil.UniqueStrings(seedIds)
	for len(pending) > 0 {
		rows, err := l.svcCtx.Repo.Resource.ListByIds(l.ctx, pending)
		if err != nil {
			return nil, err
		}
		next := make([]string, 0)
		for i := range rows {
			// 必须拷贝，避免循环变量地址复用
			cp := rows[i]
			resMap[cp.Id] = &cp
			if cp.ParentId != "" && cp.ParentId != "0" {
				if _, ok := resMap[cp.ParentId]; !ok {
					next = append(next, cp.ParentId)
				}
			}
		}
		pending = make([]string, 0, len(next))
		for _, id := range batchutil.UniqueStrings(next) {
			if _, ok := resMap[id]; !ok {
				pending = append(pending, id)
			}
		}
	}
	return resMap, nil
}
