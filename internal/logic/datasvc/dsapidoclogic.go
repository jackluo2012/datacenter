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

type DsApiDocLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// API 文档导出
func NewDsApiDocLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DsApiDocLogic {
	return &DsApiDocLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DsApiDoc 转发 datasvc rpc 导出 API 文档（Markdown + OpenAPI 片段）:
// path/beid 逐字段映射转发, 四个文档字段逐字段转出
func (l *DsApiDocLogic) DsApiDoc(req *types.DsApiDocReq) (resp *types.DsApiDocResp, err error) {
	reply, err := l.svcCtx.DataSvcRpc.GetApiDoc(l.ctx, &dataSvcClient.PathReq{
		Path: req.Path,
		Beid: req.Beid,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	return &types.DsApiDocResp{
		ApiName:     reply.GetApiName(),
		Path:        reply.GetPath(),
		Markdown:    reply.GetMarkdown(),
		OpenapiJson: reply.GetOpenapiJson(),
	}, nil
}
