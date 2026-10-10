// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/standard/rpc/standardClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishStdModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发布标准模型
func NewPublishStdModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishStdModelLogic {
	return &PublishStdModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 发布标准模型: 网关请求逐字段映射为 rpc 请求后转发, rpc 侧版本+1并落快照, 返回新版本号与快照
func (l *PublishStdModelLogic) PublishStdModel(req *types.PublishStdModelReq) (resp *types.PublishStdModelResp, err error) {
	reply, err := l.svcCtx.StandardRpc.PublishModel(l.ctx, &standardClient.PublishReq{
		Id:       req.Id,
		Beid:     req.Beid,
		Operator: req.Operator,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.PublishStdModelResp{
		Version:      reply.GetVersion(),
		SnapshotJson: reply.GetSnapshotJson(),
	}, nil
}
