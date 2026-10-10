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

type DictInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 字典详情
func NewDictInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictInfoLogic {
	return &DictInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 字典详情: 网关请求逐字段映射为 rpc 请求, 转发后把字典连同 items 转为网关类型返回
func (l *DictInfoLogic) DictInfo(req *types.StdIdReq) (resp *types.StdDict, err error) {
	reply, err := l.svcCtx.StandardRpc.GetDict(l.ctx, &standardClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return dictToTypes(reply.GetDict(), reply.GetItems()), nil
}
