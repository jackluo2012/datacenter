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

type SecCheckPermLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 鉴权检查（subject 对 resource_id 的许可是否覆盖其密级）
func NewSecCheckPermLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecCheckPermLogic {
	return &SecCheckPermLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 鉴权检查: 网关请求逐字段映射为 rpc 请求, rpc 侧汇总主体授权后判定是否放行并给出原因。
// 未获许可不是错误, allowed=false + reason 随正常响应返回。
func (l *SecCheckPermLogic) SecCheckPerm(req *types.SecCheckPermReq) (resp *types.SecCheckPermResp, err error) {
	reply, err := l.svcCtx.SecurityRpc.CheckPerm(l.ctx, &securityClient.CheckPermReq{
		Subject:    req.Subject,
		Resource:   req.Resource,
		ResourceId: req.ResourceId,
		LevelCode:  req.LevelCode,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SecCheckPermResp{
		Allowed: reply.GetAllowed(),
		Reason:  reply.GetReason(),
	}, nil
}
