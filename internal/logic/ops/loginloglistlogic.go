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

type LoginLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登录日志列表
func NewLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogListLogic {
	return &LoginLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 登录日志列表: SuccessFilter(0 全部/1 成功/2 失败)直传, 页码默认值处理(page<=0 取 1, size<=0 取 10)
// 后转发, 业务错误直接透传
func (l *LoginLogListLogic) LoginLogList(req *types.LoginLogListReq) (resp *types.LoginLogListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.OpsRpc.ListLoginLogs(l.ctx, &opsClient.LoginLogListReq{
		Username:      req.Username,
		SuccessFilter: req.SuccessFilter,
		Page:          page,
		Size:          size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.OpsLoginLog, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *loginLogToTypes(item))
	}
	return &types.LoginLogListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
