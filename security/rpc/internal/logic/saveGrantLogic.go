package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveGrantLogic {
	return &SaveGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存授权（subject+resource+resource_id 为自然键 upsert，重复授权覆盖密级与过期时间）：
// subject 需为 type:id（type ∈ user/app/role）；level_code 白名单且必须在分级目录；
// expire_at=0 表示永久（落库 NULL），>0 为秒级时间戳
func (l *SaveGrantLogic) SaveGrant(in *security.SaveGrantReq) (*security.GrantReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	subject := strings.TrimSpace(in.Subject)
	if !ValidSubject(subject) {
		return nil, ErrSubjectInvalid
	}
	resource := strings.TrimSpace(in.Resource)
	if resource == "" {
		return nil, ErrResourceRequired
	}
	levelCode := strings.TrimSpace(in.LevelCode)
	if !ValidLevelCode(levelCode) {
		return nil, ErrLevelCodeInvalid
	}
	// 许可密级必须在分级目录（防授权密级打错导致鉴权 fail-closed 全拒）
	if _, err := l.svcCtx.LevelModel.FindByCode(l.ctx, in.Beid, levelCode); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrLevelRefNotFound
		}
		return nil, err
	}

	_, err := l.svcCtx.GrantModel.Save(l.ctx, &model.Grant{
		Subject:    subject,
		Resource:   resource,
		ResourceId: in.ResourceId,
		LevelCode:  levelCode,
		ExpireAt:   expireAtToNullTime(in.ExpireAt),
		Beid:       in.Beid,
	})
	if err != nil {
		return nil, err
	}

	// 回读保证返回与库内一致
	grant, err := l.svcCtx.GrantModel.FindBySubRes(l.ctx, in.Beid, subject, resource, in.ResourceId)
	if err != nil {
		return nil, err
	}
	return &security.GrantReply{Grant: grantToPb(grant)}, nil
}
