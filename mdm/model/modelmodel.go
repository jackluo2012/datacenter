package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const modelTable = "`dc_mdm_model`"

// modelRows 查询列（显式 7 列，与 modelRow 7 字段严格对齐；fields_json 为可空 text，
// 用 ifnull 兜底避免 NULL 扫 string 报错。列数与扫描字段数不一致时 go-zero 报
// not matching destination to scan——少列扫宽 struct 同样报错，故专列窄结构体）
const modelRows = "`id`,`model_code`,`model_name`,`table_name`,IFNULL(`fields_json`,'') AS `fields_json`,`code_strategy`,`beid`"

// modelInsertFieldSize 插入字段数（不含自增 id 与 created_at/updated_at，与 modelInsertCols 对应）
const modelInsertFieldSize = 6

// modelInsertCols 插入列清单（显式写出，不用 builder.RawFieldNames，避免依赖宽结构体）
const modelInsertCols = "`model_code`,`model_name`,`table_name`,`fields_json`,`code_strategy`,`beid`"

type (
	// MdmModelRow 模型查询行：与 modelRows 的 7 列一一对应（不含时间列，pb 亦无需时间）
	MdmModelRow struct {
		Id           int64  `db:"id"`
		ModelCode    string `db:"model_code"`
		ModelName    string `db:"model_name"`
		TableName    string `db:"table_name"`
		FieldsJson   string `db:"fields_json"`
		CodeStrategy string `db:"code_strategy"`
		Beid         int64  `db:"beid"`
	}

	// ModelModel 主数据模型（dc_mdm_model）
	ModelModel interface {
		// Insert 新增模型，返回自增 id（beid+model_code 唯一，并发撞键归一为 ErrDupModelCode）
		Insert(ctx context.Context, data *MdmModelRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*MdmModelRow, error)
		// FindByCode 按模型编码查重（beid+model_code 唯一）
		FindByCode(ctx context.Context, beid int64, modelCode string) (*MdmModelRow, error)
		// Update 全量更新（不含自增 id 与时间列，带多租户过滤）
		Update(ctx context.Context, data *MdmModelRow) error
		// Delete 按主键删除（是否可删由业务侧先查记录数决定）
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页（beid 必带，keyword 模糊匹配 model_code/model_name），返回列表与总数
		FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*MdmModelRow, int64, error)
	}

	defaultModelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewModelModel 创建主数据模型 model
func NewModelModel(conn sqlx.SqlConn) ModelModel {
	return &defaultModelModel{
		conn:  conn,
		table: modelTable,
	}
}

// Insert 新增主数据模型，created_at/updated_at 由库默认值填充
func (m *defaultModelModel) Insert(ctx context.Context, data *MdmModelRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, modelInsertCols, placeholders(modelInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.ModelCode, data.ModelName, data.TableName,
		data.FieldsJson, data.CodeStrategy, data.Beid)
	if err != nil {
		//查重后并发窗口撞 uk_beid_code 时归一为业务错误
		if isDuplicateEntry(err) {
			return 0, ErrDupModelCode
		}
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询模型（beid>0 时追加多租户过滤）
func (m *defaultModelModel) FindOne(ctx context.Context, id, beid int64) (*MdmModelRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", modelRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp MdmModelRow
	err := m.conn.QueryRowCtx(ctx, &resp, query+" limit 1", args...)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCode 按模型编码查询（beid+model_code 唯一），供保存前查重
func (m *defaultModelModel) FindByCode(ctx context.Context, beid int64, modelCode string) (*MdmModelRow, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `model_code` = ? limit 1", modelRows, m.table)
	var resp MdmModelRow
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, modelCode)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 全量更新模型（编码唯一性由 uk_beid_code 兜底，冲突归一为业务错误）
func (m *defaultModelModel) Update(ctx context.Context, data *MdmModelRow) error {
	query := fmt.Sprintf("update %s set `model_code` = ?, `model_name` = ?, `table_name` = ?, "+
		"`fields_json` = ?, `code_strategy` = ? where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.ModelCode, data.ModelName, data.TableName,
		data.FieldsJson, data.CodeStrategy, data.Id, data.Beid)
	if err != nil && isDuplicateEntry(err) {
		return ErrDupModelCode
	}
	return err
}

// Delete 按主键删除模型
func (m *defaultModelModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 条件分页查询模型，keyword 模糊匹配 model_code/model_name，返回列表与总数
func (m *defaultModelModel) FindList(ctx context.Context, beid int64, keyword string, page, size int64) ([]*MdmModelRow, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (`model_code` like ? or `model_name` like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}

	//总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	//分页
	page, size = normalizePage(page, size)
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", modelRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*MdmModelRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
