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

type QualityErrorDataListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 错误数据下钻列表
func NewQualityErrorDataListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityErrorDataListLogic {
	return &QualityErrorDataListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 错误数据下钻列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发;
// job_id/rule_id 为 0 表示不按该维度过滤, 直接透传由 rpc 侧处理
func (l *QualityErrorDataListLogic) QualityErrorDataList(req *types.ErrorDataListReq) (resp *types.ErrorDataListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListErrorData(l.ctx, &qualityClient.ErrorDataListReq{
		JobId:  req.JobId,
		RuleId: req.RuleId,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.QualityErrorData, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, errorDataToTypes(item))
	}
	return &types.ErrorDataListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
