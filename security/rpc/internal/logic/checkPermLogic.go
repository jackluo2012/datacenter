package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermLogic {
	return &CheckPermLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 鉴权检查：查主体对资源的授权 → 过期判断 → 数据密级与许可密级比较。
// 除"数据密级 ≤ 许可密级且授权有效"外一律 allowed=false（fail-closed）：
// 无授权/授权过期/密级非法/密级超高均拒绝
func (l *CheckPermLogic) CheckPerm(in *security.CheckPermReq) (*security.CheckPermReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	subject := strings.TrimSpace(in.Subject)
	if !ValidSubject(subject) {
		return nil, ErrSubjectInvalid
	}
	// 授权按 uk_sub_res 精确定位；无授权不报错，按拒绝返回
	grant, err := l.svcCtx.GrantModel.FindBySubRes(l.ctx, in.Beid, subject, in.Resource, in.ResourceId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &security.CheckPermReply{Allowed: false, Reason: "未授权"}, nil
		}
		return nil, err
	}
	// 授权过期（expire_at=0 永久）
	if GrantExpired(grantToExpireAt(grant), time.Now().Unix()) {
		return &security.CheckPermReply{Allowed: false, Reason: "授权已过期"}, nil
	}
	// 密级比较（纯函数）：数据密级 ≤ 许可密级放行，未知密级一律拒绝
	allowed, reason := DecidePerm(grant.LevelCode, in.LevelCode)
	return &security.CheckPermReply{Allowed: allowed, Reason: reason}, nil
}

// grantToExpireAt 授权行过期时间转秒级时间戳（NULL → 0 永久）
func grantToExpireAt(g *model.Grant) int64 {
	if !g.ExpireAt.Valid {
		return 0
	}
	return g.ExpireAt.Time.Unix()
}
