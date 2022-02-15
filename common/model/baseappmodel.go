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
	baseAppFieldNames          = builder.RawFieldNames(&BaseApp{})
	baseAppRows                = strings.Join(baseAppFieldNames, ",")
	baseAppRowsExpectAutoSet   = strings.Join(stringx.Remove(baseAppFieldNames, "id", "create_time", "update_time"), ",")
	baseAppRowsWithPlaceHolder = strings.Join(stringx.Remove(baseAppFieldNames, "id", "create_time", "update_time"), "=?,") + "=?"

	cacheBaseAppIdPrefix = "cache#BaseApp#id#"
)

type (
	BaseAppModel interface {
		Insert(data BaseApp) (sql.Result, error)
		FindOne(id int64) (*BaseApp, error)
		Update(data BaseApp) error
		Delete(id int64) error
	}

	defaultBaseAppModel struct {
		sqlc.CachedConn
		table string
	}

	BaseApp struct {
		CreateBy    string    `db:"create_by"`
		CreatedAt   time.Time `db:"created_at"`
		Fullwebsite string    `db:"fullwebsite"` // 完整的域名
		Isclose     int64     `db:"isclose"`     // 站点是否关闭
		Logo        string    `db:"logo"`        // 应用login
		Sname       string    `db:"sname"`       // 应用名称
		UpdateBy    string    `db:"update_by"`
		UpdatedAt   time.Time `db:"updated_at"`
		Website     string    `db:"website"` // 站点名称
		Id          int64     `db:"id"`
	}
)

func NewBaseAppModel(conn sqlx.SqlConn, c cache.CacheConf) BaseAppModel {
	return &defaultBaseAppModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "base_app",
	}
}

func (m *defaultBaseAppModel) Insert(data BaseApp) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, baseAppRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.CreateBy, data.CreatedAt, data.Fullwebsite, data.Isclose, data.Logo, data.Sname, data.UpdateBy, data.UpdatedAt, data.Website)

	return ret, err
}

func GetcacheBaseAppIdPrefix(id int64) string {
	return fmt.Sprintf("%s%v", cacheBaseAppIdPrefix, id)
}

func (m *defaultBaseAppModel) FindOne(id int64) (*BaseApp, error) {
	baseAppIdKey := fmt.Sprintf("%s%v", cacheBaseAppIdPrefix, id)
	var resp BaseApp
	err := m.QueryRow(&resp, baseAppIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where id = ? limit 1", baseAppRows, m.table)
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

func (m *defaultBaseAppModel) Update(data BaseApp) error {
	baseAppIdKey := fmt.Sprintf("%s%v", cacheBaseAppIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where id = ?", m.table, baseAppRowsWithPlaceHolder)
		return conn.Exec(query, data.CreateBy, data.CreatedAt, data.Fullwebsite, data.Isclose, data.Logo, data.Sname, data.UpdateBy, data.UpdatedAt, data.Website, data.Id)
	}, baseAppIdKey)
	return err
}

func (m *defaultBaseAppModel) Delete(id int64) error {

	baseAppIdKey := fmt.Sprintf("%s%v", cacheBaseAppIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where id = ?", m.table)
		return conn.Exec(query, id)
	}, baseAppIdKey)
	return err
}

func (m *defaultBaseAppModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheBaseAppIdPrefix, primary)
}

func (m *defaultBaseAppModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where id = ? limit 1", baseAppRows, m.table)
	return conn.QueryRow(v, query, primary)
}
