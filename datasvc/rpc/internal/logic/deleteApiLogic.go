package logic

import (
	"context"
	"errors"
	"fmt"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteApiLogic {
	return &DeleteApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除API（同时清理该 API 的授权关系）
func (l *DeleteApiLogic) DeleteApi(in *datasvc.IdReq) (*datasvc.OkReply, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("%w: id 非法", ErrParamInvalid)
	}
	api, err := l.svcCtx.ApiModel.FindOne(in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	//多租户校验
	if in.Beid != 0 && api.Beid != in.Beid {
		return nil, ErrApiNotFound
	}
	//先清理授权关系，再删 API
	if err := l.svcCtx.GrantModel.DeleteByApi(api.Id); err != nil {
		return nil, err
	}
	if err := l.svcCtx.ApiModel.Delete(api.Id); err != nil {
		return nil, err
	}
	l.Infof("删除API成功 id=%d beid=%d path=%s", api.Id, api.Beid, api.Path)
	return &datasvc.OkReply{Ok: true}, nil
}
