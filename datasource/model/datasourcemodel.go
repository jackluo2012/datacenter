package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	datasourceFieldNames          = builder.RawFieldNames(&DataSource{})
	datasourceRows                = strings.Join(datasourceFieldNames, ",")
	datasourceRowsExpectAutoSet   = strings.Join(stringx.Remove(datasourceFieldNames, "`id`", "`created_at`", "`updated_at`", "`deleted_at`"), ",")
	datasourceRowsWithPlaceHolder = strings.Join(stringx.Remove(datasourceFieldNames, "`id`", "`created_at`", "`updated_at`", "`deleted_at`"), "=?,") + "=?"
)

type (
	// DataSourceModel 数据源登记表(dc_datasource)操作接口
	DataSourceModel interface {
		// Insert 新增数据源(调用前需自行做 name+beid 查重)
		Insert(data DataSource) (sql.Result, error)
		// Update 更新基础信息(不含密码与状态, 密码走 UpdatePassword, 状态走 UpdateStatus)
		Update(data DataSource) error
		// UpdatePassword 仅更新密码密文
		UpdatePassword(id, beid int64, enc []byte) error
		// UpdateStatus 更新启停状态(1.启用 0.停用)
		UpdateStatus(id, beid int64, status int32) error
		// Delete 软删(写 deleted_at, 物理数据保留)
		Delete(id, beid int64) error
		// FindOne 按主键+平台查单个未删除数据源
		FindOne(id, beid int64) (*DataSource, error)
		// FindOneByName 按名称+平台查重(未删除范围内)
		FindOneByName(name string, beid int64) (*DataSource, error)
		// FindList 分页列表(beid 必带, type 可选), 返回列表与总数
		FindList(beid int64, typ string, page, size int64) ([]DataSource, int64, error)
	}

	defaultDataSourceModel struct {
		conn  sqlx.SqlConn
		table string
	}

	// DataSource dc_datasource 表模型
	DataSource struct {
		Id          int64          `db:"id"`           //数据源id
		Name        string         `db:"name"`         //名称
		Type        string         `db:"type"`         //类型: mysql/redis/es/http/kafka
		Host        string         `db:"host"`         //主机
		Port        int32          `db:"port"`         //端口
		DbName      string         `db:"db_name"`      //库名
		Username    string         `db:"username"`     //用户名
		PasswordEnc []byte         `db:"password_enc"` //AES-GCM 加密后的密码(nonce+密文)
		ParamsJson  sql.NullString `db:"params_json"`  //扩展参数(json, 可空)
		Status      int32          `db:"status"`       //1.启用 0.停用
		Owner       string         `db:"owner"`        //负责人
		Beid        int64          `db:"beid"`         //平台id
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
		DeletedAt   sql.NullTime   `db:"deleted_at"`
	}
)

// NewDataSourceModel 构造数据源 model
func NewDataSourceModel(conn sqlx.SqlConn) DataSourceModel {
	return &defaultDataSourceModel{
		conn:  conn,
		table: "dc_datasource",
	}
}

// Insert 新增, password_enc 列 NOT NULL, 空密码落空字节而非 NULL
func (m *defaultDataSourceModel) Insert(data DataSource) (sql.Result, error) {
	if data.PasswordEnc == nil {
		data.PasswordEnc = []byte{}
	}
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, datasourceRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.Name, data.Type, data.Host, data.Port, data.DbName,
		data.Username, data.PasswordEnc, data.ParamsJson, data.Status, data.Owner, data.Beid)
	return ret, err
}

// Update 更新基础信息(软删行不可更新)
func (m *defaultDataSourceModel) Update(data DataSource) error {
	query := fmt.Sprintf("update %s set name = ?, type = ?, host = ?, port = ?, db_name = ?, username = ?, params_json = ?, owner = ? where id = ? and beid = ? and deleted_at is null", m.table)
	_, err := m.conn.Exec(query, data.Name, data.Type, data.Host, data.Port, data.DbName,
		data.Username, data.ParamsJson, data.Owner, data.Id, data.Beid)
	return err
}

// UpdatePassword 仅更新密码密文
func (m *defaultDataSourceModel) UpdatePassword(id, beid int64, enc []byte) error {
	if enc == nil {
		enc = []byte{}
	}
	query := fmt.Sprintf("update %s set password_enc = ? where id = ? and beid = ? and deleted_at is null", m.table)
	_, err := m.conn.Exec(query, enc, id, beid)
	return err
}

// UpdateStatus 更新启停状态
func (m *defaultDataSourceModel) UpdateStatus(id, beid int64, status int32) error {
	query := fmt.Sprintf("update %s set status = ? where id = ? and beid = ? and deleted_at is null", m.table)
	_, err := m.conn.Exec(query, status, id, beid)
	return err
}

// Delete 软删: 仅写 deleted_at
func (m *defaultDataSourceModel) Delete(id, beid int64) error {
	query := fmt.Sprintf("update %s set deleted_at = now(3) where id = ? and beid = ? and deleted_at is null", m.table)
	_, err := m.conn.Exec(query, id, beid)
	return err
}

// FindOne 按主键+平台查询(过滤软删)
func (m *defaultDataSourceModel) FindOne(id, beid int64) (*DataSource, error) {
	query := fmt.Sprintf("select %s from %s where id = ? and beid = ? and deleted_at is null limit 1", datasourceRows, m.table)
	var resp DataSource
	err := m.conn.QueryRow(&resp, query, id, beid)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindOneByName 按名称+平台查重(过滤软删, 已删除的数据源名称可复用)
func (m *defaultDataSourceModel) FindOneByName(name string, beid int64) (*DataSource, error) {
	query := fmt.Sprintf("select %s from %s where name = ? and beid = ? and deleted_at is null limit 1", datasourceRows, m.table)
	var resp DataSource
	err := m.conn.QueryRow(&resp, query, name, beid)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindList 分页列表: beid 必带, typ 为空不过滤, 返回列表与 total
func (m *defaultDataSourceModel) FindList(beid int64, typ string, page, size int64) ([]DataSource, int64, error) {
	// 分页参数兜底
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	where := "where beid = ? and deleted_at is null"
	args := []interface{}{beid}
	if typ != "" {
		where += " and type = ?"
		args = append(args, typ)
	}
	// 先查总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) as total from %s %s", m.table, where)
	if err := m.conn.QueryRowPartial(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	// 再查当前页
	query := fmt.Sprintf("select %s from %s %s order by id desc limit ?, ?", datasourceRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []DataSource
	if err := m.conn.QueryRows(&list, query, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
