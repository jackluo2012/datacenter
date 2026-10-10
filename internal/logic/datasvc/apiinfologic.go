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

type ApiInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// API详情
func NewApiInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiInfoLogic {
	return &ApiInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ApiInfo 按执行路径转发 datasvc rpc 查询 API 详情
func (l *ApiInfoLogic) ApiInfo(req *types.ApiPathReq) (resp *types.ApiInfo, err error) {
	// 逐字段映射为 rpc 请求（按 path + beid 定位）
	rpcResp, err := l.svcCtx.DataSvcRpc.GetApi(l.ctx, &dataSvcClient.PathReq{
		Path: req.Path,
		Beid: req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐字段映射 rpc 返回的 Api
	info := apiToInfo(rpcResp.GetApi())
	return &info, nil
}
