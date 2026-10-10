package model

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	apiFieldNames        = builder.RawFieldNames(&Api{})
	apiRowsExpectAutoSet = strings.Join(stringx.Remove(apiFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
)

// apiSelectRows 查询列：sql_text/params_json/resp_sample_json/owner 为可空文本列，
// 用 ifnull 兜底为空串，避免 NULL 扫描报错
const apiSelectRows = "`id`,`name`,`path`,`mode`,`datasource_id`," +
	"ifnull(`sql_text`,'') as `sql_text`,ifnull(`params_json`,'') as `params_json`," +
	"`cache_ttl`,`status`,ifnull(`resp_sample_json`,'') as `resp_sample_json`," +
	"ifnull(`owner`,'') as `owner`,`beid`"

type (
	// ApiModel dc_api 表模型（配置出 API 的核心元数据）
	ApiModel interface {
		// Insert 新建 API，返回自增 id
		Insert(data *Api) (int64, error)
		// FindOne 按主键查询
		FindOne(id int64) (*Api, error)
		// FindByPath 按 beid+path 查询（执行入口查找 / 唯一性校验）
		FindByPath(beid int64, path string) (*Api, error)
		// List 按 beid（可选 status）分页查询，返回列表与总数
		List(beid int64, status string, page, size int64) ([]*Api, int64, error)
		// Update 更新可编辑字段（name/sql_text/params_json/cache_ttl/resp_sample_json）
		Update(data *Api) error
		// UpdateStatus 状态流转：draft/published/offline
		UpdateStatus(id int64, status string) error
		// Delete 删除 API
		Delete(id int64) error
	}

	defaultApiModel struct {
		conn  sqlx.SqlConn
		table string
	}

	// Api dc_api 表结构
	Api struct {
		Id             int64  `db:"id"`
		Name           string `db:"name"`
		Path           string `db:"path"` //执行路径标识
		Mode           string `db:"mode"` //wizard/sql/register
		DatasourceId   int64  `db:"datasource_id"`
		SqlText        string `db:"sql_text"` //SQL 模板，:name 占位
		ParamsJson     string `db:"params_json"`
		CacheTtl       int64  `db:"cache_ttl"` //秒，0 不缓存
		Status         string `db:"status"`    //draft/published/offline
		RespSampleJson string `db:"resp_sample_json"`
		Owner          string `db:"owner"`
		Beid           int64  `db:"beid"`
	}
)

func NewApiModel(conn sqlx.SqlConn) ApiModel {
	return &defaultApiModel{
		conn:  conn,
		table: "`dc_api`",
	}
}

// Insert 新建 API，返回自增 id
func (m *defaultApiModel) Insert(data *Api) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		m.table, apiRowsExpectAutoSet)
	ret, err := m.conn.Exec(query,
		data.Name, data.Path, data.Mode, data.DatasourceId, data.SqlText,
		data.ParamsJson, data.CacheTtl, data.Status, data.RespSampleJson,
		data.Owner, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

// FindOne 按主键查询
func (m *defaultApiModel) FindOne(id int64) (*Api, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", apiSelectRows, m.table)
	var resp Api
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

// FindByPath 按 beid+path 查询（uk_beid_path 唯一键）
func (m *defaultApiModel) FindByPath(beid int64, path string) (*Api, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `path` = ? limit 1",
		apiSelectRows, m.table)
	var resp Api
	err := m.conn.QueryRow(&resp, query, beid, path)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// List 按 beid（可选 status）分页查询，返回列表与总数
func (m *defaultApiModel) List(beid int64, status string, page, size int64) ([]*Api, int64, error) {
	//分页参数兜底
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	where := "`beid` = ?"
	args := []interface{}{beid}
	if strings.TrimSpace(status) != "" {
		where += " and `status` = ?"
		args = append(args, status)
	}
	var total int64
	err := m.conn.QueryRow(&total,
		fmt.Sprintf("select count(*) from %s where %s", m.table, where), args...)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*Api{}, 0, nil
	}
	query := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?",
		apiSelectRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*Api
	err = m.conn.QueryRows(&list, query, args...)
	switch err {
	case nil:
		return list, total, nil
	case sqlc.ErrNotFound:
		return []*Api{}, total, nil
	default:
		return nil, 0, err
	}
}

// Update 更新可编辑字段（name/sql_text/params_json/cache_ttl/resp_sample_json），带 beid 租户条件
func (m *defaultApiModel) Update(data *Api) error {
	query := fmt.Sprintf(
		"update %s set `name` = ?, `sql_text` = ?, `params_json` = ?, `cache_ttl` = ?, `resp_sample_json` = ? "+
			"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.Exec(query,
		data.Name, data.SqlText, data.ParamsJson, data.CacheTtl, data.RespSampleJson,
		data.Id, data.Beid)
	return err
}

// UpdateStatus 状态流转
func (m *defaultApiModel) UpdateStatus(id int64, status string) error {
	query := fmt.Sprintf("update %s set `status` = ? where `id` = ?", m.table)
	_, err := m.conn.Exec(query, status, id)
	return err
}

// Delete 删除 API
func (m *defaultApiModel) Delete(id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}
