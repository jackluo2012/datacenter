package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	dictItemTable           = "`dc_std_dict_item`"
	dictItemInsertFieldSize = 5   // 插入字段数（与 dictItemInsertColumns 对应）
	dictItemInsertChunk     = 500 // 批量插入分片大小
	// dictItemInsertColumns 插入字段（不含自增 id），与 BatchInsert 传参顺序一一对应
	dictItemInsertColumns = "`dict_id`,`item_code`,`item_value`,`sort`,`beid`"
	// dictItemRows 查询列（6 列与 DictItem 6 字段一一对应，
	// 规避 go-zero 少列查询扫宽 struct 报 not matching destination to scan）
	dictItemRows = "`id`,`dict_id`,`item_code`,`item_value`,`sort`,`beid`"
)

// 标准字典项（dc_std_dict_item）
type (
	// DictItem 字典项行
	DictItem struct {
		Id        int64  `db:"id"`
		DictId    int64  `db:"dict_id"`
		ItemCode  string `db:"item_code"`
		ItemValue string `db:"item_value"`
		Sort      int64  `db:"sort"`
		Beid      int64  `db:"beid"`
	}

	// DictItemModel 字典项表操作接口
	// 写方法带 session 参数：随 SaveDict/DeleteDict 在事务内整存整取/级联删
	DictItemModel interface {
		// BatchInsert 批量写入字典项（session 非空时在事务内执行）
		BatchInsert(ctx context.Context, session sqlx.Session, items []*DictItem) error
		// DeleteByDictId 删除指定字典的全部字典项（session 非空时在事务内执行）
		DeleteByDictId(ctx context.Context, session sqlx.Session, dictId int64) error
		// FindByDictId 查询指定字典的全部字典项（按 sort,id 升序）
		FindByDictId(ctx context.Context, dictId int64) ([]*DictItem, error)
	}

	defaultDictItemModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewDictItemModel 创建字典项 model
func NewDictItemModel(conn sqlx.SqlConn) DictItemModel {
	return &defaultDictItemModel{
		conn:  conn,
		table: dictItemTable,
	}
}

// BatchInsert 批量写入字典项：分片拼接多行 values
func (m *defaultDictItemModel) BatchInsert(ctx context.Context, session sqlx.Session, items []*DictItem) error {
	if len(items) == 0 {
		return nil
	}
	prefix := fmt.Sprintf("insert into %s (%s) values ", m.table, dictItemInsertColumns)
	for start := 0; start < len(items); start += dictItemInsertChunk {
		end := start + dictItemInsertChunk
		if end > len(items) {
			end = len(items)
		}
		batch := items[start:end]
		query := prefix + batchValues(dictItemInsertFieldSize, len(batch))
		args := make([]any, 0, len(batch)*dictItemInsertFieldSize)
		for _, item := range batch {
			args = append(args, item.DictId, item.ItemCode, item.ItemValue, item.Sort, item.Beid)
		}
		if _, err := execCtx(ctx, m.conn, session, query, args...); err != nil {
			return fmt.Errorf("批量写入字典项失败: %w", err)
		}
	}
	return nil
}

// DeleteByDictId 删除指定字典的全部字典项（全删全插的"删"半边）
func (m *defaultDictItemModel) DeleteByDictId(ctx context.Context, session sqlx.Session, dictId int64) error {
	query := fmt.Sprintf("delete from %s where `dict_id` = ?", m.table)
	_, err := execCtx(ctx, m.conn, session, query, dictId)
	return err
}

// FindByDictId 查询指定字典的全部字典项
func (m *defaultDictItemModel) FindByDictId(ctx context.Context, dictId int64) ([]*DictItem, error) {
	query := fmt.Sprintf("select %s from %s where `dict_id` = ? order by `sort` asc, `id` asc",
		dictItemRows, m.table)
	var list []*DictItem
	err := m.conn.QueryRowsCtx(ctx, &list, query, dictId)
	switch {
	case err == nil:
		return list, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return []*DictItem{}, nil
	default:
		return nil, err
	}
}
