package logic

import (
	"context"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色列表：keyword 模糊匹配 role_key/role_name + 分页
func (l *ListRolesLogic) ListRoles(in *sys.SysRoleListReq) (*sys.SysRoleListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.RoleModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*sys.SysRole, 0, len(list))
	for _, r := range list {
		// 列表不回填 menu_ids（详情接口 GetRole 才返回）
		pbList = append(pbList, roleToPb(r, nil))
	}
	return &sys.SysRoleListReply{List: pbList, Total: total}, nil
}
