package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const recordTable = "`dc_mdm_record`"

// recordRows 查询列（显式 7 列，与 recordRow 7 字段严格对齐，不含 created_at/updated_at）
const recordRows = "`id`,`model_id`,`biz_code`,`data_json`,`status`,`version`,`beid`"

type (
	// RecordRow 记录查询行：与 recordRows 的 7 列一一对应
	RecordRow struct {
		Id       int64  `db:"id"`
		ModelId  int64  `db:"model_id"`
		BizCode  string `db:"biz_code"`
		DataJson string `db:"data_json"`
		Status   string `db:"status"` //draft/published/archived
		Version  int32  `db:"version"`
		Beid     int64  `db:"beid"`
	}

	// RecordModel 主数据记录（dc_mdm_record）
	RecordModel interface {
		// Insert 新增记录（model_id+biz_code 唯一，并发撞键归一为 ErrDupBizCode）
		Insert(ctx context.Context, data *RecordRow) (int64, error)
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*RecordRow, error)
		// FindByBizCode 按模型内业务编码查询（model_id+biz_code 唯一），保存前查重用
		FindByBizCode(ctx context.Context, modelId, beid int64, bizCode string) (*RecordRow, error)
		// FindList 条件分页（beid 必带，modelId 可选，keyword 模糊匹配 biz_code），返回列表与总数
		FindList(ctx context.Context, beid, modelId int64, keyword string, page, size int64) ([]*RecordRow, int64, error)
		// Update 更新编码与内容并自增版本（version = version + 1），返回自增后的版本
		Update(ctx context.Context, data *RecordRow) (int32, error)
		// UpdateStatus 更新记录状态（bumpVersion=true 时同步 version+1，用于已发布记录重发布）
		UpdateStatus(ctx context.Context, id, beid int64, status string, bumpVersion bool) error
		// Delete 按主键删除（分发留痕不级联，供审计回溯）
		Delete(ctx context.Context, id, beid int64) error
		// CountByModel 统计模型下记录数（DeleteModel 删除前校验用）
		CountByModel(ctx context.Context, modelId, beid int64) (int64, error)
	}

	defaultRecordModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewRecordModel 创建主数据记录 model
func NewRecordModel(conn sqlx.SqlConn) RecordModel {
	return &defaultRecordModel{
		conn:  conn,
		table: recordTable,
	}
}

// Insert 新增主数据记录，version 固定初始 1，created_at/updated_at 由库默认值填充
func (m *defaultRecordModel) Insert(ctx context.Context, data *RecordRow) (int64, error) {
	query := fmt.Sprintf("insert into %s (`model_id`,`biz_code`,`data_json`,`status`,`version`,`beid`) values (?,?,?,?,?,?)", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, data.ModelId, data.BizCode, data.DataJson, data.Status, data.Version, data.Beid)
	if err != nil {
		//查重后并发窗口撞 uk_model_code 时归一为业务错误
		if isDuplicateEntry(err) {
			return 0, ErrDupBizCode
		}
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindOne 按主键查询记录（beid>0 时追加多租户过滤）
func (m *defaultRecordModel) FindOne(ctx context.Context, id, beid int64) (*RecordRow, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ?", recordRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and `beid` = ?"
		args = append(args, beid)
	}
	var resp RecordRow
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

// FindByBizCode 按模型内业务编码查询记录（model_id+biz_code 唯一）
func (m *defaultRecordModel) FindByBizCode(ctx context.Context, modelId, beid int64, bizCode string) (*RecordRow, error) {
	query := fmt.Sprintf("select %s from %s where `model_id` = ? and `biz_code` = ? and `beid` = ? limit 1", recordRows, m.table)
	var resp RecordRow
	err := m.conn.QueryRowCtx(ctx, &resp, query, modelId, bizCode, beid)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindList 条件分页查询记录，keyword 模糊匹配 biz_code，返回列表与总数
func (m *defaultRecordModel) FindList(ctx context.Context, beid, modelId int64, keyword string, page, size int64) ([]*RecordRow, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "`beid` = ?"
	args := []any{beid}
	if modelId > 0 {
		where += " and `model_id` = ?"
		args = append(args, modelId)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and `biz_code` like ?"
		args = append(args, "%"+escapeLike(kw)+"%")
	}

	//总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	//分页
	page, size = normalizePage(page, size)
	listQuery := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?", recordRows, m.table, where)
	args = append(args, size, (page-1)*size)
	var list []*RecordRow
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 更新记录编码与内容并自增版本号（version = version + 1），返回自增后的版本。
// 状态不变，由 UpdateStatus 单独流转
func (m *defaultRecordModel) Update(ctx context.Context, data *RecordRow) (int32, error) {
	query := fmt.Sprintf("update %s set `biz_code` = ?, `data_json` = ?, `version` = `version` + 1 "+
		"where `id` = ? and `beid` = ?", m.table)
	if _, err := m.conn.ExecCtx(ctx, query, data.BizCode, data.DataJson, data.Id, data.Beid); err != nil {
		if isDuplicateEntry(err) {
			return 0, ErrDupBizCode
		}
		return 0, err
	}
	return data.Version + 1, nil
}

// UpdateStatus 更新记录状态；bumpVersion=true 时同步 version+1（已发布记录重发布）
func (m *defaultRecordModel) UpdateStatus(ctx context.Context, id, beid int64, status string, bumpVersion bool) error {
	setVer := ""
	if bumpVersion {
		setVer = ", `version` = `version` + 1"
	}
	query := fmt.Sprintf("update %s set `status` = ?%s where `id` = ? and `beid` = ?", m.table, setVer)
	_, err := m.conn.ExecCtx(ctx, query, status, id, beid)
	return err
}

// Delete 按主键删除记录（dc_mdm_distribution 分发留痕不级联删除，供审计回溯）
func (m *defaultRecordModel) Delete(ctx context.Context, id, beid int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ? and `beid` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, beid)
	return err
}

// CountByModel 统计模型下记录总数（DeleteModel 存在记录即拒绝删除）
func (m *defaultRecordModel) CountByModel(ctx context.Context, modelId, beid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `model_id` = ? and `beid` = ?", m.table)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, modelId, beid); err != nil {
		return 0, err
	}
	return total, nil
}
