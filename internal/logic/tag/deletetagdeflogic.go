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

type DeleteTagDefLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除标签定义
func NewDeleteTagDefLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagDefLogic {
	return &DeleteTagDefLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除标签定义: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteTag 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteTagDefLogic) DeleteTagDef(req *types.TagDefIdReq) error {
	_, err := l.svcCtx.TagRpc.DeleteTag(l.ctx, &tagClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
