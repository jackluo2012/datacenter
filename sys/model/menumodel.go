package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	menuTable = "`sys_menu`"
	// menuInsertFieldSize 插入字段数（与 menuInsertColumns 对应）
	menuInsertFieldSize = 10
	// menuInsertColumns 插入列（不含自增 id；sys_menu 无时间列），与 Insert 传参顺序一一对应
	menuInsertColumns = "`parent_id`,`name`,`path`,`component`,`sort`,`icon`,`menu_type`,`perm`,`visible`,`beid`"
	// menuRows 查询列（11 列与 Menu 11 字段一一对应，显式列清单规避 go-zero RawFieldNames 全列扫描错位）
	menuRows = "`id`,`parent_id`,`name`,`path`,`component`,`sort`,`icon`,`menu_type`,`perm`,`visible`,`beid`"
	// menuRowsAliased join 查询用的带表别名列清单（别名固定为 m）
	menuRowsAliased = "`m`.`id`,`m`.`parent_id`,`m`.`name`,`m`.`path`,`m`.`component`,`m`.`sort`," +
		"`m`.`icon`,`m`.`menu_type`,`m`.`perm`,`m`.`visible`,`m`.`beid`"
)

// 菜单（sys_menu：目录/菜单/按钮三级，parent_id=0 为根）
type (
	// Menu 菜单行
	Menu struct {
		Id        int64  `db:"id"`
		ParentId  int64  `db:"parent_id"`
		Name      string `db:"name"`
		Path      string `db:"path"`      // 前端路由路径（目录/菜单用，如 /system、user）
		Component string `db:"component"` // 前端组件路径（如 system/UserList → web/src/views/system/UserList.vue）
		Sort      int64  `db:"sort"`
		Icon      string `db:"icon"`
		MenuType  int64  `db:"menu_type"` // 1.目录 2.菜单 3.按钮
		Perm      string `db:"perm"`      // 按钮权限标识（如 sys:user:create）
		Visible   int64  `db:"visible"`   // 1.显示 0.隐藏
		Beid      int64  `db:"beid"`
	}

	// MenuModel 菜单表操作接口
	// 写方法带 session 参数：DeleteMenu 需在事务内联动角色菜单引用清理
	MenuModel interface {
		// Insert 新增菜单（session 非空时在事务内执行），返回自增 id
		Insert(ctx context.Context, session sqlx.Session, data *Menu) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Menu, error)
		// FindOneByName 按 beid + 父节点 + 名称定位菜单（父子路径下名称唯一，种子幂等定位用）
		FindOneByName(ctx context.Context, beid, parentId int64, name string) (*Menu, error)
		// Update 更新菜单全字段（按 id 定位；session 非空时在事务内执行）
		Update(ctx context.Context, session sqlx.Session, data *Menu) error
		// Delete 删除菜单（按 id 定位；session 非空时在事务内执行）
		Delete(ctx context.Context, session sqlx.Session, id, beid int64) error
		// FindAll 全量菜单（sort,id 升序，内存建树数据源；beid>0 时追加多租户过滤）
		FindAll(ctx context.Context, beid int64) ([]*Menu, error)
		// FindByUid 查用户可见菜单（join sys_user_role + sys_role_menu，distinct 去重多角色重叠；
		// 含按钮，建树后由 logic 层剔除；beid>0 时追加多租户过滤）
		FindByUid(ctx context.Context, uid, beid int64) ([]*Menu, error)
		// FindPermsByUid 查用户按钮权限标识集（menu_type=按钮 且 perm 非空，distinct 去重）
		FindPermsByUid(ctx context.Context, uid, beid int64) ([]string, error)
		// HasPerm 判断用户是否持有某权限标识（网关/下游服务鉴权用）
		HasPerm(ctx context.Context, uid, beid int64, perm string) (bool, error)
		// CountChildren 统计子菜单数（菜单删除前校验：有子拒删）
		CountChildren(ctx context.Context, beid, parentId int64) (int64, error)
	}

	defaultMenuModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewMenuModel 创建菜单 model
func NewMenuModel(conn sqlx.SqlConn) MenuModel {
	return &defaultMenuModel{
		conn:  conn,
		table: menuTable,
	}
}

