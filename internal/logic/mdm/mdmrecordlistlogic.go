// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package mdm

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/mdm/rpc/mdmClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MdmRecordListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 主数据记录列表
func NewMdmRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MdmRecordListLogic {
	return &MdmRecordListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 主数据记录列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *MdmRecordListLogic) MdmRecordList(req *types.MdmRecordListReq) (resp *types.MdmRecordListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.MdmRpc.ListRecords(l.ctx, &mdmClient.RecordListReq{
		Beid:    req.Beid,
		ModelId: req.ModelId,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.MdmRecordInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *mdmRecordToTypes(item))
	}
	return &types.MdmRecordListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
