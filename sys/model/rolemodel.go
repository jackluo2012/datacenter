package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	roleTable = "`sys_role`"
	// roleInsertFieldSize 插入字段数（与 roleInsertColumns 对应）
	roleInsertFieldSize = 5
	// roleInsertColumns 插入列（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	roleInsertColumns = "`role_key`,`role_name`,`status`,`remark`,`beid`"
	// roleRows 查询列（7 列与 Role 7 字段一一对应，显式列清单规避 go-zero RawFieldNames 全列扫描错位）
	roleRows = "`id`,`role_key`,`role_name`,`status`,`remark`,`beid`,`created_at`"
	// roleMenuTable 角色菜单关联表
	roleMenuTable = "`sys_role_menu`"
)

// 角色（sys_role，唯一键 uk_beid_key）
type (
	// Role 角色行
	Role struct {
		Id        int64     `db:"id"`
		RoleKey   string    `db:"role_key"` // 权限字符（如 admin/operator/viewer）
		RoleName  string    `db:"role_name"`
		Status    int64     `db:"status"` // 1.启用 0.停用
		Remark    string    `db:"remark"`
		Beid      int64     `db:"beid"`
		CreatedAt time.Time `db:"created_at"`
	}

	// RoleModel 角色表操作接口
	// 写方法带 session 参数：DeleteRole 需在事务内联动角色菜单绑定清理
	RoleModel interface {
		// Insert 新增角色（session 非空时在事务内执行），返回自增 id
		Insert(ctx context.Context, session sqlx.Session, data *Role) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Role, error)
		// FindByKey 按 beid + role_key 查询（创建/修改时查重用）
		FindByKey(ctx context.Context, beid int64, roleKey string) (*Role, error)
		// Update 更新角色基础字段（按 id 定位，role_key 可改；session 非空时在事务内执行）
		Update(ctx context.Context, session sqlx.Session, data *Role) error
		// Delete 删除角色（按 id 定位；session 非空时在事务内执行）
		Delete(ctx context.Context, session sqlx.Session, id, beid int64) error
		// FindList 条件分页查询：beid 必带，keyword 对 role_key/role_name 模糊匹配；返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*Role, int64, error)
		// GetMenuIds 查角色菜单 id 集（升序）
		GetMenuIds(ctx context.Context, roleId int64) ([]int64, error)
		// SetRoleMenus 重置角色菜单绑定（事务内先删后插，全量整存；menuIds 去重滤非法值）
		SetRoleMenus(ctx context.Context, session sqlx.Session, roleId int64, menuIds []int64) error
		// AddRoleMenus 幂等追加角色菜单绑定（insert ignore，种子补绑用，不清既有绑定）
		AddRoleMenus(ctx context.Context, roleId int64, menuIds []int64) error
		// CountByMenu 统计引用某菜单的角色数（菜单删除前引用校验）
		CountByMenu(ctx context.Context, menuId int64) (int64, error)
		// RemoveMenuRef 删除某菜单的全部角色引用（菜单删除时同事务清理）
		RemoveMenuRef(ctx context.Context, session sqlx.Session, menuId int64) error
		// FindKeysByUser 查用户持有的角色 key 集（join sys_user_role；GetInfo roles 字段）
		FindKeysByUser(ctx context.Context, userId, beid int64) ([]string, error)
	}

	defaultRoleModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewRoleModel 创建角色 model
func NewRoleModel(conn sqlx.SqlConn) RoleModel {
	return &defaultRoleModel{
		conn:  conn,
		table: roleTable,
	}
}