// Insert 新增菜单
func (m *defaultMenuModel) Insert(ctx context.Context, session sqlx.Session, data *Menu) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, menuInsertColumns, placeholders(menuInsertFieldSize))
	ret, err := execCtx(ctx, m.conn, session, query,
		data.ParentId, data.Name, data.Path, data.Component, data.Sort,
		data.Icon, data.MenuType, data.Perm, data.Visible, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询菜单
func (m *defaultMenuModel) FindOne(ctx context.Context, id, beid int64) (*Menu, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", menuRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp Menu
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	return scanMenu(resp, err)
}

// FindOneByName 按 beid + 父节点 + 名称定位菜单（种子幂等：已存在则跳过插入）
func (m *defaultMenuModel) FindOneByName(ctx context.Context, beid, parentId int64, name string) (*Menu, error) {
	query := fmt.Sprintf("select %s from %s where `parent_id` = ? and `name` = ?", menuRows, m.table)
	args := []any{parentId, name}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	query += " order by `id` limit 1"
	var resp Menu
	err := m.conn.QueryRowCtx(ctx, &resp, query, args...)
	return scanMenu(resp, err)
}

// scanMenu 统一处理单行扫描结果：未找到归一化为 ErrNotFound
func scanMenu(resp Menu, err error) (*Menu, error) {
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新菜单全字段
func (m *defaultMenuModel) Update(ctx context.Context, session sqlx.Session, data *Menu) error {
	query := fmt.Sprintf("update %s set `parent_id` = ?, `name` = ?, `path` = ?, `component` = ?, `sort` = ?, "+
		"`icon` = ?, `menu_type` = ?, `perm` = ?, `visible` = ? where `id` = ?", m.table)
	args := []any{data.ParentId, data.Name, data.Path, data.Component, data.Sort,
		data.Icon, data.MenuType, data.Perm, data.Visible, data.Id}
	if data.Beid > 0 {
		query += " and `beid` = ?"
		args = append(args, data.Beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// Delete 删除菜单（角色引用由调用方在同一事务内通过 RemoveMenuRef 清理）
func (m *defaultMenuModel) Delete(ctx context.Context, session sqlx.Session, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// FindAll 全量菜单（内存建树数据源）
func (m *defaultMenuModel) FindAll(ctx context.Context, beid int64) ([]*Menu, error) {
	query := fmt.Sprintf("select %s from %s", menuRows, m.table)
	args := []any{}
	if beid > 0 {
		query += " where `beid` = ?"
		args = append(args, beid)
	}
	query += " order by `sort`, `id`"
	var list []*Menu
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// findMenuJoin 基于用户角色 join 的菜单查询公共部分（distinct 去重多角色重叠菜单）
const findMenuJoin = " from `sys_menu` m " +
	"join `sys_role_menu` rm on rm.`menu_id` = m.`id` " +
	"join `sys_user_role` ur on ur.`role_id` = rm.`role_id` where ur.`user_id` = ?"

// FindByUid 查用户可见菜单（含按钮，建树后由 logic 层剔除按钮）
func (m *defaultMenuModel) FindByUid(ctx context.Context, uid, beid int64) ([]*Menu, error) {
	query := "select distinct " + menuRowsAliased + findMenuJoin
	args := []any{uid}
	if beid > 0 {
		query += " and m.`beid` = ?"
		args = append(args, beid)
	}
	query += " order by m.`sort`, m.`id`"
	var list []*Menu
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// FindPermsByUid 查用户按钮权限标识集
func (m *defaultMenuModel) FindPermsByUid(ctx context.Context, uid, beid int64) ([]string, error) {
	query := "select distinct m.`perm`" + findMenuJoin +
		" and m.`menu_type` = ? and m.`perm` <> ''"
	args := []any{uid, MenuTypeButton}
	if beid > 0 {
		query += " and m.`beid` = ?"
		args = append(args, beid)
	}
	query += " order by m.`perm`"
	var perms []string
	if err := m.conn.QueryRowsCtx(ctx, &perms, query, args...); err != nil {
		return nil, err
	}
	return perms, nil
}

// HasPerm 判断用户是否持有某权限标识
func (m *defaultMenuModel) HasPerm(ctx context.Context, uid, beid int64, perm string) (bool, error) {
	if strings.TrimSpace(perm) == "" {
		return false, nil
	}
	query := "select count(1)" + findMenuJoin + " and m.`perm` = ?"
	args := []any{uid, perm}
	if beid > 0 {
		query += " and m.`beid` = ?"
		args = append(args, beid)
	}
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return false, err
	}
	return total > 0, nil
}

// CountChildren 统计子菜单数
func (m *defaultMenuModel) CountChildren(ctx context.Context, beid, parentId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `parent_id` = ?", m.table)
	args := []any{parentId}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}
