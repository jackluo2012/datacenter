// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagObjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除标签对象（对象下仍有标签定义或人群时拒绝）
func NewDeleteTagObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagObjectLogic {
	return &DeleteTagObjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteTagObject: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteObject 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteTagObjectLogic) DeleteTagObject(req *types.TagDefIdReq) error {
	_, err := l.svcCtx.TagRpc.DeleteObject(l.ctx, &tagClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
