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

type SecGrantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 授权矩阵列表
func NewSecGrantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecGrantListLogic {
	return &SecGrantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 授权矩阵列表: 页码默认值处理后转发(subject/resource 为过滤条件, 空串表示不过滤), 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *SecGrantListLogic) SecGrantList(req *types.SecGrantListReq) (resp *types.SecGrantListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.SecurityRpc.ListGrants(l.ctx, &securityClient.GrantListReq{
		Beid:     req.Beid,
		Subject:  req.Subject,
		Resource: req.Resource,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SecGrant, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *secGrantToTypes(item))
	}
	return &types.SecGrantListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
