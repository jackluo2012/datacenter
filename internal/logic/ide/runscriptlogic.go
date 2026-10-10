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

type RunScriptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行脚本（仅 SELECT，行数保护）
func NewRunScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunScriptLogic {
	return &RunScriptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 执行脚本: 行数上限默认值处理(maxRows<=0 取 100)后转发, 转发后把执行结果转为网关类型返回
func (l *RunScriptLogic) RunScript(req *types.RunScriptReq) (resp *types.RunScriptResp, err error) {
	maxRows := req.MaxRows
	if maxRows <= 0 {
		maxRows = 100
	}
	reply, err := l.svcCtx.IdeRpc.RunScript(l.ctx, &ideClient.RunScriptReq{
		Id:      req.Id,
		Beid:    req.Beid,
		MaxRows: maxRows,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.RunScriptResp{
		ColumnsJson: reply.GetColumnsJson(),
		RowsJson:    reply.GetRowsJson(),
		RowCount:    reply.GetRowCount(),
		CostMs:      reply.GetCostMs(),
		Truncated:   reply.GetTruncated(),
	}, nil
}
