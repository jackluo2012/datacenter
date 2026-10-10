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

type CreateQualityJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建质量任务
func NewCreateQualityJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateQualityJobLogic {
	return &CreateQualityJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建质量任务: 网关请求逐字段映射为 rpc 请求, 转发后把任务转为网关类型返回
func (l *CreateQualityJobLogic) CreateQualityJob(req *types.CreateQualityJobReq) (resp *types.QualityJob, err error) {
	reply, err := l.svcCtx.QualityRpc.CreateJob(l.ctx, &qualityClient.CreateJobReq{
		Name:         req.Name,
		DatasourceId: req.DatasourceId,
		DbName:       req.DbName,
		RuleIds:      req.RuleIds,
		Cron:         req.Cron,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return qualityJobToTypes(reply.GetJob()), nil
}
