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

type SaveMdmRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存主数据记录
func NewSaveMdmRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMdmRecordLogic {
	return &SaveMdmRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存主数据记录: 网关请求逐字段映射为 rpc 请求后转发(Id=0 新建, >0 更新;
// BizCode 新建时可为空直接透传, rpc 侧按模型编码策略自动生成),
// 转发后把记录转为网关类型返回, 业务错误直接透传
func (l *SaveMdmRecordLogic) SaveMdmRecord(req *types.SaveMdmRecordReq) (resp *types.MdmRecordInfo, err error) {
	reply, err := l.svcCtx.MdmRpc.SaveRecord(l.ctx, &mdmClient.SaveRecordReq{
		Id:       req.Id,
		ModelId:  req.ModelId,
		BizCode:  req.BizCode,
		DataJson: req.DataJson,
		Beid:     req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return mdmRecordToTypes(reply.GetRecord()), nil
}
