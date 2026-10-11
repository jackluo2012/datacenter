package logic

import (
	"context"

	"datacenter/sys/model"
	"datacenter/sys/rpc/internal/svc"
)

// seedRBAC 播种内置 RBAC 数据（幂等可重入）：
// 系统管理+数据中台菜单树、内置角色 admin/viewer、内置账号 admin/admin123
func seedRBAC(ctx context.Context, svcCtx *svc.ServiceContext, beid int64) error {
	return model.SeedRBAC(ctx, svcCtx.UserModel, svcCtx.RoleModel, svcCtx.MenuModel, beid)
}
