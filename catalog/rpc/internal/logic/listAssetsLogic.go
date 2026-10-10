package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetsLogic {
	return &ListAssetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 资产列表（分页 + 分类过滤 + keyword 模糊匹配 name/comment）
func (l *ListAssetsLogic) ListAssets(in *catalog.AssetListReq) (*catalog.AssetListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	list, total, err := l.svcCtx.AssetModel.FindList(in.Beid, in.CategoryId, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reply := make([]*catalog.Asset, 0, len(list))
	for _, a := range list {
		reply = append(reply, assetToPb(a))
	}
	return &catalog.AssetListReply{List: reply, Total: total}, nil
}
