package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"
)

// 圈群业务错误（error.go 为既有文件不可改，圈群新增错误集中在本文件）
var (
	// ErrInvalidGroupId group_id 非正整数（表名拼接前必须拦截）
	ErrInvalidGroupId = errors.New("group_id 必须为正整数")
	// ErrGroupNotFound 人群不存在
	ErrGroupNotFound = errors.New("人群不存在")
	// ErrInvalidCircleSql 圈选 SQL 未通过校验（非 select 开头/缺 as object_id 别名/含分号等）
	ErrInvalidCircleSql = errors.New("circle_rule_sql 校验未通过")
	// ErrCircleExec 圈选 SQL 或物化表读写执行失败
	ErrCircleExec = errors.New("圈选 SQL 执行失败")
)

const (
	// groupMemberChunk 物化表批量写入分片大小（与 dc_tag_value ReplaceBatch 的 500/片对齐）
	groupMemberChunk = 500
	// 人群成员分页参数防御（与 model 侧 tagDef/object 分页口径一致）
	groupMemberDefaultPageSize = 10
	groupMemberMaxPageSize     = 200
)

// ValidateCircleRuleSql 圈选 SQL 校验（纯函数，便于单测），复用规则 SQL 同一套安全底线：
// 1) 非空；2) 必须以 select 开头（只读限制，杜绝 update/delete 等写语句）；
// 3) 必须包含 as object_id 别名列（圈选结果按该列取成员主键）；
// 4) 禁止分号（单语句限制，天然拒绝 `; drop table ...` 等注入片段）。
// 与打标 rule_sql（ValidateRuleSql）的差异：as tag_value 成员值列可选（纯圈人场景无值列）
func ValidateCircleRuleSql(sqlText string) error {
	s := strings.TrimSpace(sqlText)
	if s == "" {
		return fmt.Errorf("circle_rule_sql 不能为空")
	}
	if len(s) < 6 || !strings.EqualFold(s[:6], "select") {
		return fmt.Errorf("circle_rule_sql 必须以 select 开头")
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("circle_rule_sql 仅允许单条语句，禁止分号（拒绝 ; drop 等注入片段）")
	}
	if !ruleAliasObjectIdRe.MatchString(s) {
		return fmt.Errorf("circle_rule_sql 缺少 as object_id 别名列")
	}
	return nil
}

// GroupMaterializedTable 构造人群物化表名 dc_tag_group_member_<group_id>。
// group_id 必须为正整数才允许拼接表名（int64 类型化入参 + 正整数校验双保险，
// 杜绝 `member_1;drop table x` 类表名注入）
func GroupMaterializedTable(groupId int64) (string, error) {
	if groupId <= 0 {
		return "", fmt.Errorf("%w: %d", ErrInvalidGroupId, groupId)
	}
	return fmt.Sprintf("dc_tag_group_member_%d", groupId), nil
}

// circleRow 圈选结果行（成员主键 + 可选成员值），圈选执行与物化写入间的内存载体
type circleRow struct {
	ObjectPk string
	TagValue string
}

