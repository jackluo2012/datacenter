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

type OfflineApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOfflineApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineApiLogic {
	return &OfflineApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 下线API（offline 后 ExecuteApi 以 ErrNotPublished 拒绝执行）
func (l *OfflineApiLogic) OfflineApi(in *datasvc.IdReq) (*datasvc.OkReply, error) {
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
	if err := l.svcCtx.ApiModel.UpdateStatus(api.Id, "offline"); err != nil {
		return nil, err
	}
	l.Infof("下线API成功 id=%d beid=%d path=%s", api.Id, api.Beid, api.Path)
	return &datasvc.OkReply{Ok: true}, nil
}
