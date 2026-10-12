// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecLevelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分级目录列表
func NewSecLevelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecLevelListLogic {
	return &SecLevelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 分级目录列表: rpc ListLevels 入参为 IdReq 且无分页字段, 一次返回全量分级, 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *SecLevelListLogic) SecLevelList(req *types.SecIdReq) (resp *types.SecLevelListResp, err error) {
	reply, err := l.svcCtx.SecurityRpc.ListLevels(l.ctx, &securityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SecLevelInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *secLevelToTypes(item))
	}
	return &types.SecLevelListResp{
		List: list,
	}, nil
}
