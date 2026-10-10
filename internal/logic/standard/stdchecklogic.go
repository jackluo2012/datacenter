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

type StdCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标准核对
func NewStdCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StdCheckLogic {
	return &StdCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 标准核对: types.StdCheckReq 逐字段映射为 pb CheckReq, 转发后把核对报告转为网关类型返回。
// 网关 StdCheckResp.Report 为值类型, report 为 nil 时兜底为零值结构
func (l *StdCheckLogic) StdCheck(req *types.StdCheckReq) (resp *types.StdCheckResp, err error) {
	reply, err := l.svcCtx.StandardRpc.CheckTable(l.ctx, &standardClient.CheckReq{
		ModelId:      req.ModelId,
		DatasourceId: req.DatasourceId,
		Beid:         req.Beid,
		DbName:       req.DbName,
		TableName:    req.TableName,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.StdCheckResp{
		Report: *checkReportToTypes(reply.GetReport()),
	}, nil
}
