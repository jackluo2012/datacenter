package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVersionLogic {
	return &CreateVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 定版快照
func (l *CreateVersionLogic) CreateVersion(in *metadata.IdReq) (*metadata.VersionReply, error) {
	if in.Id <= 0 {
		return nil, ErrEntityNotFound
	}
	// 1. 加载实体当前状态
	entity, err := l.svcCtx.EntityModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}

	// 2. 版本号 = 当前最大版本 + 1（实体内自增，uk_entity_version 保证唯一）
	maxVersion, err := l.svcCtx.VersionModel.MaxVersion(l.ctx, entity.Id)
	if err != nil {
		return nil, err
	}

	// 3. 快照：实体当前字段 + props 序列化为 JSON
	snapshot, err := json.Marshal(newEntitySnapshot(entity))
	if err != nil {
		return nil, err
	}
	v := &model.Version{
		EntityId:     entity.Id,
		Version:      maxVersion + 1,
		SnapshotJson: string(snapshot),
		PublishedAt:  time.Now(),
	}
	id, err := l.svcCtx.VersionModel.Insert(l.ctx, v)
	if err != nil {
		return nil, err
	}
	v.Id = id

	l.Infof("[CreateVersion] entity_id=%d version=%d", v.EntityId, v.Version)
	return &metadata.VersionReply{Version: versionToPb(v)}, nil
}
