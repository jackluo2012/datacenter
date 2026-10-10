package logic

import (
	"context"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGrantsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGrantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGrantsLogic {
	return &ListGrantsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 授权列表（分页，subject/resource 非空时精确过滤）
func (l *ListGrantsLogic) ListGrants(in *security.GrantListReq) (*security.GrantListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.GrantModel.List(l.ctx, in.Beid, in.Subject, in.Resource, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*security.SecGrant, 0, len(list))
	for _, grant := range list {
		pbList = append(pbList, grantToPb(grant))
	}
	return &security.GrantListReply{List: pbList, Total: total}, nil
}
