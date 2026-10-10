package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 实体类型
const (
	EntityTypeDatabase = "database"
	EntityTypeTable    = "table"
	EntityTypeColumn   = "column"
)

const (
	entityTable           = "`dc_meta_entity`"
	entityInsertFieldSize = 11  // 插入字段数（与 entityInsertColumns 对应）
	entityInsertChunk     = 200 // 批量插入分片大小
	fqnQueryChunk         = 500 // fqn in 查询分片大小
	idQueryChunk          = 500 // id in 查询分片大小
)

// entityInsertColumns 插入字段（不含自增 id 与时间列）
const entityInsertColumns = "entity_type,fqn,name,`comment`,engine,rows_cnt,props_json,classification,owner,ds_id,beid"

// entityRows 查询列（props_json 允许 NULL，用 ifnull 兜底避免扫描报错）
const entityRows = "id,entity_type,fqn,name,`comment`,engine,rows_cnt,ifnull(props_json,'') as props_json,classification,owner,ds_id,beid,created_at,updated_at"

// entityUpsertClause uk_beid_fqn 冲突时的更新子句：
// 覆盖采集可得的基础字段；classification/owner 为人工维护字段，不做覆盖
const entityUpsertClause = " on duplicate key update name = values(name), `comment` = values(`comment`), " +
	"engine = values(engine), rows_cnt = values(rows_cnt), props_json = values(props_json), ds_id = values(ds_id)"

