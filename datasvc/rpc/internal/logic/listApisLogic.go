package logic

import (
	"context"
	"strings"

	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListApisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListApisLogic {
	return &ListApisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API列表（beid+可选 status 分页，分页参数非法时在 model 层兜底）
func (l *ListApisLogic) ListApis(in *datasvc.ApiListReq) (*datasvc.ApiListReply, error) {
	list, total, err := l.svcCtx.ApiModel.List(in.Beid, strings.TrimSpace(in.Status), in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*datasvc.Api, 0, len(list))
	for _, api := range list {
		pbList = append(pbList, apiToPb(api))
	}
	return &datasvc.ApiListReply{List: pbList, Total: total}, nil
}
