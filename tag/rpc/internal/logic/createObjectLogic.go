package logic

import (
	"context"
	"fmt"
	"strings"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateObjectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateObjectLogic {
	return &CreateObjectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建标签对象：object_type 走白名单（user/member/goods），
// pk_column 未传时落默认值 id，保证后续打标链路可用
func (l *CreateObjectLogic) CreateObject(in *tag.CreateObjectReq) (*tag.ObjectReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if !ValidObjectType(in.ObjectType) {
		return nil, ErrInvalidObjectType
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: 对象名不能为空", ErrInvalidParams)
	}
	pkColumn := strings.TrimSpace(in.PkColumn)
	if pkColumn == "" {
		pkColumn = defaultPkColumn
	}
	id, err := l.svcCtx.ObjectModel.Insert(l.ctx, &model.TagObject{
		ObjectType: in.ObjectType,
		Name:       strings.TrimSpace(in.Name),
		DsId:       in.DsId,
		DbName:     in.DbName,
		TableName:  strings.TrimSpace(in.TableName),
		PkColumn:   pkColumn,
		Beid:       in.Beid,
	})
	if err != nil {
		l.Errorf("创建标签对象失败, name: %s, beid: %d, err: %v", in.Name, in.Beid, err)
		return nil, err
	}
	obj, err := l.svcCtx.ObjectModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &tag.ObjectReply{Object: toPbObject(obj)}, nil
}
