package model

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	// AppModel dc_api_app 表模型（调用方应用）
	AppModel interface {
		// Insert 注册应用，返回自增 id
		Insert(data *ApiApp) (int64, error)
		// FindOne 按主键查询
		FindOne(id int64) (*ApiApp, error)
		// FindByKey 按 app_key 查询（执行链鉴权入口）
		FindByKey(appKey string) (*ApiApp, error)
	}

	defaultAppModel struct {
		conn  sqlx.SqlConn
		table string
	}

	// ApiApp dc_api_app 表结构
	ApiApp struct {
		Id           int64  `db:"id"`
		AppKey       string `db:"app_key"`
		AppSecretEnc []byte `db:"app_secret_enc"` //AES-GCM 加密后的密钥
		Name         string `db:"name"`
		RateLimitQps int64  `db:"rate_limit_qps"`
		IpWhitelist  string `db:"ip_whitelist"` //逗号分隔，空为不限
		Beid         int64  `db:"beid"`
	}
)

func NewAppModel(conn sqlx.SqlConn) AppModel {
	return &defaultAppModel{
		conn:  conn,
		table: "`dc_api_app`",
	}
}

// Insert 注册应用，返回自增 id
func (m *defaultAppModel) Insert(data *ApiApp) (int64, error) {
	query := fmt.Sprintf(
		"insert into %s (`app_key`, `app_secret_enc`, `name`, `rate_limit_qps`, `ip_whitelist`, `beid`) "+
			"values (?, ?, ?, ?, ?, ?)", m.table)
	ret, err := m.conn.Exec(query,
		data.AppKey, data.AppSecretEnc, data.Name, data.RateLimitQps,
		data.IpWhitelist, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询
func (m *defaultAppModel) FindOne(id int64) (*ApiApp, error) {
	query := fmt.Sprintf(
		"select `id`, `app_key`, `app_secret_enc`, `name`, `rate_limit_qps`, `ip_whitelist`, `beid` "+
			"from %s where `id` = ? limit 1", m.table)
	var resp ApiApp
	err := m.conn.QueryRow(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByKey 按 app_key 查询（uk_app_key 唯一键）
func (m *defaultAppModel) FindByKey(appKey string) (*ApiApp, error) {
	query := fmt.Sprintf(
		"select `id`, `app_key`, `app_secret_enc`, `name`, `rate_limit_qps`, `ip_whitelist`, `beid` "+
			"from %s where `app_key` = ? limit 1", m.table)
	var resp ApiApp
	err := m.conn.QueryRow(&resp, query, appKey)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
