package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishApiLogic {
	return &PublishApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布API（前置校验：sql_text 非空、params_json 合法）
func (l *PublishApiLogic) PublishApi(in *datasvc.IdReq) (*datasvc.OkReply, error) {
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
	//发布前置校验：SQL 模板非空
	if strings.TrimSpace(api.SqlText) == "" {
		return nil, fmt.Errorf("%w: api=%d 发布前需配置 SQL 模板", ErrSqlEmpty, api.Id)
	}
	//发布前置校验：参数定义必须合法，避免上线后执行链才发现
	if _, err := ParseParamDefs(api.ParamsJson); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
	}
	if err := l.svcCtx.ApiModel.UpdateStatus(api.Id, "published"); err != nil {
		return nil, err
	}
	l.Infof("发布API成功 id=%d beid=%d path=%s", api.Id, api.Beid, api.Path)
	return &datasvc.OkReply{Ok: true}, nil
}
