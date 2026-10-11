package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProfilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProfilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProfilesLogic {
	return &ListProfilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 剖析记录分页列表（beid 必带 + datasource_id/table_name 可选过滤，按 id 倒序）
func (l *ListProfilesLogic) ListProfiles(in *quality.ProfileListReq) (*quality.ProfileListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	// table_name 参与精确过滤，先过白名单防注入
	if in.TableName != "" && !ValidIdentifier(in.TableName) {
		return nil, errInvalidIdentifier("table_name", in.TableName)
	}
	list, total, err := l.svcCtx.ProfileModel.FindList(l.ctx, in.Beid, in.DatasourceId, in.Page, in.Size, in.TableName)
	if err != nil {
		return nil, err
	}
	profiles := make([]*quality.TableProfile, 0, len(list))
	for _, p := range list {
		profiles = append(profiles, profileToPb(p))
	}
	return &quality.ProfileListReply{List: profiles, Total: total}, nil
}
