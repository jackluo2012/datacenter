package logic

import (
	"context"

	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoutersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoutersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoutersLogic {
	return &GetRoutersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 前端动态路由：当前用户可见菜单（目录+菜单）转 RouterBranch 树。
// component=path 映射约定：component "datasource/DataSourceList" → web/src/views/datasource/DataSourceList.vue；
// 根路径"/"的目录子节点拼全路径。uid 由网关从 JWT 解出后传入，rpc 内不做 JWT 校验（信任网关）
func (l *GetRoutersLogic) GetRouters(in *sys.GetRoutersReq) (*sys.GetRoutersReply, error) {
	if in.Uid <= 0 {
		return nil, ErrIdRequired
	}
	menus, err := l.svcCtx.MenuModel.FindByUid(l.ctx, in.Uid, in.Beid)
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		// 兜底播种：全局菜单为空（首次部署直调 GetRouters）时初始化种子，可重入
		all, err := l.svcCtx.MenuModel.FindAll(l.ctx, in.Beid)
		if err != nil {
			return nil, err
		}
		if len(all) == 0 {
			if err = seedRBAC(l.ctx, l.svcCtx, in.Beid); err != nil {
				return nil, err
			}
			if menus, err = l.svcCtx.MenuModel.FindByUid(l.ctx, in.Uid, in.Beid); err != nil {
				return nil, err
			}
		}
	}
	// 建树 → 剔除按钮 → 转路由结构
	tree := StripButtons(BuildMenuTree(menusToPb(menus)))
	return &sys.GetRoutersReply{Routers: BuildRouters(tree)}, nil
}
