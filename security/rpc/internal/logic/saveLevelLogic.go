package logic

import (
	"context"
	"strings"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLevelLogic {
	return &SaveLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存分级目录（code 白名单 ^L[1-9]$；code 为自然键 upsert：
// uk_beid_code 冲突时覆盖 name/comment，入参 id 不参与定位仅作参考）
func (l *SaveLevelLogic) SaveLevel(in *security.SaveLevelReq) (*security.LevelReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	code := strings.TrimSpace(in.Code)
	if !ValidLevelCode(code) {
		return nil, ErrLevelCodeInvalid
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrLevelNameRequired
	}

	id, err := l.svcCtx.LevelModel.Save(l.ctx, &model.SecLevel{
		Code:    code,
		Name:    name,
		Comment: in.Comment,
		Beid:    in.Beid,
	})
	if err != nil {
		return nil, err
	}

	// 回读保证返回与库内一致
	level, err := l.svcCtx.LevelModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &security.LevelReply{Level: levelToPb(level)}, nil
}
