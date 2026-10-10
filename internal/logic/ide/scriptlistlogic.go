// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ide

import (
	"context"

	"datacenter/ide/rpc/ideClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScriptListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 脚本列表
func NewScriptListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScriptListLogic {
	return &ScriptListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 脚本列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *ScriptListLogic) ScriptList(req *types.ScriptListReq) (resp *types.ScriptListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.IdeRpc.ListScripts(l.ctx, &ideClient.ScriptListReq{
		Beid:    req.Beid,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.Script, 0, len(reply.List))
	for _, item := range reply.List {
		list = append(list, *scriptToTypes(item))
	}
	return &types.ScriptListResp{
		List:  list,
		Total: reply.Total,
	}, nil
}
