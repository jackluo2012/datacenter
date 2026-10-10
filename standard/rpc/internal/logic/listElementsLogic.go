package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListElementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListElementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListElementsLogic {
	return &ListElementsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据元列表（分页 + keyword 对 element_code/element_name 模糊匹配）
func (l *ListElementsLogic) ListElements(in *standard.ElementListReq) (*standard.ElementListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.ElementModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	elements := make([]*standard.DataElement, 0, len(list))
	for _, e := range list {
		elements = append(elements, elementToPb(e))
	}
	return &standard.ElementListReply{List: elements, Total: total}, nil
}
