package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建人群（圈群定义）：name/object_id 必填 → 对象须存在（带租户隔离）→
// circle_rule_sql 强校验（select 单语句 + as object_id 别名，成员值列可选，禁分号防注入）
// → 落库（est_count/物化表名/最近圈选时间留待 RunCircle 回填）
func (l *CreateGroupLogic) CreateGroup(in *tag.CreateGroupReq) (*tag.GroupReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name 不能为空", ErrInvalidParams)
	}
	if in.ObjectId <= 0 {
		return nil, fmt.Errorf("%w: object_id 不能为空", ErrInvalidParams)
	}
	// 所属对象必须存在（带租户隔离）
	if _, err := l.svcCtx.ObjectModel.FindOne(l.ctx, in.ObjectId, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	// 圈选 SQL 强校验（复用规则 SQL 同一套安全底线，成员值列可选）
	if err := ValidateCircleRuleSql(in.CircleRuleSql); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCircleSql, err)
	}

	groupModel := model.NewGroupModel(l.svcCtx.Conn)
	id, err := groupModel.Insert(l.ctx, &model.Group{
		Name:          name,
		ObjectId:      in.ObjectId,
		CircleRuleSql: strings.TrimSpace(in.CircleRuleSql),
		Beid:          in.Beid,
	})
	if err != nil {
		l.Errorf("创建人群失败, name: %s, object_id: %d, beid: %d, err: %v", name, in.ObjectId, in.Beid, err)
		return nil, err
	}
	group, err := groupModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &tag.GroupReply{Group: toPbGroup(group)}, nil
}
