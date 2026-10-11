package logic

import (
	"context"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户列表：keyword 模糊匹配 username/nickname/mobile + status 过滤（<0 不过滤）+ 分页；
// 每行附带 role_ids（批量查询避免 N+1）
func (l *ListUsersLogic) ListUsers(in *sys.SysUserListReq) (*sys.SysUserListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.UserModel.FindList(l.ctx, in.Beid, in.Keyword, int64(in.Status), in.Page, in.Size)
	if err != nil {
		return nil, err
	}

	// 批量取角色绑定
	userIds := make([]int64, 0, len(list))
	for _, u := range list {
		userIds = append(userIds, u.Id)
	}
	roleMap, err := l.svcCtx.UserModel.GetRoleIdsBatch(l.ctx, userIds)
	if err != nil {
		return nil, err
	}

	pbList := make([]*sys.SysUser, 0, len(list))
	for _, u := range list {
		pbList = append(pbList, userToPb(u, roleMap[u.Id]))
	}
	return &sys.SysUserListReply{List: pbList, Total: total}, nil
}
