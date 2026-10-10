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

	"github.com/zeromicro/go-zero/core/logx"
)

type TaggingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaggingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaggingLogic {
	return &TaggingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行打标（核心链路）：
// tag_id>0 单标签打标（须为 rule 类型）；tag_id=0 时对 object 全部 status=1 且
// tag_type=rule 的标签逐个打标。每个标签：
// 对象表信息(ds_id/db_name/table_name) → datasource rpc 换目标库 DSN →
// 目标库执行 rule_sql（RawDB 手工扫描 object_id/tag_value 两列）→
// ReplaceBatch 事务内整体替换写 dc_tag_value；累计返回 tagged_rows。
func (l *TaggingLogic) Tagging(in *tag.TaggingReq) (*tag.TaggingReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}

	//1. 圈定待打标标签集合
	var defs []*model.TagDef
	if in.TagId > 0 {
		def, err := l.svcCtx.TagDefModel.FindOne(l.ctx, in.TagId, in.Beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrTagNotFound
			}
			return nil, err
		}
		//单标签仅要求 rule 类型（显式指定时不限制 status，便于停用标签手动补跑）
		if def.TagType != TagTypeRule {
			return nil, ErrNotRuleTag
		}
		defs = []*model.TagDef{def}
	} else {
		if in.ObjectId <= 0 {
			return nil, ErrTagTargetRequired
		}
		//批量：对象下全部启用中的 rule 标签
		list, err := l.svcCtx.TagDefModel.FindEnabledRuleByObject(l.ctx, in.ObjectId, in.Beid)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return &tag.TaggingReply{}, nil //无标签可打，空结果直接返回
		}
		defs = list
	}

	//2. 对象表信息（同批标签同属一个对象，查一次即可）
	objectId := in.ObjectId
	if in.TagId > 0 {
		objectId = defs[0].ObjectId
	}
	obj, err := l.svcCtx.ObjectModel.FindOne(l.ctx, objectId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	if obj.DsId <= 0 || strings.TrimSpace(obj.TableName) == "" {
		return nil, ErrObjectTargetMissing
	}

	//3. 构造目标库连接：datasource rpc GetDataSource(with_secret=true) 换 DSN
	dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         obj.DsId,
		Beid:       in.Beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: 数据源 %d: %v", ErrDatasource, obj.DsId, err)
	}
	if strings.TrimSpace(dsReply.GetDsn()) == "" {
		return nil, fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, obj.DsId)
	}
	//连接目标库（ServiceContext 按 DSN 缓存连接池）；RawDB 不可 Close，底层连接池需复用
	conn, err := l.svcCtx.TargetConn(dsReply.Dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	db, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}

	//4. 逐标签打标：跑 rule_sql → 事务内替换写 dc_tag_value，累计 tagged_rows
	var total int64
	for _, def := range defs {
		rows, err := execRuleSql(l.ctx, db, def.RuleSql)
		if err != nil {
			return nil, fmt.Errorf("%w: 标签 %d(%s): %v", ErrRuleExec, def.Id, def.TagKey, err)
		}
		if err := l.svcCtx.TagValueModel.ReplaceBatch(l.ctx, def.Id, def.Beid, rows); err != nil {
			return nil, fmt.Errorf("标签 %d(%s) 写入标签值失败: %w", def.Id, def.TagKey, err)
		}
		total += int64(len(rows))
		l.Infof("打标完成 object=%d tag=%d(%s) rows=%d", objectId, def.Id, def.TagKey, len(rows))
	}
	return &tag.TaggingReply{TaggedRows: total, TagId: in.TagId}, nil
}

// execRuleSql 在目标库执行规则 SQL（参考 datasvc QueryRowsAsMaps 的手工扫描思路）：
// 按结果集列名定位 object_id/tag_value 两列（列顺序不限），驱动返回的 []byte 统一转 string；
// 同批主键重复时后值覆盖前值（dc_tag_value 对 (tag_id,object_pk) 唯一，避免撞键）
func execRuleSql(ctx context.Context, db *sql.DB, ruleSql string) ([]*model.TagValue, error) {
	rs, err := db.QueryContext(ctx, ruleSql)
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
	if idIdx < 0 || valIdx < 0 {
		return nil, fmt.Errorf("结果集缺少 object_id/tag_value 列")
	}

	rows := make([]*model.TagValue, 0, 64)
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
		value := ruleCellToString(container[valIdx])
		if idx, ok := pos[pk]; ok {
			rows[idx].TagValue = value //后值覆盖前值
			continue
		}
		pos[pk] = len(rows)
		rows = append(rows, &model.TagValue{ObjectPk: pk, TagValue: value})
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

// ruleCellToString 单元格值转字符串（[]byte/string 原样，其余走 %v，nil 视为空串）
func ruleCellToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []byte:
		return strings.TrimSpace(string(t))
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}
