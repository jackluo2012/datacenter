// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package catalog

import (
	"context"

	"datacenter/catalog/rpc/catalogClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 资产列表
func NewAssetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetListLogic {
	return &AssetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssetList 转发 catalog rpc 查询资产分页列表
func (l *AssetListLogic) AssetList(req *types.AssetListReq) (resp *types.AssetListResp, err error) {
	// 分页默认值：page<=0 取 1，size<=0 取 10
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 10
	}

	// 逐字段映射为 rpc 请求
	rpcResp, err := l.svcCtx.CatalogRpc.ListAssets(l.ctx, &catalogClient.AssetListReq{
		Beid:       req.Beid,
		CategoryId: req.CategoryId,
		Keyword:    req.Keyword,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐条映射资产列表（make 保证空列表序列化为 [] 而非 null）
	list := make([]types.Asset, 0, len(rpcResp.GetList()))
	for _, a := range rpcResp.GetList() {
		list = append(list, types.Asset{
			Id:             a.GetId(),
			EntityId:       a.GetEntityId(),
			CategoryId:     a.GetCategoryId(),
			Name:           a.GetName(),
			EntityType:     a.GetEntityType(),
			Comment:        a.GetComment(),
			Classification: a.GetClassification(),
			AssetTagsJson:  a.GetAssetTagsJson(),
			Visibility:     a.GetVisibility(),
		})
	}

	return &types.AssetListResp{
		List:  list,
		Total: rpcResp.GetTotal(),
	}, nil
}
