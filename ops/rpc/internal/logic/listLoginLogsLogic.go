package logic

import (
	"context"
	"strings"

	"datacenter/ops/rpc/internal/svc"
	"datacenter/ops/rpc/ops"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLoginLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLoginLogsLogic {
	return &ListLoginLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录日志查询：username 精确匹配可选 + success_filter（0.全部 1.仅成功 2.仅失败）+ 分页
func (l *ListLoginLogsLogic) ListLoginLogs(in *ops.LoginLogListReq) (*ops.LoginLogListReply, error) {
	list, total, err := l.svcCtx.LoginLogModel.FindList(l.ctx,
		strings.TrimSpace(in.Username), int64(in.SuccessFilter), in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	logs := make([]*ops.LoginLog, 0, len(list))
	for _, lg := range list {
		logs = append(logs, loginLogToPb(lg))
	}
	return &ops.LoginLogListReply{List: logs, Total: total}, nil
}
