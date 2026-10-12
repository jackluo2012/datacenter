// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"context"

	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DsCallLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// API 调用日志列表
func NewDsCallLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DsCallLogListLogic {
	return &DsCallLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DsCallLogList 转发 datasvc rpc 查询 API 调用日志分页列表:
// 页码默认值处理(page<=0 取 1, size<=0 取 10)后逐字段映射转发;
// api_id/app_id/status 为空值表示不按该维度过滤, 直接透传由 rpc 侧处理
func (l *DsCallLogListLogic) DsCallLogList(req *types.CallLogListReq) (resp *types.CallLogListResp, err error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}

	rpcResp, err := l.svcCtx.DataSvcRpc.ListCallLogs(l.ctx, &dataSvcClient.CallLogListReq{
		Beid:   req.Beid,
		ApiId:  req.ApiId,
		AppId:  req.AppId,
		Status: req.Status,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐条映射调用日志（make 保证空列表序列化为 [] 而非 null）
	list := make([]types.DsApiCallLog, 0, len(rpcResp.GetList()))
	for _, item := range rpcResp.GetList() {
		list = append(list, callLogToTypes(item))
	}

	return &types.CallLogListResp{
		List:  list,
		Total: rpcResp.GetTotal(),
	}, nil
}
