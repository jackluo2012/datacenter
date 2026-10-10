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

type SaveStdModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存标准模型
func NewSaveStdModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveStdModelLogic {
	return &SaveStdModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存标准模型: 网关请求逐字段映射为 rpc 请求, 转发后把模型转为网关类型返回。
// 契约差异: 网关 SaveStdModelReq 无 Id 字段, rpc SaveModelReq.Id 恒为 0, 即 rpc 侧始终按新建处理。
func (l *SaveStdModelLogic) SaveStdModel(req *types.SaveStdModelReq) (resp *types.StdModel, err error) {
	reply, err := l.svcCtx.StandardRpc.SaveModel(l.ctx, &standardClient.SaveModelReq{
		Id:           req.Id,
		ModelCode:    req.ModelCode,
		ModelName:    req.ModelName,
		ElementsJson: req.ElementsJson,
		Beid:         req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return stdModelToTypes(reply.GetModel()), nil
}
