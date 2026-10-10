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

type CreateApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建数据API
func NewCreateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateApiLogic {
	return &CreateApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateApi 转发 datasvc rpc 创建数据 API，返回含 id 的 ApiInfo
func (l *CreateApiLogic) CreateApi(req *types.CreateApiReq) (resp *types.ApiInfo, err error) {
	// mode 未传时默认 sql（proto 约定：wizard/sql/register 三种模式）
	mode := req.Mode
	if mode == "" {
		mode = "sql"
	}

	// 逐字段映射为 rpc 请求（网关契约无 resp_sample_json，置空由 rpc 侧处理）
	rpcResp, err := l.svcCtx.DataSvcRpc.CreateApi(l.ctx, &dataSvcClient.CreateApiReq{
		Name:         req.Name,
		Path:         req.Path,
		Mode:         mode,
		DatasourceId: req.DatasourceId,
		SqlText:      req.SqlText,
		ParamsJson:   req.ParamsJson,
		CacheTtl:     req.CacheTtl,
		Owner:        req.Owner,
		Beid:         req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 返回创建结果（含 id）
	info := apiToInfo(rpcResp.GetApi())
	return &info, nil
}
