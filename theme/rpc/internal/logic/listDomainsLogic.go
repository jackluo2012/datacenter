package logic

import (
	"context"

	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDomainsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDomainsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDomainsLogic {
	return &ListDomainsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDomains 主题域全量列表（按 id 倒序；IdReq 无分页参数，域数量少不分页）
func (l *ListDomainsLogic) ListDomains(in *theme.IdReq) (*theme.DomainListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, err := l.svcCtx.DomainModel.FindList(l.ctx, in.Beid)
	if err != nil {
		return nil, err
	}
	pbList := make([]*theme.ThemeDomain, 0, len(list))
	for _, d := range list {
		pbList = append(pbList, toPbDomain(d))
	}
	return &theme.DomainListReply{List: pbList, Total: int64(len(pbList))}, nil
}
