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

type TagObjectListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标签对象列表
func NewTagObjectListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagObjectListLogic {
	return &TagObjectListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 标签对象列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *TagObjectListLogic) TagObjectList(req *types.TagObjectListReq) (resp *types.TagObjectListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.TagRpc.ListObjects(l.ctx, &tagClient.ObjectListReq{
		Beid:       req.Beid,
		ObjectType: req.ObjectType,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.TagObject, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *tagObjectToTypes(item))
	}
	return &types.TagObjectListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}
