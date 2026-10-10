package logic

import (
	"context"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLevelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLevelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelsLogic {
	return &ListLevelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分级目录列表；首次使用（目录为空）自动播种内置四级：L1 公开/L2 内部/L3 敏感/L4 机密
func (l *ListLevelsLogic) ListLevels(in *security.IdReq) (*security.LevelListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, err := l.svcCtx.LevelModel.List(l.ctx, in.Beid)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		// 种子初始化：仅目录为空时执行，insert ignore 幂等可重入
		if err := l.svcCtx.LevelModel.SeedDefaults(l.ctx, in.Beid); err != nil {
			return nil, err
		}
		if list, err = l.svcCtx.LevelModel.List(l.ctx, in.Beid); err != nil {
			return nil, err
		}
	}
	pbList := make([]*security.SecLevel, 0, len(list))
	for _, lv := range list {
		pbList = append(pbList, levelToPb(lv))
	}
	return &security.LevelListReply{List: pbList}, nil
}
