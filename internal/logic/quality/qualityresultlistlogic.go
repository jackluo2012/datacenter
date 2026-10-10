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

type QualityResultListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 质量结果列表
func NewQualityResultListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityResultListLogic {
	return &QualityResultListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 质量结果列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发;
// pass_filter 取值 0 表示不过滤, 直接透传由 rpc 侧处理
func (l *QualityResultListLogic) QualityResultList(req *types.ResultListReq) (resp *types.ResultListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListResults(l.ctx, &qualityClient.ResultListReq{
		JobId:      req.JobId,
		PassFilter: req.PassFilter,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.QualityResult, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *qualityResultToTypes(item))
	}
	return &types.ResultListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
