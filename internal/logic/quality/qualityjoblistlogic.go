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

type QualityJobListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 质量任务列表
func NewQualityJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityJobListLogic {
	return &QualityJobListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 质量任务列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *QualityJobListLogic) QualityJobList(req *types.JobListReq) (resp *types.JobListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListJobs(l.ctx, &qualityClient.JobListReq{
		Beid: req.Beid,
		Page: page,
		Size: size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.QualityJob, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *qualityJobToTypes(item))
	}
	return &types.JobListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
