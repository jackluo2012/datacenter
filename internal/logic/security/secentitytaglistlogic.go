// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/security/rpc/securityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecEntityTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 实体打标列表
func NewSecEntityTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecEntityTagListLogic {
	return &SecEntityTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 实体打标列表: 页码默认值处理后转发(level_code 为密级过滤, 空串表示不过滤), 业务错误直接透传。
// 空列表仍 make 出空切片, 保证 JSON 序列化为 [] 而非 null。
func (l *SecEntityTagListLogic) SecEntityTagList(req *types.EntityTagListReq) (resp *types.EntityTagListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.SecurityRpc.ListEntityTags(l.ctx, &securityClient.EntityTagListReq{
		Beid:      req.Beid,
		LevelCode: req.LevelCode,
		Page:      page,
		Size:      size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.SecEntityTag, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *entityTagToTypes(item))
	}
	return &types.EntityTagListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
