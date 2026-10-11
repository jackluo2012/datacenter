package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	userTable = "`sys_user`"
	// userInsertFieldSize 插入字段数（与 userInsertColumns 对应）
	userInsertFieldSize = 8
	// userInsertColumns 插入列（不含自增 id 与时间列），与 Insert 传参顺序一一对应
	userInsertColumns = "`username`,`password`,`nickname`,`mobile`,`email`,`status`,`remark`,`beid`"
	// userRows 查询列（11 列与 User 11 字段一一对应，显式列清单规避 go-zero RawFieldNames 全列扫描错位）
	userRows = "`id`,`username`,`password`,`nickname`,`mobile`,`email`,`status`,`remark`,`beid`,`created_at`,`updated_at`"
	// userRoleTable 用户角色关联表
	userRoleTable = "`sys_user_role`"
)

// 系统用户（sys_user，管理面账号，与 C 端 base_member 分离）
type (
	// User 用户行
	User struct {
		Id        int64     `db:"id"`
		Username  string    `db:"username"`
		Password  string    `db:"password"` // bcrypt 哈希，禁止外透
		Nickname  string    `db:"nickname"`
		Mobile    string    `db:"mobile"`
		Email     string    `db:"email"`
		Status    int64     `db:"status"` // 1.启用 0.停用
		Remark    string    `db:"remark"`
		Beid      int64     `db:"beid"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	// UserRoleId 用户-角色关联行（批量聚合扫描用）
	UserRoleId struct {
		UserId int64 `db:"user_id"`
		RoleId int64 `db:"role_id"`
	}

	// UserModel 用户表操作接口
	// 写方法带 session 参数：CreateUser/DeleteUser 需在事务内联动用户角色绑定（先删后插）
	UserModel interface {
		// Insert 新增用户（password 须为 bcrypt 哈希；session 非空时在事务内执行），返回自增 id
		Insert(ctx context.Context, session sqlx.Session, data *User) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*User, error)
		// FindByUsername 按登录名查询（全局唯一键 uk_username，登录与查重共用）
		FindByUsername(ctx context.Context, username string) (*User, error)
		// Update 更新用户基础字段（按 id 定位，不改 username/password；session 非空时在事务内执行）
		Update(ctx context.Context, session sqlx.Session, data *User) error
		// UpdatePwd 更新密码哈希（按 id 定位；beid>0 时追加多租户过滤；session 非空时在事务内执行）
		UpdatePwd(ctx context.Context, session sqlx.Session, id int64, passwordHash string, beid int64) error
		// Delete 删除用户（按 id 定位；session 非空时在事务内执行）
		Delete(ctx context.Context, session sqlx.Session, id, beid int64) error
		// FindList 条件分页查询：beid 必带；keyword 对 username/nickname/mobile 模糊匹配；
		// status<0 不过滤（0/1 为精确过滤）；返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, status int64, page, size int64) ([]*User, int64, error)
		// GetRoleIds 查用户角色 id 集（升序）
		GetRoleIds(ctx context.Context, userId int64) ([]int64, error)
		// GetRoleIdsBatch 批量查多个用户的角色 id 集（列表页避免 N+1 查询）
		GetRoleIdsBatch(ctx context.Context, userIds []int64) (map[int64][]int64, error)
		// SetUserRoles 重置用户角色绑定（事务内先删后插，全量整存；roleIds 去重滤非法值）
		SetUserRoles(ctx context.Context, session sqlx.Session, userId int64, roleIds []int64) error
		// AddUserRoles 幂等追加角色绑定（insert ignore，种子补绑用，不清既有绑定）
		AddUserRoles(ctx context.Context, userId int64, roleIds []int64) error
		// CountByRole 统计引用某角色的用户数（角色删除前引用校验）
		CountByRole(ctx context.Context, roleId int64) (int64, error)
	}

	defaultUserModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewUserModel 创建用户 model
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &defaultUserModel{
		conn:  conn,
		table: userTable,
	}
}

// Insert 新增用户（password 须已 bcrypt 哈希，由 logic 层负责）
func (m *defaultUserModel) Insert(ctx context.Context, session sqlx.Session, data *User) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table, userInsertColumns, placeholders(userInsertFieldSize))
	ret, err := execCtx(ctx, m.conn, session, query,
		data.Username, data.Password, data.Nickname, data.Mobile, data.Email, data.Status, data.Remark, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询用户
func (m *defaultUserModel) FindOne(ctx context.Context, id, beid int64) (*User, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", userRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	return scanUser(resp, err)
}

// FindByUsername 按登录名查询用户
func (m *defaultUserModel) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := fmt.Sprintf("select %s from %s where `username` = ? limit 1", userRows, m.table)
	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, username)
	return scanUser(resp, err)
}

// scanUser 统一处理单行扫描结果：未找到归一化为 ErrNotFound
func scanUser(resp User, err error) (*User, error) {
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新用户基础字段（username/password 不在列清单：登录名不可变，密码走 UpdatePwd）
func (m *defaultUserModel) Update(ctx context.Context, session sqlx.Session, data *User) error {
	query := fmt.Sprintf("update %s set `nickname` = ?, `mobile` = ?, `email` = ?, `status` = ?, `remark` = ? "+
		"where `id` = ?", m.table)
	args := []any{data.Nickname, data.Mobile, data.Email, data.Status, data.Remark, data.Id}
	if data.Beid > 0 {
		query += " and `beid` = ?"
		args = append(args, data.Beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// UpdatePwd 更新密码哈希（入参须为 bcrypt 哈希）
func (m *defaultUserModel) UpdatePwd(ctx context.Context, session sqlx.Session, id int64, passwordHash string, beid int64) error {
	query := fmt.Sprintf("update %s set `password` = ? where `id` = ?", m.table)
	args := []any{passwordHash, id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// Delete 删除用户（角色绑定由调用方在同一事务内通过 SetUserRoles(id, nil) 清理）
func (m *defaultUserModel) Delete(ctx context.Context, session sqlx.Session, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	_, err := execCtx(ctx, m.conn, session, query, args...)
	return err
}

// FindList 用户列表：keyword 模糊匹配 username/nickname/mobile，status>=0 精确过滤
func (m *defaultUserModel) FindList(ctx context.Context, beid int64, keyword string,
	status int64, page, size int64) ([]*User, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	page, size = normalizePage(page, size)

	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`username` like ? or `nickname` like ? or `mobile` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like, like)
	}
	if status >= 0 {
		where += " and `status` = ?"
		args = append(args, status)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 分页数据
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ?, ?",
		userRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*User
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetRoleIds 查用户角色 id 集
func (m *defaultUserModel) GetRoleIds(ctx context.Context, userId int64) ([]int64, error) {
	query := fmt.Sprintf("select `role_id` from %s where `user_id` = ? order by `role_id`", userRoleTable)
	var roleIds []int64
	if err := m.conn.QueryRowsCtx(ctx, &roleIds, query, userId); err != nil {
		return nil, err
	}
	return roleIds, nil
}

// GetRoleIdsBatch 批量查用户角色 id 集
func (m *defaultUserModel) GetRoleIdsBatch(ctx context.Context, userIds []int64) (map[int64][]int64, error) {
	result := make(map[int64][]int64, len(userIds))
	if len(userIds) == 0 {
		return result, nil
	}
	query := fmt.Sprintf("select `user_id`,`role_id` from %s where `user_id` in (%s) order by `user_id`,`role_id`",
		userRoleTable, placeholders(len(userIds)))
	args := make([]any, 0, len(userIds))
	for _, id := range userIds {
		args = append(args, id)
	}
	var rows []UserRoleId
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.UserId] = append(result[row.UserId], row.RoleId)
	}
	return result, nil
}

// SetUserRoles 重置用户角色绑定：先删后插（整存），roleIds 为空即清空绑定
func (m *defaultUserModel) SetUserRoles(ctx context.Context, session sqlx.Session, userId int64, roleIds []int64) error {
	// 空绑定仅清空，无需事务包裹
	if session == nil {
		return m.conn.TransactCtx(ctx, func(txCtx context.Context, tx sqlx.Session) error {
			return m.SetUserRoles(txCtx, tx, userId, roleIds)
		})
	}
	delQuery := fmt.Sprintf("delete from %s where `user_id` = ?", userRoleTable)
	if _, err := session.ExecCtx(ctx, delQuery, userId); err != nil {
		return err
	}
	roleIds = dedupeInt64(roleIds)
	if len(roleIds) == 0 {
		return nil
	}
	insQuery := fmt.Sprintf("insert into %s (`user_id`,`role_id`) values %s",
		userRoleTable, placeholders(len(roleIds)))
	args := make([]any, 0, len(roleIds)+1)
	args = append(args, userId)
	for _, id := range roleIds {
		args = append(args, id)
	}
	_, err := session.ExecCtx(ctx, insQuery, args...)
	return err
}

// AddUserRoles 幂等追加角色绑定（insert ignore：已存在的绑定不受影响）
func (m *defaultUserModel) AddUserRoles(ctx context.Context, userId int64, roleIds []int64) error {
	roleIds = dedupeInt64(roleIds)
	if len(roleIds) == 0 {
		return nil
	}
	query := fmt.Sprintf("insert ignore into %s (`user_id`,`role_id`) values %s",
		userRoleTable, placeholders(len(roleIds)))
	args := make([]any, 0, len(roleIds)+1)
	args = append(args, userId)
	for _, id := range roleIds {
		args = append(args, id)
	}
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// CountByRole 统计引用角色的用户数
func (m *defaultUserModel) CountByRole(ctx context.Context, roleId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `role_id` = ?", userRoleTable)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, roleId); err != nil {
		return 0, err
	}
	return total, nil
}
