package logic

import (
	"context"

	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListThemesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListThemesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListThemesLogic {
	return &ListThemesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListThemes 主题分页列表：domain_id>0 时按域过滤
func (l *ListThemesLogic) ListThemes(in *theme.ThemeListReq) (*theme.ThemeListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.ThemeModel.FindList(l.ctx, in.Beid, in.DomainId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*theme.ThemeInfo, 0, len(list))
	for _, t := range list {
		pbList = append(pbList, toPbTheme(t))
	}
	return &theme.ThemeListReply{List: pbList, Total: total}, nil
}
