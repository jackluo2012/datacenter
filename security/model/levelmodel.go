package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	levelTable = "`dc_sec_level`"
	// levelInsertColumns 插入字段（不含自增 id，uk_beid_code 冲突时走更新分支）
	levelInsertColumns = "`code`,`name`,`comment`,`beid`"
	// levelRows 查询列（5 列与 SecLevel 5 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
	levelRows = "`id`,`code`,`name`,`comment`,`beid`"
)

// 分级目录（dc_sec_level）：内置 L1 公开/L2 内部/L3 敏感/L4 机密，L5-L9 留给租户自定义
type (
	// SecLevel 分级行
	SecLevel struct {
		Id      int64  `db:"id"`
		Code    string `db:"code"` //L1-L9
		Name    string `db:"name"` //公开/内部/敏感/机密
		Comment string `db:"comment"`
		Beid    int64  `db:"beid"`
	}

	// LevelModel 分级目录表操作接口
	LevelModel interface {
		// Save 按 uk_beid_code upsert（code 为自然键：不存在则新建，已存在则覆盖 name/comment），返回记录 id
		Save(ctx context.Context, data *SecLevel) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*SecLevel, error)
		// FindByCode 按 beid + code 查询（打标/授权时密级存在性校验用）
		FindByCode(ctx context.Context, beid int64, code string) (*SecLevel, error)
		// List 全量分级目录（按 code 升序）
		List(ctx context.Context, beid int64) ([]*SecLevel, error)
		// SeedDefaults 播种内置四级目录（幂等：insert ignore，仅缺行时补插）
		SeedDefaults(ctx context.Context, beid int64) error
	}

	defaultLevelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewLevelModel 创建分级目录 model
func NewLevelModel(conn sqlx.SqlConn) LevelModel {
	return &defaultLevelModel{
		conn:  conn,
		table: levelTable,
	}
}

// Save 按 uk_beid_code upsert 分级。
// 更新分支借 `id = last_insert_id(id)` 让 LastInsertId 返回既有主键（insert/update 两分支都可靠）
func (m *defaultLevelModel) Save(ctx context.Context, data *SecLevel) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?) "+
		"on duplicate key update `id` = last_insert_id(`id`), `name` = values(`name`), `comment` = values(`comment`)",
		m.table, levelInsertColumns)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Code, data.Name, data.Comment, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询分级
func (m *defaultLevelModel) FindOne(ctx context.Context, id, beid int64) (*SecLevel, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", levelRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp SecLevel
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCode 按 beid + 分级编码查询
func (m *defaultLevelModel) FindByCode(ctx context.Context, beid int64, code string) (*SecLevel, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `code` = ? limit 1", levelRows, m.table)
	var resp SecLevel
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, code)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// List 全量分级目录（按 code 升序，L1-L9 天然有序）
func (m *defaultLevelModel) List(ctx context.Context, beid int64) ([]*SecLevel, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where `beid` = ? order by `code` asc", levelRows, m.table)
	var list []*SecLevel
	if err := m.conn.QueryRowsCtx(ctx, &list, query, beid); err != nil {
		return nil, err
	}
	return list, nil
}

// SeedDefaults 播种内置四级目录（L1 公开/L2 内部/L3 敏感/L4 机密）。
// insert ignore 保证幂等：目录已存在的行不受影响，仅补插缺失行
func (m *defaultLevelModel) SeedDefaults(ctx context.Context, beid int64) error {
	query := fmt.Sprintf("insert ignore into %s (%s) values "+
		"('L1', '公开', '系统内置：公开数据', ?), "+
		"('L2', '内部', '系统内置：内部数据', ?), "+
		"('L3', '敏感', '系统内置：敏感数据', ?), "+
		"('L4', '机密', '系统内置：机密数据', ?)",
		m.table, levelInsertColumns)
	_, err := m.conn.ExecCtx(ctx, query, beid, beid, beid, beid)
	return err
}
