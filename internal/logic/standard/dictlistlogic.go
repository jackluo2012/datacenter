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

type DictListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 字典列表
func NewDictListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictListLogic {
	return &DictListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 字典列表: 页码默认值处理后转发, 业务错误直接透传。
// 列表项 rpc 不回传 items, dictToTypes 第二参传 nil, items 序列化为 [] 而非 null; 明细走 dictInfo 接口。
func (l *DictListLogic) DictList(req *types.DictListReq) (resp *types.DictListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.StandardRpc.ListDicts(l.ctx, &standardClient.DictListReq{
		Beid:    req.Beid,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.StdDict, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *dictToTypes(item, nil))
	}
	return &types.DictListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
