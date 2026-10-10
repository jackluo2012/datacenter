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

type SaveMdmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存主数据模型
func NewSaveMdmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMdmModelLogic {
	return &SaveMdmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存主数据模型: 编码策略缺省为号段模式(segment), 网关请求逐字段映射为 rpc 请求,
// 转发后把模型转为网关类型返回, 业务错误直接透传
func (l *SaveMdmModelLogic) SaveMdmModel(req *types.SaveMdmModelReq) (resp *types.MdmModelInfo, err error) {
	codeStrategy := req.CodeStrategy
	if codeStrategy == "" {
		// 编码策略缺省: segment 号段模式
		codeStrategy = "segment"
	}
	reply, err := l.svcCtx.MdmRpc.SaveModel(l.ctx, &mdmClient.SaveModelReq{
		Id:           req.Id,
		ModelCode:    req.ModelCode,
		ModelName:    req.ModelName,
		TableName:    req.TableName,
		FieldsJson:   req.FieldsJson,
		CodeStrategy: codeStrategy,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return mdmModelToTypes(reply.GetModel()), nil
}
