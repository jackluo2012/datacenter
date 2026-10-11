// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	sysclient "datacenter/sys/rpc/sysClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuTreeLogic {
	return &SysMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 全量菜单树(管理用, 含目录/菜单/按钮): beid 缺省取 JWT ctx, 转发 sys.rpc MenuTree
func (l *SysMenuTreeLogic) SysMenuTree(req *types.SysMenuTreeReq) (*types.SysMenuTreeResp, error) {
	reply, err := l.svcCtx.SysRpc.MenuTree(l.ctx, &sysclient.IdReq{
		Id:   0, //树查询不按 id 过滤
		Beid: beidOrCtx(req.Beid, l.ctx),
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SysMenuInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *sysMenuToTypes(item))
	}
	return &types.SysMenuTreeResp{List: list}, nil
}
