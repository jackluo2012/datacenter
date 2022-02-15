package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	baseMemberFieldNames          = builder.RawFieldNames(&BaseMember{})
	baseMemberRows                = strings.Join(baseMemberFieldNames, ",")
	baseMemberRowsExpectAutoSet   = strings.Join(stringx.Remove(baseMemberFieldNames, "id", "create_time", "update_time"), ",")
	baseMemberRowsWithPlaceHolder = strings.Join(stringx.Remove(baseMemberFieldNames, "id", "create_time", "update_time"), "=?,") + "=?"

	cacheBaseMemberMobilePrefix = "cache#BaseMember#mobile#"
	cacheBaseMemberIdPrefix     = "cache#BaseMember#id#"
)

type (
	BaseMemberModel struct {
		sqlc.CachedConn
		table string
	}

	BaseMember struct {
		Mobile     string    `db:"mobile"`      // 手机号
		CreateTime time.Time `db:"create_time"` // 创建时间
		DeletedAt  time.Time `db:"deleted_at"`
		Icard      string    `db:"icard"`    // 身份证号码
		Password   string    `db:"password"` // 密码
		Realname   string    `db:"realname"` // 真实姓名
		Salt       string    `db:"salt"`     // 密码加盐
		Status     int64     `db:"status"`   // 状态
		UpdateTime time.Time `db:"update_time"`
		Username   string    `db:"username"` // 帐号
		Id         int64     `db:"id"`       // 用户id
	}
)

func NewBaseMemberModel(conn sqlx.SqlConn, c cache.CacheConf) *BaseMemberModel {
	return &BaseMemberModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "base_member",
	}
}

func (m *BaseMemberModel) Insert(data BaseMember) (sql.Result, error) {
	baseMemberMobileKey := fmt.Sprintf("%s%v", cacheBaseMemberMobilePrefix, data.Mobile)
	ret, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, baseMemberRowsExpectAutoSet)
		return conn.Exec(query, data.Mobile, data.DeletedAt, data.Icard, data.Password, data.Realname, data.Salt, data.Status, data.Username)
	}, baseMemberMobileKey)
	return ret, err
}

func (m *BaseMemberModel) FindOne(id int64) (*BaseMember, error) {
	baseMemberIdKey := fmt.Sprintf("%s%v", cacheBaseMemberIdPrefix, id)
	var resp BaseMember
	err := m.QueryRow(&resp, baseMemberIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where id = ? limit 1", baseMemberRows, m.table)
		return conn.QueryRow(v, query, id)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *BaseMemberModel) FindOneByMobile(mobile string) (*BaseMember, error) {
	baseMemberMobileKey := fmt.Sprintf("%s%v", cacheBaseMemberMobilePrefix, mobile)
	var resp BaseMember
	err := m.QueryRowIndex(&resp, baseMemberMobileKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where mobile = ? limit 1", baseMemberRows, m.table)
		if err := conn.QueryRow(&resp, query, mobile); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *BaseMemberModel) Update(data BaseMember) error {
	baseMemberIdKey := fmt.Sprintf("%s%v", cacheBaseMemberIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where id = ?", m.table, baseMemberRowsWithPlaceHolder)
		return conn.Exec(query, data.Mobile, data.DeletedAt, data.Icard, data.Password, data.Realname, data.Salt, data.Status, data.Username, data.Id)
	}, baseMemberIdKey)
	return err
}

func (m *BaseMemberModel) Delete(id int64) error {
	data, err := m.FindOne(id)
	if err != nil {
		return err
	}

	baseMemberMobileKey := fmt.Sprintf("%s%v", cacheBaseMemberMobilePrefix, data.Mobile)
	baseMemberIdKey := fmt.Sprintf("%s%v", cacheBaseMemberIdPrefix, id)
	_, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where id = ?", m.table)
		return conn.Exec(query, id)
	}, baseMemberMobileKey, baseMemberIdKey)
	return err
}

func (m *BaseMemberModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheBaseMemberIdPrefix, primary)
}

func (m *BaseMemberModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where id = ? limit 1", baseMemberRows, m.table)
	return conn.QueryRow(v, query, primary)
}
