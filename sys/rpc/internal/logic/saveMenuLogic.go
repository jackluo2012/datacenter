package logic

import (
	"context"
	"errors"
	"regexp"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
	"datacenter/sys/rpc/sys"

	"github.com/zeromicro/go-zero/core/logx"
)

// permPattern 按钮权限标识格式：模块:动作（段为字母开头，可含数字/下划线，驼峰合法——
// 对齐 RuoYi 惯例如 sys:user:resetPwd；至少一段冒号，三段形式同样匹配）
var permPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(?::[a-zA-Z][a-zA-Z0-9_]*)+$`)

type SaveMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMenuLogic {
	return &SaveMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存菜单（id>0 更新，否则新建）：
//   - menu_type 白名单（1.目录 2.菜单 3.按钮）
//   - perm 非空时校验 模块:动作 格式
//   - 父节点校验：存在性 + 类型匹配（目录←目录；菜单←目录；按钮←菜单）+ 更新时防环
func (l *SaveMenuLogic) SaveMenu(in *sys.SaveMenuReq) (*sys.MenuReply, error) {
	if in.Id > 0 {
		if _, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrMenuNotFound
			}
			return nil, err
		}
	}
	if err := l.validate(in); err != nil {
		return nil, err
	}

	// 新建缺省可见（proto3 缺省值 0 与"隐藏"歧义，新建默认显示；更新按传入值整存）
	visible := int64(in.Visible)
	if in.Id == 0 {
		visible = visibleOrDefault(visible)
	}
	data := &model.Menu{
		Id:        in.Id,
		ParentId:  in.ParentId,
		Name:      in.Name,
		Path:      in.Path,
		Component: in.Component,
		Sort:      int64(in.Sort),
		Icon:      in.Icon,
		MenuType:  int64(in.MenuType),
		Perm:      in.Perm,
		Visible:   visible,
		Beid:      in.Beid,
	}
	var err error
	if in.Id > 0 {
		err = l.svcCtx.MenuModel.Update(l.ctx, nil, data)
	} else {
		data.Id, err = l.svcCtx.MenuModel.Insert(l.ctx, nil, data)
	}
	if err != nil {
		return nil, err
	}

	menu, err := l.svcCtx.MenuModel.FindOne(l.ctx, data.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	l.Infof("保存菜单成功 id=%d name=%s type=%d", data.Id, in.Name, in.MenuType)
	return &sys.MenuReply{Menu: menuToPb(menu)}, nil
}

// validate 菜单保存校验（类型白名单/名称/perm 格式/父节点合法性/防环）
func (l *SaveMenuLogic) validate(in *sys.SaveMenuReq) error {
	switch in.MenuType {
	case model.MenuTypeDir, model.MenuTypeMenu, model.MenuTypeButton:
	default:
		return ErrMenuTypeInvalid
	}
	if in.Name == "" {
		return ErrNameRequired
	}
	// perm 非空时须为 模块:动作 格式
	if in.Perm != "" && !permPattern.MatchString(in.Perm) {
		return ErrMenuPermInvalid
	}
	if in.ParentId == 0 {
		return nil
	}
	// 更新分支防环：父不能是自身
	if in.Id > 0 && in.ParentId == in.Id {
		return ErrMenuParentInvalid
	}
	parent, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.ParentId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return ErrMenuParentInvalid
		}
		return err
	}
	// 父子类型匹配：目录父=目录；菜单父=目录；按钮父=菜单
	switch in.MenuType {
	case model.MenuTypeDir, model.MenuTypeMenu:
		if parent.MenuType != model.MenuTypeDir {
			return ErrMenuParentInvalid
		}
	case model.MenuTypeButton:
		if parent.MenuType != model.MenuTypeMenu {
			return ErrMenuParentInvalid
		}
	}
	// 更新为按钮时不得携带子节点（按钮是叶子，防类型漂移破坏树语义）
	if in.Id > 0 && in.MenuType == model.MenuTypeButton {
		children, err := l.svcCtx.MenuModel.CountChildren(l.ctx, in.Beid, in.Id)
		if err != nil {
			return err
		}
		if children > 0 {
			return ErrMenuHasChildren
		}
	}
	// 更新分支防环：沿父链上溯不得回到自身（父链成环会让子树不可达）
	if in.Id > 0 {
		if err = l.ensureNoCycle(in.Id, in.Beid, in.ParentId); err != nil {
			return err
		}
	}
	return nil
}

// ensureNoCycle 沿新父链上溯，遇到自身说明会把自身子树挂到自己下面（成环）
func (l *SaveMenuLogic) ensureNoCycle(id, beid, parentId int64) error {
	seen := make(map[int64]struct{}, 8)
	cur := parentId
	for cur != 0 {
		if cur == id {
			return ErrMenuParentInvalid
		}
		if _, ok := seen[cur]; ok {
			// 既有数据已存在环：不再上溯（不因历史脏数据阻塞保存）
			return nil
		}
		seen[cur] = struct{}{}
		parent, err := l.svcCtx.MenuModel.FindOne(l.ctx, cur, beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil // 父链断裂（孤儿），不构成环
			}
			return err
		}
		cur = parent.ParentId
	}
	return nil
}

// visibleOrDefault 可见性缺省：未传（proto3 缺省 0）默认显示，仅新建分支使用
func visibleOrDefault(visible int64) int64 {
	if visible == 0 {
		return 1
	}
	return visible
}
