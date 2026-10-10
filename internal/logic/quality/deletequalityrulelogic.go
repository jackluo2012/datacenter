// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteQualityRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除质量规则
func NewDeleteQualityRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteQualityRuleLogic {
	return &DeleteQualityRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除质量规则: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteRule 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteQualityRuleLogic) DeleteQualityRule(req *types.QualityIdReq) error {
	_, err := l.svcCtx.QualityRpc.DeleteRule(l.ctx, &qualityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
