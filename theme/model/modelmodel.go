package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const modelTable = "`dc_theme_model`"

// modelRows 查询列（显式 10 列，与 ThemeModelRow 10 字段严格对齐；fields_json/dim_json
// 为可空 text，用 ifnull 兜底避免 NULL 扫 string 报错。注：pb 的 create_sql 非库列，
// 由 logic 层按 fields_json 实时生成，故不在查询列中）
const modelRows = "`id`,`theme_id`,`layer`,`table_name`,`model_name`," +
	"IFNULL(`fields_json`,'') AS `fields_json`,IFNULL(`dim_json`,'') AS `dim_json`," +
	"`datasource_id`,`db_name`,`beid`"

// modelInsertCols 插入列清单（不含自增 id 与 created_at/updated_at）
const modelInsertCols = "`theme_id`,`layer`,`table_name`,`model_name`,`fields_json`," +
	"`dim_json`,`datasource_id`,`db_name`,`beid`"

// modelInsertFieldSize 插入字段数（与 modelInsertCols 严格对应）
const modelInsertFieldSize = 9

type (
	// ThemeModelRow 数仓模型查询行：与 modelRows 的 10 列一一对应（不含时间列）
	ThemeModelRow struct {
		Id           int64  `db:"id"`
		ThemeId      int64  `db:"theme_id"`
		Layer        string `db:"layer"`
		TableName    string `db:"table_name"`
		ModelName    string `db:"model_name"`
		FieldsJson   string `db:"fields_json"`
		DimJson      string `db:"dim_json"`
		DatasourceId int64  `db:"datasource_id"`
		DbName       string `db:"db_name"`
		Beid         int64  `db:"beid"`
	}

	// ModelModel 数仓分层模型（dc_theme_model，uk_beid_table 唯一）
	ModelModel interface {
		// Insert 新增模型，返回自增 id（beid+table_name 唯一，并发撞键归一为 ErrDupTableName）
		Insert(ctx context.Context, data *ThemeModelRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*ThemeModelRow, error)
		// FindByTable 按模型表名查重（beid+table_name 唯一）
		FindByTable(ctx context.Context, beid int64, tableName string) (*ThemeModelRow, error)
		// Update 全量更新（不含自增 id 与时间列，带多租户过滤）
		Update(ctx context.Context, data *ThemeModelRow) error
		// Delete 按主键删除
		Delete(ctx context.Context, id, beid int64) error
		// FindList 条件分页（beid 必带，themeId>0 按主题过滤，layer 非空按层过滤）
		FindList(ctx context.Context, beid, themeId int64, layer string, page, size int64) ([]*ThemeModelRow, int64, error)
		// CountByTheme 统计主题下模型数（删除主题前校验用）
		CountByTheme(ctx context.Context, themeId int64) (int64, error)
	}

	defaultModelModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewModelModel 创建数仓模型 model
func NewModelModel(conn sqlx.SqlConn) ModelModel {
	return &defaultModelModel{
		conn:  conn,
		table: modelTable,
	}
}

// Insert 新增数仓模型，created_at/updated_at 由库默认值填充
func (m *defaultModelModel) Insert(ctx context.Context, data *ThemeModelRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (%s) values (%s)", m.table, modelInsertCols, placeholders(modelInsertFieldSize))
	ret, err := m.conn.ExecCtx(ctx, query, data.ThemeId, data.Layer, data.TableName, data.ModelName,
		data.FieldsJson, data.DimJson, data.DatasourceId, data.DbName, data.Beid)
	if err != nil {
		//查重后并发窗口撞 uk_beid_table 时归一为业务错误
		if isDuplicateEntry(err) {
			return 0, ErrDupTableName
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
func (m *defaultModelModel) FindOne(ctx context.Context, id, beid int64) (*ThemeModelRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", modelRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp ThemeModelRow
	if err := queryOne(ctx, m.conn, &resp, query+" limit 1", args...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByTable 按模型表名查询（beid+table_name 唯一），供保存前查重
func (m *defaultModelModel) FindByTable(ctx context.Context, beid int64, tableName string) (*ThemeModelRow, error) {
	query := fmt.Sprintf("select %s from %s where `beid` = ? and `table_name` = ? limit 1", modelRows, m.table)
	var resp ThemeModelRow
	if err := queryOne(ctx, m.conn, &resp, query, beid, tableName); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 全量更新模型（表名唯一性由 uk_beid_table 兜底，冲突归一为业务错误）
func (m *defaultModelModel) Update(ctx context.Context, data *ThemeModelRow) error {
	query := fmt.Sprintf("update %s set `theme_id` = ?, `layer` = ?, `table_name` = ?, `model_name` = ?, "+
		"`fields_json` = ?, `dim_json` = ?, `datasource_id` = ?, `db_name` = ? "+
		"where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.ThemeId, data.Layer, data.TableName, data.ModelName,
		data.FieldsJson, data.DimJson, data.DatasourceId, data.DbName, data.Id, data.Beid)
	if err != nil && isDuplicateEntry(err) {
		return ErrDupTableName
	}
	return err
}

// Delete 按主键删除模型
func (m *defaultModelModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// FindList 条件分页查询模型，themeId>0 按主题过滤，layer 非空按层过滤，返回列表与总数
func (m *defaultModelModel) FindList(ctx context.Context, beid, themeId int64, layer string, page, size int64) ([]*ThemeModelRow, int64, error) {
	if err := ensureBeid(beid); err != nil {
		return nil, 0, err
	}
	where := "`beid` = ?"
	args := []any{beid}
	if themeId > 0 {
		where += " and `theme_id` = ?"
		args = append(args, themeId)
	}
	if lw := strings.TrimSpace(layer); lw != "" {
		where += " and `layer` = ?"
		args = append(args, lw)
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
	var list []*ThemeModelRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountByTheme 统计指定主题下的模型数
func (m *defaultModelModel) CountByTheme(ctx context.Context, themeId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `theme_id` = ?", m.table)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, themeId); err != nil {
		return 0, err
	}
	return total, nil
}
