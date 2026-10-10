// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ops

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/ops/rpc/opsClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunArchiveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行归档
func NewRunArchiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunArchiveLogic {
	return &RunArchiveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 执行归档: 网关请求逐字段映射为 rpc 请求后转发(真实跑库分批 insert-select + delete 事务由 rpc 侧处理),
// 转发后把归档结果逐字段映射为网关类型返回, 业务错误直接透传
func (l *RunArchiveLogic) RunArchive(req *types.OpsIdReq) (resp *types.RunArchiveResp, err error) {
	reply, err := l.svcCtx.OpsRpc.RunArchive(l.ctx, &opsClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.RunArchiveResp{
		ArchivedRows: reply.GetArchivedRows(),
		DeletedRows:  reply.GetDeletedRows(),
		Status:       reply.GetStatus(),
	}, nil
}
