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

type ApiListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// API列表
func NewApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApiListLogic {
	return &ApiListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ApiList 转发 datasvc rpc 查询 API 分页列表
func (l *ApiListLogic) ApiList(req *types.ApiListReq) (resp *types.ApiListResp, err error) {
	// 分页默认值：page<=0 取 1，size<=0 取 10
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}

	// 逐字段映射为 rpc 请求
	rpcResp, err := l.svcCtx.DataSvcRpc.ListApis(l.ctx, &dataSvcClient.ApiListReq{
		Beid:   req.Beid,
		Status: req.Status,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐条映射 Api 列表（make 保证空列表序列化为 [] 而非 null）
	list := make([]types.ApiInfo, 0, len(rpcResp.GetList()))
	for _, api := range rpcResp.GetList() {
		list = append(list, apiToInfo(api))
	}

	return &types.ApiListResp{
		List:  list,
		Total: rpcResp.GetTotal(),
	}, nil
}
