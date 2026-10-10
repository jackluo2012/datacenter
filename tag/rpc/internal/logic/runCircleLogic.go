package logic

import (
	"context"
	"errors"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunCircleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunCircleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunCircleLogic {
	return &RunCircleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行圈选（核心链路）：
// 人群定义 → 对象表信息(ds_id/db_name/table_name/pk_column) → datasource rpc 换目标库 DSN →
// 目标库执行 circle_rule_sql（RawDB 手工扫描 object_id 必备 + tag_value 可选两列）→
// 物化表自动创建 dc_tag_group_member_<group_id>(object_pk 主键, tag_value) →
// replace into 按 500/片批量写入 → 回写运行状态(est_count/last_run_at/物化表名)
func (l *RunCircleLogic) RunCircle(in *tag.RunCircleReq) (*tag.RunCircleReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	// 表名拼接前先校验 group_id 为正整数（杜绝表名注入）
	if in.GroupId <= 0 {
		return nil, ErrInvalidGroupId
	}
	groupModel := model.NewGroupModel(l.svcCtx.Conn)
	group, err := groupModel.FindOne(l.ctx, in.GroupId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	// 对象表信息 → 数据源 DSN → 目标库连接（与打标同一套装配）
	db, err := openObjectTargetDB(l.ctx, l.svcCtx, group.ObjectId, in.Beid)
	if err != nil {
		return nil, err
	}

	// 执行圈选 SQL：object_id 列必备，tag_value 成员值列可选
	rows, err := execCircleSql(l.ctx, db, group.CircleRuleSql)
	if err != nil {
		return nil, err
	}

	// 物化表自动创建 + replace into 分片批量写入（幂等，可重跑）
	tableName, err := GroupMaterializedTable(group.Id)
	if err != nil {
		return nil, err
	}
	if err := materializeGroupMembers(l.ctx, db, tableName, rows); err != nil {
		l.Errorf("人群物化写入失败, group_id: %d, table: %s, err: %v", group.Id, tableName, err)
		return nil, err
	}

	// 回写运行状态：est_count=成员数, last_run_at=now(3), materialized_table
	if err := groupModel.UpdateRunStatus(l.ctx, group.Id, group.Beid, int64(len(rows)), tableName); err != nil {
		return nil, err
	}
	l.Infof("圈选完成 group=%d(%s) members=%d table=%s", group.Id, group.Name, len(rows), tableName)
	return &tag.RunCircleReply{MemberCount: int64(len(rows)), Msg: "圈选完成"}, nil
}
