package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"
)

var (
	appConfigFieldNames          = builderx.FieldNames(&AppConfig{})
	appConfigRows                = strings.Join(appConfigFieldNames, ",")
	appConfigRowsExpectAutoSet   = strings.Join(stringx.Remove(appConfigFieldNames, "id", "create_time", "update_time"), ",")
	appConfigRowsWithPlaceHolder = strings.Join(stringx.Remove(appConfigFieldNames, "id", "create_time", "update_time"), "=?,") + "=?"

	cacheAppConfigIdPrefix      = "cache#AppConfig#id#"
	CacheAppConfigIdPtyidPrefix = "cache#AppConfig#id#ptyid#"
)

type (
	AppConfigModel interface {
		Insert(data AppConfig) (sql.Result, error)
		FindOne(id int64) (*AppConfig, error)
		Update(data AppConfig) error
		Delete(id int64) error
		FindOneByAppid(id, ptyid int64) (*AppConfig, error)
	}

	defaultAppConfigModel struct {
		sqlc.CachedConn
		table string
	}

	AppConfig struct {
		Appid     string    `db:"appid"`     // appid
		Appsecret string    `db:"appsecret"` // 配置密钥
		Beid      int64     `db:"beid"`      // 对应的平台
		CreateBy  string    `db:"create_by"`
		CreatedAt time.Time `db:"created_at"`
		Ptyid     int64     `db:"ptyid"` // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Title     string    `db:"title"` // 社交描述
		UpdateBy  string    `db:"update_by"`
		UpdatedAt time.Time `db:"updated_at"`
		Id        int64     `db:"id"`
	}
)

func NewAppConfigModel(conn sqlx.SqlConn, c cache.CacheConf) AppConfigModel {
	return &defaultAppConfigModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "app_config",
	}
}

func (m *defaultAppConfigModel) Insert(data AppConfig) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?,  ?, ?, ?)", m.table, appConfigRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Appid, data.Appsecret, data.Beid, data.CreateBy, data.CreatedAt, data.Ptyid, data.Title, data.UpdateBy, data.UpdatedAt)

	return ret, err
}

func (m *defaultAppConfigModel) FindOne(id int64) (*AppConfig, error) {
	appConfigIdKey := fmt.Sprintf("%s%v", cacheAppConfigIdPrefix, id)
	var resp AppConfig
	err := m.QueryRow(&resp, appConfigIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where id = ? limit 1", appConfigRows, m.table)
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

func (m defaultAppConfigModel) FindOneByAppid(id, ptyid int64) (*AppConfig, error) {
	appConfigIdKey := fmt.Sprintf("%s%v%v", CacheAppConfigIdPtyidPrefix, id, ptyid)
	var resp AppConfig
	err := m.QueryRow(&resp, appConfigIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where beid = ? AND ptyid=? limit 1", appConfigRows, m.table)
		return conn.QueryRow(v, query, id, ptyid)
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

func GetCacheAppConfigIdPtyidPrefix(id, ptyid int64) string {
	return fmt.Sprintf("%s%v%v", CacheAppConfigIdPtyidPrefix, id, ptyid)
}

func (m *defaultAppConfigModel) Update(data AppConfig) error {
	appConfigIdKey := fmt.Sprintf("%s%v", cacheAppConfigIdPrefix, data.Id)
	appConfigIdPtyidKey := fmt.Sprintf("%s%v%v", CacheAppConfigIdPtyidPrefix, data.Beid, data.Ptyid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where id = ?", m.table, appConfigRowsWithPlaceHolder)
		return conn.Exec(query, data.Appid, data.Appsecret, data.Beid, data.CreateBy, data.CreatedAt, data.Ptyid, data.Title, data.UpdateBy, data.UpdatedAt, data.Id)
	}, appConfigIdKey, appConfigIdPtyidKey)
	return err
}

func (m *defaultAppConfigModel) Delete(id int64) error {

	appConfigIdKey := fmt.Sprintf("%s%v", cacheAppConfigIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where id = ?", m.table)
		return conn.Exec(query, id)
	}, appConfigIdKey)
	return err
}

func (m *defaultAppConfigModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppConfigIdPrefix, primary)
}

func (m *defaultAppConfigModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where id = ? limit 1", appConfigRows, m.table)
	return conn.QueryRow(v, query, primary)
}
