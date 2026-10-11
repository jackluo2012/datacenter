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

type SysRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleListLogic {
	return &SysRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 角色列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *SysRoleListLogic) SysRoleList(req *types.SysRoleListReq) (*types.SysRoleListResp, error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.SysRpc.ListRoles(l.ctx, &sysclient.SysRoleListReq{
		Beid:    beidOrCtx(req.Beid, l.ctx),
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SysRoleInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *sysRoleToTypes(item))
	}
	return &types.SysRoleListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
