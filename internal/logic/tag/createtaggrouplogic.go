// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTagGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建人群
func NewCreateTagGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagGroupLogic {
	return &CreateTagGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTagGroupLogic) CreateTagGroup(req *types.CreateTagGroupReq) (resp *types.TagGroupInfo, err error) {
	reply, err := l.svcCtx.TagRpc.CreateGroup(l.ctx, &tagClient.CreateGroupReq{
		Name:          req.Name,
		ObjectId:      req.ObjectId,
		CircleRuleSql: req.CircleRuleSql,
		Beid:          req.Beid,
	})
	if err != nil {
		return nil, err //业务错误直接透传 rpc error
	}
	return groupToTypes(reply.GetGroup()), nil
}
