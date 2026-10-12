// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/theme/rpc/themeClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ThemeDomainListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 主题域列表
func NewThemeDomainListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThemeDomainListLogic {
	return &ThemeDomainListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 主题域列表: rpc ListDomains 入参为 IdReq 且无分页字段, 一次返回全量主题域, 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *ThemeDomainListLogic) ThemeDomainList(req *types.ThemeIdReq) (resp *types.ThemeDomainListResp, err error) {
	reply, err := l.svcCtx.ThemeRpc.ListDomains(l.ctx, &themeClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.ThemeDomainInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *themeDomainToTypes(item))
	}
	return &types.ThemeDomainListResp{
		List: list,
	}, nil
}
