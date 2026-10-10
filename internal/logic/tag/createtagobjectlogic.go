// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTagObjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建标签对象
func NewCreateTagObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagObjectLogic {
	return &CreateTagObjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 创建标签对象: 网关请求逐字段映射为 rpc 请求, 对象类型缺省为 user, 转发后把对象转为网关类型返回
func (l *CreateTagObjectLogic) CreateTagObject(req *types.CreateTagObjectReq) (resp *types.TagObject, err error) {
	objectType := req.ObjectType
	if objectType == "" {
		objectType = "user"
	}
	reply, err := l.svcCtx.TagRpc.CreateObject(l.ctx, &tagClient.CreateObjectReq{
		ObjectType: objectType,
		Name:       req.Name,
		DsId:       req.DsId,
		DbName:     req.DbName,
		TableName:  req.TableName,
		PkColumn:   req.PkColumn,
		Beid:       req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return tagObjectToTypes(reply.GetObject()), nil
}
