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

type RunQualityJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行质量任务
func NewRunQualityJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunQualityJobLogic {
	return &RunQualityJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 执行质量任务: 网关请求逐字段映射为 rpc 请求, 转发后把执行结果转为网关类型返回
func (l *RunQualityJobLogic) RunQualityJob(req *types.QualityIdReq) (resp *types.RunJobResp, err error) {
	reply, err := l.svcCtx.QualityRpc.RunJob(l.ctx, &qualityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.RunJobResp{
		JobId:      reply.GetJobId(),
		TotalRules: reply.GetTotalRules(),
		Passed:     reply.GetPassed(),
		LastStatus: reply.GetLastStatus(),
	}, nil
}
