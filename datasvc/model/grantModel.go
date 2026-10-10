package model

import (
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	// GrantModel dc_api_grant 表模型（API 授权关系）
	GrantModel interface {
		// Insert 授权（幂等：uk_api_app 冲突时刷新过期时间与脱敏字段）
		Insert(data *ApiGrant) error
		// FindByApiAndApp 查询指定 API 与应用的授权
		FindByApiAndApp(apiId, appId int64) (*ApiGrant, error)
		// ListByApp 查询应用名下全部授权
		ListByApp(appId int64) ([]*ApiGrant, error)
		// DeleteByApi 删除某 API 的全部授权（API 删除时清理）
		DeleteByApi(apiId int64) error
	}

	defaultGrantModel struct {
		conn  sqlx.SqlConn
		table string
	}

	// ApiGrant dc_api_grant 表结构
	ApiGrant struct {
		Id        int64        `db:"id"`
		ApiId     int64        `db:"api_id"`
		AppId     int64        `db:"app_id"`
		ExpireAt  sql.NullTime `db:"expire_at"`  //NULL 表示永久
		FieldMask string       `db:"field_mask"` //脱敏字段，逗号分隔
		Beid      int64        `db:"beid"`
	}
)

func NewGrantModel(conn sqlx.SqlConn) GrantModel {
	return &defaultGrantModel{
		conn:  conn,
		table: "`dc_api_grant`",
	}
}

// grantSelectRows 查询列：field_mask 为可空列，用 ifnull 兜底为空串
const grantSelectRows = "`id`, `api_id`, `app_id`, `expire_at`, " +
	"ifnull(`field_mask`,'') as `field_mask`, `beid`"

// Insert 授权（幂等）：重复授权同 api+app 时不报错，刷新过期时间/脱敏字段/租户
func (m *defaultGrantModel) Insert(data *ApiGrant) error {
	query := fmt.Sprintf(
		"insert into %s (`api_id`, `app_id`, `expire_at`, `field_mask`, `beid`) values (?, ?, ?, ?, ?) "+
			"on duplicate key update `expire_at` = values(`expire_at`), "+
			"`field_mask` = values(`field_mask`), `beid` = values(`beid`)", m.table)
	_, err := m.conn.Exec(query,
		data.ApiId, data.AppId, data.ExpireAt, data.FieldMask, data.Beid)
	return err
}

// FindByApiAndApp 查询指定 API 与应用的授权
func (m *defaultGrantModel) FindByApiAndApp(apiId, appId int64) (*ApiGrant, error) {
	query := fmt.Sprintf("select %s from %s where `api_id` = ? and `app_id` = ? limit 1",
		grantSelectRows, m.table)
	var resp ApiGrant
	err := m.conn.QueryRow(&resp, query, apiId, appId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// ListByApp 查询应用名下全部授权
func (m *defaultGrantModel) ListByApp(appId int64) ([]*ApiGrant, error) {
	query := fmt.Sprintf("select %s from %s where `app_id` = ? order by `id` desc",
		grantSelectRows, m.table)
	var list []*ApiGrant
	err := m.conn.QueryRows(&list, query, appId)
	switch err {
	case nil:
		return list, nil
	case sqlc.ErrNotFound:
		return []*ApiGrant{}, nil
	default:
		return nil, err
	}
}

// DeleteByApi 删除某 API 的全部授权
func (m *defaultGrantModel) DeleteByApi(apiId int64) error {
	query := fmt.Sprintf("delete from %s where `api_id` = ?", m.table)
	_, err := m.conn.Exec(query, apiId)
	return err
}