// openObjectTargetDB 按对象配置解析目标库连接（RunCircle/ListGroupMembers 共用，
// 与 taggingLogic.Tagging 同一套装配思路）：
// object(ds_id/db_name/table_name/pk_column) → datasource rpc(with_secret=true) 换 DSN →
// ServiceContext 按 DSN 复用目标库连接池 → RawDB。db_name/pk_column 已由圈选 SQL 中的
// 库表限定与 as object_id 别名承载，此处仅作可达性校验
func openObjectTargetDB(ctx context.Context, svcCtx *svc.ServiceContext, objectId, beid int64) (*sql.DB, error) {
	obj, err := svcCtx.ObjectModel.FindOne(ctx, objectId, beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	if obj.DsId <= 0 || strings.TrimSpace(obj.TableName) == "" {
		return nil, ErrObjectTargetMissing
	}
	dsReply, err := svcCtx.DataSourceRpc.GetDataSource(ctx, &dataSourceClient.GetDataSourceReq{
		Id:         obj.DsId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: 数据源 %d: %v", ErrDatasource, obj.DsId, err)
	}
	if strings.TrimSpace(dsReply.GetDsn()) == "" {
		return nil, fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, obj.DsId)
	}
	// 连接目标库（按 DSN 缓存连接池）；RawDB 不可 Close，底层连接池需复用
	conn, err := svcCtx.TargetConn(dsReply.Dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	db, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	return db, nil
}

// execCircleSql 在目标库执行圈选 SQL（参考 taggingLogic 的 execRuleSql 手工扫描思路）：
// 按结果集列名定位 object_id 列（必备）与 tag_value 列（可选，缺省置空串），列顺序不限；
// 驱动返回的 []byte 统一转 string；同批主键重复时后值覆盖前值
// （物化表 object_pk 为主键，replace into 语义下保持一行）
func execCircleSql(ctx context.Context, db *sql.DB, circleSql string) ([]*circleRow, error) {
	rs, err := db.QueryContext(ctx, circleSql)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	cols, err := rs.Columns()
	if err != nil {
		return nil, err
	}
	idIdx, valIdx := -1, -1
	for i, c := range cols {
		switch strings.ToLower(strings.TrimSpace(c)) {
		case "object_id":
			idIdx = i
		case "tag_value":
			valIdx = i
		}
	}
	if idIdx < 0 {
		return nil, fmt.Errorf("结果集缺少 object_id 列")
	}

	rows := make([]*circleRow, 0, 64)
	pos := make(map[string]int, 64) //object_pk -> rows 下标（主键去重）
	for rs.Next() {
		container := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range container {
			ptrs[i] = &container[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, err
		}
		pk := ruleCellToString(container[idIdx])
		if pk == "" {
			continue //主键值为空的行跳过
		}
		var value string
		if valIdx >= 0 {
			value = ruleCellToString(container[valIdx])
		}
		if idx, ok := pos[pk]; ok {
			rows[idx].TagValue = value //后值覆盖前值
			continue
		}
		pos[pk] = len(rows)
		rows = append(rows, &circleRow{ObjectPk: pk, TagValue: value})
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

// materializeGroupMembers 在目标库物化人群成员（幂等，可重复圈选）。
// 先自动建表：create table if not exists dc_tag_group_member_<group_id>
// (object_pk varchar(128) primary key, tag_value varchar(255))；
// 再 replace into 按 500/片分批写入（replace 语义：同主键覆盖，重跑整体替换）。
// tableName 由 GroupMaterializedTable 基于"已校验正整数"的 group_id 构造，无注入面
func materializeGroupMembers(ctx context.Context, db *sql.DB, tableName string, rows []*circleRow) error {
	// 物化表自动创建（已存在则跳过，历史结构留档）
	createDdl := "create table if not exists `" + tableName + "` (" +
		"`object_pk` varchar(128) primary key, `tag_value` varchar(255))"
	if _, err := db.ExecContext(ctx, createDdl); err != nil {
		return fmt.Errorf("创建物化表 %s 失败: %w", tableName, err)
	}
	if len(rows) == 0 {
		return nil //空结果也是一次有效圈选（清空语义由 replace into 幂等保证：本次无写入）
	}
	// replace into 分片批量写入
	prefix := "replace into `" + tableName + "` (`object_pk`,`tag_value`) values "
	for start := 0; start < len(rows); start += groupMemberChunk {
		end := start + groupMemberChunk
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[start:end]
		query := prefix + strings.TrimSuffix(strings.Repeat("(?,?),", len(batch)), ",")
		args := make([]any, 0, len(batch)*2)
		for _, r := range batch {
			args = append(args, r.ObjectPk, r.TagValue)
		}
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

// listMaterializedMembers 分页查询人群物化表：select object_pk,tag_value limit/offset + count 总数。
// 物化表尚未创建（从未圈选）时按空结果返回而非报错
func listMaterializedMembers(ctx context.Context, db *sql.DB, tableName string, page, size int64) ([]*tag.GroupMember, int64, error) {
	// 分页参数防御
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = groupMemberDefaultPageSize
	}
	if size > groupMemberMaxPageSize {
		size = groupMemberMaxPageSize
	}

	// 总数（物化表不存在=从未圈选，返回空结果）
	var total int64
	countQuery := "select count(*) from `" + tableName + "`"
	if err := db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		if isMysqlTableNotExist(err) {
			return []*tag.GroupMember{}, 0, nil
		}
		return nil, 0, err
	}

	rs, err := db.QueryContext(ctx, "select `object_pk`, `tag_value` from `"+tableName+"` limit ? offset ?", size, (page-1)*size)
	if err != nil {
		if isMysqlTableNotExist(err) {
			return []*tag.GroupMember{}, 0, nil
		}
		return nil, 0, err
	}
	defer rs.Close()
	members := make([]*tag.GroupMember, 0, size)
	for rs.Next() {
		container := make([]interface{}, 2)
		ptrs := make([]interface{}, 2)
		for i := range container {
			ptrs[i] = &container[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, 0, err
		}
		members = append(members, &tag.GroupMember{
			ObjectPk: ruleCellToString(container[0]),
			TagValue: ruleCellToString(container[1]),
		})
	}
	if err := rs.Err(); err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

// isMysqlTableNotExist 识别 MySQL 表不存在错误（错误码 1146 / 文本 doesn't exist），
// 用于人群从未圈选时物化表缺失的优雅降级
func isMysqlTableNotExist(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "1146") || strings.Contains(msg, "doesn't exist")
}

// toPbGroup model → pb（标签人群；convert.go 为既有文件不可改，圈群转换集中在本文件）
func toPbGroup(g *model.Group) *tag.TagGroup {
	if g == nil {
		return nil
	}
	return &tag.TagGroup{
		Id:                g.Id,
		Name:              g.Name,
		ObjectId:          g.ObjectId,
		CircleRuleSql:     g.CircleRuleSql,
		EstCount:          g.EstCount,
		MaterializedTable: g.MaterializedTable,
		LastRunAt:         g.LastRunAt,
		Beid:              g.Beid,
	}
}
