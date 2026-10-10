package logic

import (
	"context"
	"strings"

	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListThemeModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListThemeModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListThemeModelsLogic {
	return &ListThemeModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListThemeModels 模型分页列表：theme_id>0 按主题过滤；layer 非空按层过滤
// （传入值先过白名单，防止脏过滤条件静默返回空集）
func (l *ListThemeModelsLogic) ListThemeModels(in *theme.ModelListReq) (*theme.ModelListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	layer := strings.TrimSpace(in.Layer)
	if layer != "" && !ValidLayer(layer) {
		return nil, ErrInvalidLayer
	}
	list, total, err := l.svcCtx.ModelModel.FindList(l.ctx, in.Beid, in.ThemeId, layer, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*theme.ThemeModel, 0, len(list))
	for _, m := range list {
		pbList = append(pbList, toPbThemeModel(m))
	}
	return &theme.ModelListReply{List: pbList, Total: total}, nil
}
