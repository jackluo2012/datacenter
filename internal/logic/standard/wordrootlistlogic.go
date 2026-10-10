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

type WordRootListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 词根列表
func NewWordRootListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WordRootListLogic {
	return &WordRootListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 词根列表: 页码默认值处理后转发, 业务错误直接透传
func (l *WordRootListLogic) WordRootList(req *types.WordRootListReq) (resp *types.WordRootListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.StandardRpc.ListWordRoots(l.ctx, &standardClient.WordRootListReq{
		Beid:    req.Beid,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.StdWordRoot, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *wordRootToTypes(item))
	}
	return &types.WordRootListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
