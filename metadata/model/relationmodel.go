package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	relationTable       = "`dc_meta_relation`"
	relationInsertChunk = 500 // 批量插入分片大小
)

// relationInsertColumns 插入字段（不含自增 id 与 created_at）
const relationInsertColumns = "rel_type,from_id,to_id,props_json,beid"

// relationRows 查询列（props_json 允许 NULL，用 ifnull 兜底避免扫描报错）
const relationRows = "id,rel_type,from_id,to_id,ifnull(props_json,'') as props_json,beid,created_at"

// 元数据关系（contains/lineage/depends）
type (
	// Relation dc_meta_relation 行
	Relation struct {
		Id        int64     `db:"id"`
		RelType   string    `db:"rel_type"`
		FromId    int64     `db:"from_id"`
		ToId      int64     `db:"to_id"`
		PropsJson string    `db:"props_json"`
		Beid      int64     `db:"beid"`
		CreatedAt time.Time `db:"created_at"`
	}

	RelationModel interface {
		// Insert 幂等写入单条关系（uk_edge 冲突时忽略）
		Insert(ctx context.Context, session sqlx.Session, rel *Relation) error
		// BatchInsert 幂等批量写入关系
		BatchInsert(ctx context.Context, session sqlx.Session, rels []*Relation) error
		// FindNeighbors 查询与 entityId 直连的边
		// upstream：以 entityId 为终点(to_id)的边，邻居在 from_id 侧
		// downstream：以 entityId 为起点(from_id)的边，邻居在 to_id 侧
		FindNeighbors(ctx context.Context, relType string, entityId int64, direction string) ([]*Relation, error)
		// ListByFrom 按 from_id 查询（relType 为空查全部类型）
		ListByFrom(ctx context.Context, relType string, fromId int64) ([]*Relation, error)
		// ListByTo 按 to_id 查询（relType 为空查全部类型）
		ListByTo(ctx context.Context, relType string, toId int64) ([]*Relation, error)
	}

	defaultRelationModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewRelationModel 创建元数据关系 model
func NewRelationModel(conn sqlx.SqlConn) RelationModel {
	return &defaultRelationModel{
		conn:  conn,
		table: relationTable,
	}
}

// Insert 幂等写入单条关系：uk_edge(rel_type,from_id,to_id) 冲突时 insert ignore 跳过
func (m *defaultRelationModel) Insert(ctx context.Context, session sqlx.Session, rel *Relation) error {
	query := fmt.Sprintf("insert ignore into %s (%s) values (?,?,?,?,?)", m.table, relationInsertColumns)
	_, err := execCtx(ctx, m.conn, session, query, rel.RelType, rel.FromId, rel.ToId, rel.PropsJson, rel.Beid)
	return err
}

// BatchInsert 幂等批量写入关系：分片拼接多行 values，uk_edge 冲突时忽略
func (m *defaultRelationModel) BatchInsert(ctx context.Context, session sqlx.Session, rels []*Relation) error {
	if len(rels) == 0 {
		return nil
	}
	prefix := fmt.Sprintf("insert ignore into %s (%s) values ", m.table, relationInsertColumns)
	for start := 0; start < len(rels); start += relationInsertChunk {
		end := start + relationInsertChunk
		if end > len(rels) {
			end = len(rels)
		}
		batch := rels[start:end]
		rows := strings.TrimSuffix(strings.Repeat("("+placeholders(5)+"),", len(batch)), ",")
		query := prefix + rows
		args := make([]any, 0, len(batch)*5)
		for _, r := range batch {
			args = append(args, r.RelType, r.FromId, r.ToId, r.PropsJson, r.Beid)
		}
		if _, err := execCtx(ctx, m.conn, session, query, args...); err != nil {
			return fmt.Errorf("批量写入关系失败: %w", err)
		}
	}
	return nil
}

// FindNeighbors 查询与 entityId 直连的边（血缘 BFS 逐层展开用）
func (m *defaultRelationModel) FindNeighbors(ctx context.Context, relType string,
	entityId int64, direction string) ([]*Relation, error) {
	var where string
	switch direction {
	case DirectionUpstream:
		// 上游：当前实体是边的终点，邻居在 from_id 侧
		where = "to_id = ?"
	case DirectionDownstream:
		// 下游：当前实体是边的起点，邻居在 to_id 侧
		where = "from_id = ?"
	default:
		return nil, ErrInvalidDirection
	}
	query := fmt.Sprintf("select %s from %s where rel_type = ? and %s", relationRows, m.table, where)
	var list []*Relation
	if err := m.conn.QueryRowsCtx(ctx, &list, query, relType, entityId); err != nil {
		return nil, err
	}
	return list, nil
}

// ListByFrom 按 from_id 查询出边（relType 为空查全部类型）
func (m *defaultRelationModel) ListByFrom(ctx context.Context, relType string, fromId int64) ([]*Relation, error) {
	query := fmt.Sprintf("select %s from %s where from_id = ?", relationRows, m.table)
	args := []any{fromId}
	if relType != "" {
		query += " and rel_type = ?"
		args = append(args, relType)
	}
	var list []*Relation
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// ListByTo 按 to_id 查询入边（relType 为空查全部类型）
func (m *defaultRelationModel) ListByTo(ctx context.Context, relType string, toId int64) ([]*Relation, error) {
	query := fmt.Sprintf("select %s from %s where to_id = ?", relationRows, m.table)
	args := []any{toId}
	if relType != "" {
		query += " and rel_type = ?"
		args = append(args, relType)
	}
	var list []*Relation
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}