// 元数据实体（database/table/column）
type (
	// Entity dc_meta_entity 行
	Entity struct {
		Id             int64     `db:"id"`
		EntityType     string    `db:"entity_type"`
		Fqn            string    `db:"fqn"`
		Name           string    `db:"name"`
		Comment        string    `db:"comment"`
		Engine         string    `db:"engine"`
		RowsCnt        int64     `db:"rows_cnt"`
		PropsJson      string    `db:"props_json"`
		Classification string    `db:"classification"`
		Owner          string    `db:"owner"`
		DsId           int64     `db:"ds_id"`
		Beid           int64     `db:"beid"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	EntityModel interface {
		// Insert 批量写入实体（uk_beid_fqn 冲突时按 upsert 语义更新基础字段）
		Insert(ctx context.Context, session sqlx.Session, entities []*Entity) error
		// FindOne 按主键查询（beid>0 时追加多租户过滤）
		FindOne(ctx context.Context, id, beid int64) (*Entity, error)
		// FindOneByFqn 按 beid + 全限定名查询
		FindOneByFqn(ctx context.Context, beid int64, fqn string) (*Entity, error)
		// FindList 条件分页查询（entity_type/keyword/ds_id 均可选，keyword 对 fqn/name 模糊匹配）
		FindList(ctx context.Context, beid int64, entityType, keyword string, dsId, page, size int64) ([]*Entity, int64, error)
		// FindFqnMap 批量查询 fqn -> id 映射（采集建边用）
		FindFqnMap(ctx context.Context, session sqlx.Session, beid int64, fqns []string) (map[string]int64, error)
		// FindListByIds 按主键批量查询（血缘节点信息回填用，beid>0 时追加多租户过滤）
		FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Entity, error)
		// UpdateProps 更新实体扩展属性
		UpdateProps(ctx context.Context, id int64, propsJson string) error
	}

	defaultEntityModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewEntityModel 创建元数据实体 model
func NewEntityModel(conn sqlx.SqlConn) EntityModel {
	return &defaultEntityModel{
		conn:  conn,
		table: entityTable,
	}
}

// Insert 批量写入实体：分片拼接多行 values，uk_beid_fqn 冲突时走 upsert 更新而非报错
func (m *defaultEntityModel) Insert(ctx context.Context, session sqlx.Session, entities []*Entity) error {
	if len(entities) == 0 {
		return nil
	}
	prefix := fmt.Sprintf("insert into %s (%s) values ", m.table, entityInsertColumns)
	for start := 0; start < len(entities); start += entityInsertChunk {
		end := start + entityInsertChunk
		if end > len(entities) {
			end = len(entities)
		}
		batch := entities[start:end]
		rows := strings.TrimSuffix(strings.Repeat("("+placeholders(entityInsertFieldSize)+"),", len(batch)), ",")
		query := prefix + rows + entityUpsertClause
		args := make([]any, 0, len(batch)*entityInsertFieldSize)
		for _, e := range batch {
			args = append(args, e.EntityType, e.Fqn, e.Name, e.Comment, e.Engine,
				e.RowsCnt, e.PropsJson, e.Classification, e.Owner, e.DsId, e.Beid)
		}
		if _, err := execCtx(ctx, m.conn, session, query, args...); err != nil {
			return fmt.Errorf("批量写入实体失败: %w", err)
		}
	}
	return nil
}

// FindOne 按主键查询实体
func (m *defaultEntityModel) FindOne(ctx context.Context, id, beid int64) (*Entity, error) {
	query := fmt.Sprintf("select %s from %s where id = ?", entityRows, m.table)
	args := []any{id}
	if beid > 0 {
		query += " and beid = ?"
		args = append(args, beid)
	}
	var resp Entity
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

// FindOneByFqn 按 beid + 全限定名查询实体
func (m *defaultEntityModel) FindOneByFqn(ctx context.Context, beid int64, fqn string) (*Entity, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	query := fmt.Sprintf("select %s from %s where beid = ? and fqn = ? limit 1", entityRows, m.table)
	var resp Entity
	err := m.conn.QueryRowCtx(ctx, &resp, query, beid, fqn)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlx.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindList 条件分页查询实体，返回列表与总数
func (m *defaultEntityModel) FindList(ctx context.Context, beid int64, entityType, keyword string,
	dsId, page, size int64) ([]*Entity, int64, error) {
	if beid <= 0 {
		return nil, 0, ErrBeidRequired
	}
	where := "beid = ?"
	args := []any{beid}
	if entityType != "" {
		where += " and entity_type = ?"
		args = append(args, entityType)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " and (fqn like ? or name like ?)"
		like := "%" + escapeLike(kw) + "%"
		args = append(args, like, like)
	}
	if dsId > 0 {
		where += " and ds_id = ?"
		args = append(args, dsId)
	}

	// 总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) as total from %s where %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	listQuery := fmt.Sprintf("select %s from %s where %s order by id desc limit ?, ?", entityRows, m.table, where)
	args = append(args, (page-1)*size, size)
	var list []*Entity
	if err := m.conn.QueryRowsCtx(ctx, &list, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindFqnMap 批量查询 fqn -> id 映射，fqns 自动去重并分片
func (m *defaultEntityModel) FindFqnMap(ctx context.Context, session sqlx.Session,
	beid int64, fqns []string) (map[string]int64, error) {
	if beid <= 0 {
		return nil, ErrBeidRequired
	}
	result := make(map[string]int64, len(fqns))
	if len(fqns) == 0 {
		return result, nil
	}

	// 去重，避免 in 条件里出现重复占位
	seen := make(map[string]struct{}, len(fqns))
	uniq := make([]string, 0, len(fqns))
	for _, f := range fqns {
		if f == "" {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		uniq = append(uniq, f)
	}

	for start := 0; start < len(uniq); start += fqnQueryChunk {
		end := start + fqnQueryChunk
		if end > len(uniq) {
			end = len(uniq)
		}
		part := uniq[start:end]
		query := fmt.Sprintf("select id,fqn from %s where beid = ? and fqn in (%s)",
			m.table, placeholders(len(part)))
		args := make([]any, 0, len(part)+1)
		args = append(args, beid)
		for _, f := range part {
			args = append(args, f)
		}
		// 扫描目标必须是"列数与字段数一致"的轻量结构（go-zero 对少列查询扫宽 struct 会报 not matching destination）
		var rows []*idFqnRow
		if err := queryRowsCtx(ctx, m.conn, session, &rows, query, args...); err != nil {
			return nil, err
		}
		for _, r := range rows {
			result[r.Fqn] = r.Id
		}
	}
	return result, nil
}

// idFqnRow FindFqnMap 专用扫描结构（与 select id,fqn 列数对齐）
type idFqnRow struct {
	Id  int64  `db:"id"`
	Fqn string `db:"fqn"`
}

// FindListByIds 按主键批量查询实体（自动去重并分片）
func (m *defaultEntityModel) FindListByIds(ctx context.Context, beid int64, ids []int64) ([]*Entity, error) {
	if len(ids) == 0 {
		return []*Entity{}, nil
	}

	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}

	list := make([]*Entity, 0, len(uniq))
	for start := 0; start < len(uniq); start += idQueryChunk {
		end := start + idQueryChunk
		if end > len(uniq) {
			end = len(uniq)
		}
		part := uniq[start:end]
		query := fmt.Sprintf("select %s from %s where id in (%s)", entityRows, m.table, placeholders(len(part)))
		args := make([]any, 0, len(part)+1)
		for _, id := range part {
			args = append(args, id)
		}
		if beid > 0 {
			query += " and beid = ?"
			args = append(args, beid)
		}
		var rows []*Entity
		if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
			return nil, err
		}
		list = append(list, rows...)
	}
	return list, nil
}

// UpdateProps 更新实体扩展属性
func (m *defaultEntityModel) UpdateProps(ctx context.Context, id int64, propsJson string) error {
	query := fmt.Sprintf("update %s set props_json = ? where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, propsJson, id)
	return err
}
