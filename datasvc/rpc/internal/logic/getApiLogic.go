package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiLogic {
	return &GetApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API详情（按 beid+path 查询）
func (l *GetApiLogic) GetApi(in *datasvc.PathReq) (*datasvc.ApiReply, error) {
	api, err := l.svcCtx.ApiModel.FindByPath(in.Beid, strings.TrimSpace(in.Path))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	return &datasvc.ApiReply{Api: apiToPb(api)}, nil
}
