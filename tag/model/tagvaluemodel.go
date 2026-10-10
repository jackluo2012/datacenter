package model

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	tagValueTable       = "`dc_tag_value`"
	tagValueInsertChunk = 500 // ReplaceBatch 批量插入分片大小
	tagValueFieldSize   = 4   // 每行插入字段数（tag_id/object_pk/tag_value/beid）
)

type (
	// TagValue dc_tag_value 行（打标结果；ReplaceBatch 的内存行载体）
	TagValue struct {
		Id       int64     `db:"id"`
		TagId    int64     `db:"tag_id"`
		ObjectPk string    `db:"object_pk"` //对象主键值
		TagValue string    `db:"tag_value"`
		Beid     int64     `db:"beid"`
		TaggedAt time.Time `db:"tagged_at"` //打标时间（落库时由默认值填充）
	}

	// TagValueRow ListByObjectPk 专用扫描结构：与 select 的 4 列严格对齐
	// （go-zero 对少列查询扫宽 struct 会报 not matching destination to scan）
	TagValueRow struct {
		TagKey   string    `db:"tag_key"`
		TagName  string    `db:"tag_name"`
		TagValue string    `db:"tag_value"`
		TaggedAt time.Time `db:"tagged_at"`
	}

	TagValueModel interface {
		// ReplaceBatch 事务内先按 tag_id 删除旧值，再分片批量写入新值（整体替换语义）
		ReplaceBatch(ctx context.Context, tagId, beid int64, rows []*TagValue) error
		// ListByObjectPk 按对象+主键值查标签画像，join dc_tag_def 出 tag_key/tag_name
		ListByObjectPk(ctx context.Context, objectId, beid int64, objectPk string) ([]*TagValueRow, error)
	}

	defaultTagValueModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewTagValueModel 创建标签值 model
func NewTagValueModel(conn sqlx.SqlConn) TagValueModel {
	return &defaultTagValueModel{
		conn:  conn,
		table: tagValueTable,
	}
}

// ReplaceBatch 整体替换某标签的打标值：同一事务内先 delete by tag_id，
// 再按 tagValueInsertChunk 分片批量 insert（tagged_at 由库默认值填充为当前时间）。
// rows 为空时仅清空旧值（空结果也是一次有效打标）。
// 调用方需保证 rows 内 object_pk 已去重，否则会撞 uk_tag_pk 唯一键。
func (m *defaultTagValueModel) ReplaceBatch(ctx context.Context, tagId, beid int64, rows []*TagValue) error {
	if tagId <= 0 {
		return ErrBeidRequired
	}
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先清空该标签旧值（整体替换语义）
		delQuery := "delete from `dc_tag_value` where `tag_id` = ?"
		if _, err := session.ExecCtx(ctx, delQuery, tagId); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		// 分片拼接多行 values 批量写入
		prefix := "insert into " + m.table + " (`tag_id`,`object_pk`,`tag_value`,`beid`) values "
		for start := 0; start < len(rows); start += tagValueInsertChunk {
			end := start + tagValueInsertChunk
			if end > len(rows) {
				end = len(rows)
			}
			batch := rows[start:end]
			query := prefix + strings.TrimSuffix(strings.Repeat("("+placeholders(tagValueFieldSize)+"),", len(batch)), ",")
			args := make([]any, 0, len(batch)*tagValueFieldSize)
			for _, r := range batch {
				args = append(args, tagId, r.ObjectPk, r.TagValue, beid)
			}
			if _, err := session.ExecCtx(ctx, query, args...); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListByObjectPk 查某对象某主键值的标签画像：join dc_tag_def 出 tag_key/tag_name；
// tagged_at 可空用 ifnull 兜底避免扫描报错；select 4 列与 TagValueRow 4 字段严格对齐
func (m *defaultTagValueModel) ListByObjectPk(ctx context.Context, objectId, beid int64, objectPk string) ([]*TagValueRow, error) {
	query := "select d.`tag_key`, d.`tag_name`, v.`tag_value`, ifnull(v.`tagged_at`, now(3)) as `tagged_at` " +
		"from " + m.table + " v join `dc_tag_def` d on d.`id` = v.`tag_id` " +
		"where d.`object_id` = ? and v.`object_pk` = ?"
	args := []any{objectId, objectPk}
	if beid > 0 {
		query += " and d.`beid` = ?"
		args = append(args, beid)
	}
	var list []*TagValueRow
	if err := m.conn.QueryRowsCtx(ctx, &list, query+" order by d.`id` asc", args...); err != nil {
		return nil, err
	}
	return list, nil
}
