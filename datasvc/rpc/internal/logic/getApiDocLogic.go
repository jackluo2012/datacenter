package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiDocLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApiDocLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiDocLogic {
	return &GetApiDocLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API 文档门户：按 path 取 API，导出 Markdown 文档 + OpenAPI 3.0 JSON 片段
// （两个生成器均为纯函数，见 apiDocGen.go；文档不限制 API 状态，draft 亦可预览）
func (l *GetApiDocLogic) GetApiDoc(in *datasvc.PathReq) (*datasvc.ApiDocReply, error) {
	path := strings.TrimSpace(in.Path)
	if path == "" {
		return nil, fmt.Errorf("%w: path 不能为空", ErrParamInvalid)
	}
	api, err := l.svcCtx.ApiModel.FindByPath(in.Beid, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	markdown := BuildApiMarkdown(api)
	openapiJson, err := BuildOpenApiJSON(api)
	if err != nil {
		return nil, fmt.Errorf("生成 OpenAPI 文档失败: %v", err)
	}
	return &datasvc.ApiDocReply{
		ApiName:     api.Name,
		Path:        api.Path,
		Markdown:    markdown,
		OpenapiJson: openapiJson,
	}, nil
}
