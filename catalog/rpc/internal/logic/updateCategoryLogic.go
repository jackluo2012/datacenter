package logic

import (
	"context"
	"strings"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新分类（改名/挂父）：存在性校验 → 父分类存在性 + 成环校验（父不能是自身或自身的后代）→ 落库
func (l *UpdateCategoryLogic) UpdateCategory(in *catalog.UpdateCategoryReq) (*catalog.CategoryReply, error) {
	// 参数校验：id/平台必带，名称必填
	if in.Id <= 0 || in.Beid <= 0 || strings.TrimSpace(in.Name) == "" {
		return nil, ErrInvalidParam
	}
	// 存在性校验（带平台隔离，软删的查不到即视为不存在）
	if _, err := l.svcCtx.CategoryModel.FindOne(in.Id, in.Beid); err != nil {
		if err == model.ErrNotFound {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	// 指定父分类时校验其存在，并防挂成环
	if in.ParentId > 0 {
		if _, err := l.svcCtx.CategoryModel.FindOne(in.ParentId, in.Beid); err != nil {
			if err == model.ErrNotFound {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
		list, err := l.svcCtx.CategoryModel.ListByBeid(in.Beid)
		if err != nil {
			return nil, err
		}
		// 父分类落在自身子树内会形成环，拒绝
		if in.ParentId == in.Id || inSubtree(list, in.Id, in.ParentId) {
			return nil, ErrCategoryCircle
		}
	}
	if err := l.svcCtx.CategoryModel.Update(model.AssetCategory{
		Id:       in.Id,
		ParentId: in.ParentId,
		Name:     strings.TrimSpace(in.Name),
		Beid:     in.Beid,
	}); err != nil {
		l.Errorf("更新分类失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &catalog.CategoryReply{
		Category: &catalog.AssetCategory{
			Id:       in.Id,
			ParentId: in.ParentId,
			Name:     in.Name,
			Beid:     in.Beid,
		},
	}, nil
}

// inSubtree 判断 target 是否在以 root 为根的子树内（纯函数，便于单元测试）。
// all 为该平台下全部未删除分类；按 parent_id 建子树索引后从 root 广搜，命中 target 即成环
func inSubtree(all []*model.AssetCategory, root, target int64) bool {
	if root == target {
		return true
	}
	children := make(map[int64][]int64, len(all))
	for _, c := range all {
		children[c.ParentId] = append(children[c.ParentId], c.Id)
	}
	// 广搜 root 子树，访问集合防 parent 环脏数据导致死循环
	visited := make(map[int64]struct{}, len(all))
	queue := []int64{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if _, ok := visited[cur]; ok {
			continue
		}
		visited[cur] = struct{}{}
		for _, child := range children[cur] {
			if child == target {
				return true
			}
			queue = append(queue, child)
		}
	}
	return false
}
