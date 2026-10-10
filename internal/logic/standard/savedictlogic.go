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

type SaveDictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存字典（含项，整存整取）
func NewSaveDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDictLogic {
	return &SaveDictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存字典: items 切片逐项拷贝为 pb DictItem, 整存整取(全量覆盖), 转发后把字典连同 items 转为网关类型返回。
// 契约差异: 网关 SaveDictReq 无 Id 字段, rpc SaveDictReq.Id 恒为 0; 网关 StdDictItem 无 Beid,
// pb DictItem.Beid 置 0, 由 rpc 侧按 SaveDictReq.Beid 统一归一。
func (l *SaveDictLogic) SaveDict(req *types.SaveDictReq) (resp *types.StdDict, err error) {
	items := make([]*standardClient.DictItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &standardClient.DictItem{
			Id:        item.Id,
			DictId:    item.DictId,
			ItemCode:  item.ItemCode,
			ItemValue: item.ItemValue,
			Sort:      item.Sort,
		})
	}
	reply, err := l.svcCtx.StandardRpc.SaveDict(l.ctx, &standardClient.SaveDictReq{
		Id:       req.Id,
		DictCode: req.DictCode,
		DictName: req.DictName,
		Comment:  req.Comment,
		Beid:     req.Beid,
		Items:    items,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return dictToTypes(reply.GetDict(), reply.GetItems()), nil
}