// Insert 新增角色
func (m *defaultRoleModel) Insert(ctx context.Context, session sqlx.Session, data *Role) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, roleInsertColumns, placeholders(roleInsertFieldSize))
	ret, err := execCtx(ctx, m.conn, session, query,
		data.RoleKey, data.RoleName, data.Status, data.Remark, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询角色
func (m *defaultRoleModel) FindOne(ctx context.Context, id, beid int64) (*Role, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", roleRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	return scanRole(resp, err)
}

// FindByKey 按 beid + role_key 查询角色（查重；beid<=0 时全局匹配）
func (m *defaultRoleModel) FindByKey(ctx context.Context, beid int64, roleKey string) (*Role, error) {
	query := fmt.Sprintf("select %s from %s where `role_key` = ?", roleRows, m.table)
	args := []any{roleKey}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	return scanRole(resp, err)
}

// scanRole 统一处理单行扫描结果：未找到归一化为 ErrNotFound
func scanRole(resp Role, err error) (*Role, error) {
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新角色基础字段
func (m *defaultRoleModel) Update(ctx context.Context, session sqlx.Session, data *Role) error {
	query := fmt.Sprintf("update %s set `role_key` = ?, `role_name` = ?, `status` = ?, `remark` = ? "+
		"where `id` = ?", m.table)
	args := []any{data.RoleKey, data.RoleName, data.Status, data.Remark, data.Id}
	if data.Beid > 0 {
		query += " and `beid` = ?"
		args = append(args, data.Beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// Delete 删除角色（菜单绑定由调用方在同一事务内通过 SetRoleMenus(roleId, nil) 清理）
func (m *defaultRoleModel) Delete(ctx context.Context, session sqlx.Session, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// FindList 角色列表：keyword 模糊匹配 role_key/role_name
func (m *defaultRoleModel) FindList(ctx context.Context, beid int64, keyword string,
	page, size int64) ([]*Role, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`role_key` like ? or `role_name` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		roleRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Role
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetMenuIds 查角色菜单 id 集
func (m *defaultRoleModel) GetMenuIds(ctx context.Context, roleId int64) ([]int64, error) {
	query := fmt.Sprintf("select `menu_id` from %s where `role_id` = ? order by `menu_id`", roleMenuTable)
	var menuIds []int64
	if err := m.conn.QueryRowsCtx(ctx, &menuIds, query, roleId); err != nil {
		return nil, err
	}
	return menuIds, nil
}

// SetRoleMenus 重置角色菜单绑定：先删后插（整存），menuIds 为空即清空绑定
func (m *defaultRoleModel) SetRoleMenus(ctx context.Context, session sqlx.Session, roleId int64, menuIds []int64) error {
	// 空绑定仅清空，无需事务包裹
	if session == nil {
		return m.conn.TransactCtx(ctx, func(txCtx context.Context, tx sqlx.Session) error {
			return m.SetRoleMenus(txCtx, tx, roleId, menuIds)
		})
	}
	delQuery := fmt.Sprintf("delete from %s where `role_id` = ?", roleMenuTable)
	if _, err := session.ExecCtx(ctx, delQuery, roleId); err != nil {
		return err
	}
	menuIds = dedupeInt64(menuIds)
	if len(menuIds) == 0 {
		return nil
	}
	pairs := strings.TrimSuffix(strings.Repeat("(?, ?), ", len(menuIds)), ", ")
	insQuery := fmt.Sprintf("insert into %s (`role_id`,`menu_id`) values %s",
		roleMenuTable, pairs)
	args := make([]any, 0, len(menuIds)*2)
	for _, id := range menuIds {
		args = append(args, roleId, id)
	}
	_, err := session.ExecCtx(ctx, insQuery, args...)
	return err
}

// AddRoleMenus 幂等追加角色菜单绑定（insert ignore：已存在的绑定不受影响）
func (m *defaultRoleModel) AddRoleMenus(ctx context.Context, roleId int64, menuIds []int64) error {
	menuIds = dedupeInt64(menuIds)
	if len(menuIds) == 0 {
		return nil
	}
	// 两列值行：values (?, ?), (?, ?) ...（role_id + menu_id 成对，数量与 menuIds 一致）
	pairs := strings.TrimSuffix(strings.Repeat("(?, ?), ", len(menuIds)), ", ")
	query := fmt.Sprintf("insert ignore into %s (`role_id`,`menu_id`) values %s",
		roleMenuTable, pairs)
	args := make([]any, 0, len(menuIds)*2)
	for _, id := range menuIds {
		args = append(args, roleId, id)
	}
	logx.Errorf("[定位] AddRoleMenus roleId=%d menuIds=%d sql占位=%d args=%d sql=%s",
		roleId, len(menuIds), strings.Count(query, "?"), len(args), query[:min(120, len(query))])
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// CountByMenu 统计引用菜单的角色数（菜单删除前校验：被引用拒删）
func (m *defaultRoleModel) CountByMenu(ctx context.Context, menuId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `menu_id` = ?", roleMenuTable)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, menuId); err != nil {
		return 0, err
	}
	return total, nil
}

// RemoveMenuRef 删除某菜单的全部角色引用（菜单删除时同事务级联清理）
func (m *defaultRoleModel) RemoveMenuRef(ctx context.Context, session sqlx.Session, menuId int64) error {
	query := fmt.Sprintf("delete from %s where `menu_id` = ?", roleMenuTable)
	_, err := execCtx(ctx, m.conn, session, query, menuId)
	return err
}

// FindKeysByUser 查用户角色 key 集（GetInfo roles 字段；beid>0 时限定租户内角色）
func (m *defaultRoleModel) FindKeysByUser(ctx context.Context, userId, beid int64) ([]string, error) {
	query := fmt.Sprintf("select distinct r.`role_key` from %s r "+
		"join `sys_user_role` ur on ur.`role_id` = r.`id` where ur.`user_id` = ?", m.table)
	args := []any{userId}
	if beid > 0 {
		query += " and r.`beid` = ?"
		args = append(args, beid)
	}
	query += " order by r.`role_key`"
	var keys []string
	if err := m.conn.QueryRowsCtx(ctx, &keys, query, args...); err != nil {
		return nil, err
	}
	return keys, nil
}
