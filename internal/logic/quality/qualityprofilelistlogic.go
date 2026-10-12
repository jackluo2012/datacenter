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

type QualityProfileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 表剖析历史列表
func NewQualityProfileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QualityProfileListLogic {
	return &QualityProfileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 表剖析历史列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发;
// 剖析记录逐条经共享映射转出, make 保证空列表序列化为 [] 而非 null
func (l *QualityProfileListLogic) QualityProfileList(req *types.ProfileListQReq) (resp *types.ProfileListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.QualityRpc.ListProfiles(l.ctx, &qualityClient.ProfileListReq{
		Beid:         req.Beid,
		DatasourceId: req.DatasourceId,
		TableName:    req.TableName,
		Page:         page,
		Size:         size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.TableProfileInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, tableProfileToTypes(item))
	}
	return &types.ProfileListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
