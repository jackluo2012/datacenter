package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrantApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrantApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantApiLogic {
	return &GrantApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 授权 API 给应用（同 api+app 重复授权幂等：刷新过期时间与脱敏字段）
func (l *GrantApiLogic) GrantApi(in *datasvc.GrantApiReq) (*datasvc.OkReply, error) {
	if in.ApiId <= 0 || in.AppId <= 0 {
		return nil, ErrParamInvalid
	}
	//API 必须存在
	api, err := l.svcCtx.ApiModel.FindOne(in.ApiId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	//多租户校验（beid 未传时回落 API 所属租户）
	beid := in.Beid
	if beid == 0 {
		beid = api.Beid
	}
	if api.Beid != beid {
		return nil, ErrApiNotFound
	}
	//应用必须存在且同租户
	app, err := l.svcCtx.AppModel.FindOne(in.AppId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}
	if app.Beid != beid {
		return nil, ErrAppNotFound
	}
	//expire_at：0 表示永久（NULL）
	var expireAt sql.NullTime
	if in.ExpireAt > 0 {
		expireAt = sql.NullTime{Time: time.Unix(in.ExpireAt, 0), Valid: true}
	}
	grant := &model.ApiGrant{
		ApiId:     in.ApiId,
		AppId:     in.AppId,
		ExpireAt:  expireAt,
		FieldMask: strings.TrimSpace(in.FieldMask),
		Beid:      beid,
	}
	if err := l.svcCtx.GrantModel.Insert(grant); err != nil {
		return nil, err
	}
	l.Infof("授权API成功 api=%d app=%d beid=%d expire_at=%v field_mask=%s",
		grant.ApiId, grant.AppId, grant.Beid, grant.ExpireAt, grant.FieldMask)
	return &datasvc.OkReply{Ok: true}, nil
}
