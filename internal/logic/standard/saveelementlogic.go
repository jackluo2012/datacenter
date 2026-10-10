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

type SaveElementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存数据元
func NewSaveElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveElementLogic {
	return &SaveElementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存数据元: 网关请求逐字段映射为 rpc 请求, 转发后把数据元转为网关类型返回。
// 契约差异: 网关 SaveElementReq 无 Id 字段, rpc SaveElementReq.Id 恒为 0, 即 rpc 侧始终按新建处理。
func (l *SaveElementLogic) SaveElement(req *types.SaveElementReq) (resp *types.StdElement, err error) {
	reply, err := l.svcCtx.StandardRpc.SaveElement(l.ctx, &standardClient.SaveElementReq{
		Id:            req.Id,
		ElementCode:   req.ElementCode,
		ElementName:   req.ElementName,
		DataType:      req.DataType,
		DataLength:    req.DataLength,
		DataPrecision: req.DataPrecision,
		DictCode:      req.DictCode,
		Comment:       req.Comment,
		Beid:          req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return elementToTypes(reply.GetElement()), nil
}
