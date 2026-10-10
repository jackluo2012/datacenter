package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// infoTable information_schema.TABLES 查询行
type infoTable struct {
	TableName    string `db:"TABLE_NAME"`
	TableComment string `db:"TABLE_COMMENT"`
	Engine       string `db:"ENGINE"`
	TableRows    int64  `db:"TABLE_ROWS"`
}

// infoColumn information_schema.COLUMNS 查询行
type infoColumn struct {
	TableName     string `db:"TABLE_NAME"`
	ColumnName    string `db:"COLUMN_NAME"`
	ColumnComment string `db:"COLUMN_COMMENT"`
	ColumnType    string `db:"COLUMN_TYPE"`
	IsNullable    string `db:"IS_NULLABLE"`
}

// columnProps 列实体的扩展属性
type columnProps struct {
	ColumnType string `json:"column_type"`
	Nullable   bool   `json:"nullable"`
}

// columnItem 列实体及其所属表的 fqn（建 table->column contains 边用）
type columnItem struct {
	entity   *model.Entity
	tableFqn string
}

type CollectDatabaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCollectDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectDatabaseLogic {
	return &CollectDatabaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 采集库表元数据（含 contains 关系）
func (l *CollectDatabaseLogic) CollectDatabase(in *metadata.CollectDatabaseReq) (*metadata.CollectReply, error) {
	// 1. 参数校验（DSN 由网关编排解密后传入，本服务不做加解密）
	if in.DbType == "" || in.Dsn == "" || in.DbName == "" || in.Beid <= 0 {
		return nil, ErrInvalidCollectParams
	}

	// 2. 直连目标库读取 information_schema
	tables, columns, err := readInformationSchema(l.ctx, in.Dsn, in.DbName)
	if err != nil {
		return nil, err
	}

	// 3. 组装实体（fqn 规则：{db_type}://{db}[.{table}[.{column}]]）
	dbFqn := fmt.Sprintf("%s://%s", in.DbType, in.DbName)
	dbEntity := &model.Entity{
		EntityType:     model.EntityTypeDatabase,
		Fqn:            dbFqn,
		Name:           in.DbName,
		Classification: model.DefaultClassification,
		DsId:           in.DsId,
		Beid:           in.Beid,
	}
	tableEntities := make([]*model.Entity, 0, len(tables))
	for _, tb := range tables {
		tableEntities = append(tableEntities, &model.Entity{
			EntityType:     model.EntityTypeTable,
			Fqn:            fmt.Sprintf("%s.%s", dbFqn, tb.TableName),
			Name:           tb.TableName,
			Comment:        tb.TableComment,
			Engine:         tb.Engine,
			RowsCnt:        tb.TableRows,
			Classification: model.DefaultClassification,
			DsId:           in.DsId,
			Beid:           in.Beid,
		})
	}
	columnItems := make([]columnItem, 0, len(columns))
	for _, c := range columns {
		tableFqn := fmt.Sprintf("%s.%s", dbFqn, c.TableName)
		props, err := json.Marshal(columnProps{ColumnType: c.ColumnType, Nullable: c.IsNullable == "YES"})
		if err != nil {
			return nil, fmt.Errorf("列属性序列化失败: %w", err)
		}
		columnItems = append(columnItems, columnItem{
			tableFqn: tableFqn,
			entity: &model.Entity{
				EntityType:     model.EntityTypeColumn,
				Fqn:            fmt.Sprintf("%s.%s", tableFqn, c.ColumnName),
				Name:           c.ColumnName,
				Comment:        c.ColumnComment,
				PropsJson:      string(props),
				Classification: model.DefaultClassification,
				DsId:           in.DsId,
				Beid:           in.Beid,
			},
		})
	}

	// 4. 单事务写入：实体 upsert + contains 边幂等写入
	var tablesCreated, columnsCreated int64
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		columnEntities := make([]*model.Entity, len(columnItems))
		allFqns := make([]string, 0, 1+len(tableEntities)+len(columnItems))
		allFqns = append(allFqns, dbFqn)
		for _, te := range tableEntities {
			allFqns = append(allFqns, te.Fqn)
		}
		for i, item := range columnItems {
			columnEntities[i] = item.entity
			allFqns = append(allFqns, item.entity.Fqn)
		}

		// 4.1 预取写前已存在的实体，用于统计本次"新增"数
		before, err := l.svcCtx.EntityModel.FindFqnMap(txCtx, session, in.Beid, allFqns)
		if err != nil {
			return err
		}

		// 4.2 写入实体：database 1 条 + table/column 批量
		//     uk_beid_fqn 冲突时按 upsert 语义更新 comment/rows_cnt 等基础字段，不报错
		if err := l.svcCtx.EntityModel.Insert(txCtx, session, []*model.Entity{dbEntity}); err != nil {
			return err
		}
		if len(tableEntities) > 0 {
			if err := l.svcCtx.EntityModel.Insert(txCtx, session, tableEntities); err != nil {
				return err
			}
		}
		if len(columnEntities) > 0 {
			if err := l.svcCtx.EntityModel.Insert(txCtx, session, columnEntities); err != nil {
				return err
			}
		}

		// 4.3 回读 id 映射，构造 contains 边：database->table、table->column
		after, err := l.svcCtx.EntityModel.FindFqnMap(txCtx, session, in.Beid, allFqns)
		if err != nil {
			return err
		}
		dbId, ok := after[dbFqn]
		if !ok {
			return fmt.Errorf("database 实体写入后未找到: %s", dbFqn)
		}
		edges := make([]*model.Relation, 0, len(tableEntities)+len(columnItems))
		for _, te := range tableEntities {
			if id, ok := after[te.Fqn]; ok {
				edges = append(edges, &model.Relation{
					RelType: model.RelTypeContains,
					FromId:  dbId,
					ToId:    id,
					Beid:    in.Beid,
				})
			}
		}
		for _, item := range columnItems {
			id, ok := after[item.entity.Fqn]
			if !ok {
				continue
			}
			if tableId, ok := after[item.tableFqn]; ok {
				edges = append(edges, &model.Relation{
					RelType: model.RelTypeContains,
					FromId:  tableId,
					ToId:    id,
					Beid:    in.Beid,
				})
			}
		}
		if len(edges) > 0 {
			// uk_edge 冲突忽略，重复采集幂等
			if err := l.svcCtx.RelationModel.BatchInsert(txCtx, session, edges); err != nil {
				return err
			}
		}

		// 4.4 统计本次新增数（写前不存在、写后存在的 fqn 即为新建）
		for _, te := range tableEntities {
			if _, ok := before[te.Fqn]; !ok {
				tablesCreated++
			}
		}
		for _, ce := range columnEntities {
			if _, ok := before[ce.Fqn]; !ok {
				columnsCreated++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Infof("[CollectDatabase] beid=%d ds_id=%d db_type=%s db=%s 表=%d(新增%d) 列=%d(新增%d)",
		in.Beid, in.DsId, in.DbType, in.DbName, len(tables), tablesCreated, len(columns), columnsCreated)

	return &metadata.CollectReply{
		TablesCreated:  tablesCreated,
		ColumnsCreated: columnsCreated,
	}, nil
}

// readInformationSchema 直连目标库读取库下全部表与列的元信息
func readInformationSchema(ctx context.Context, dsn, dbName string) ([]*infoTable, []*infoColumn, error) {
	conn := sqlx.NewMysql(dsn)

	var tables []*infoTable
	// NULL 值兜底：视图行 ENGINE/TABLE_ROWS 可能为 NULL
	tableQuery := "select TABLE_NAME, ifnull(TABLE_COMMENT,'') as TABLE_COMMENT, " +
		"ifnull(ENGINE,'') as ENGINE, ifnull(TABLE_ROWS,0) as TABLE_ROWS " +
		"from information_schema.TABLES where TABLE_SCHEMA = ?"
	if err := conn.QueryRowsPartialCtx(ctx, &tables, tableQuery, dbName); err != nil {
		return nil, nil, fmt.Errorf("查询 information_schema.TABLES 失败: %w", err)
	}

	var columns []*infoColumn
	columnQuery := "select TABLE_NAME, COLUMN_NAME, ifnull(COLUMN_COMMENT,'') as COLUMN_COMMENT, " +
		"COLUMN_TYPE, ifnull(IS_NULLABLE,'') as IS_NULLABLE " +
		"from information_schema.COLUMNS where TABLE_SCHEMA = ? order by TABLE_NAME, ORDINAL_POSITION"
	if err := conn.QueryRowsPartialCtx(ctx, &columns, columnQuery, dbName); err != nil {
		return nil, nil, fmt.Errorf("查询 information_schema.COLUMNS 失败: %w", err)
	}
	return tables, columns, nil
}
