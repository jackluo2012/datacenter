package logic

import (
	"context"
	"errors"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGroupMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGroupMembersLogic {
	return &ListGroupMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人群成员查询：物化表随圈选落在对象所在目标库，需先按 人群 → 对象 → 数据源 DSN
// 解析目标库连接；再查物化表 dc_tag_group_member_<group_id> 分页取 object_pk/tag_value
// 两列 + count 总数。从未圈选（物化表不存在）时按空结果返回
func (l *ListGroupMembersLogic) ListGroupMembers(in *tag.GroupMembersReq) (*tag.GroupMembersReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	// 表名拼接前先校验 group_id 为正整数（杜绝表名注入）
	if in.GroupId <= 0 {
		return nil, ErrInvalidGroupId
	}
	group, err := model.NewGroupModel(l.svcCtx.Conn).FindOne(l.ctx, in.GroupId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	db, err := openObjectTargetDB(l.ctx, l.svcCtx, group.ObjectId, in.Beid)
	if err != nil {
		return nil, err
	}
	tableName, err := GroupMaterializedTable(group.Id)
	if err != nil {
		return nil, err
	}
	members, total, err := listMaterializedMembers(l.ctx, db, tableName, in.Page, in.Size)
	if err != nil {
		l.Errorf("查询人群成员失败, group_id: %d, table: %s, err: %v", group.Id, tableName, err)
		return nil, err
	}
	return &tag.GroupMembersReply{Members: members, Total: total}, nil
}
