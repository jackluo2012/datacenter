package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompareVersionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompareVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompareVersionsLogic {
	return &CompareVersionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 版本对比
func (l *CompareVersionsLogic) CompareVersions(in *metadata.CompareReq) (*metadata.DiffReply, error) {
	// 1. 加载两版快照（version_a 为 before，version_b 为 after）
	va, err := l.svcCtx.VersionModel.FindOne(l.ctx, in.VersionA)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}
	vb, err := l.svcCtx.VersionModel.FindOne(l.ctx, in.VersionB)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	// 2. 解析快照 JSON
	sa, err := parseEntitySnapshot(va.SnapshotJson)
	if err != nil {
		return nil, fmt.Errorf("版本 %d 快照解析失败: %w", va.Id, err)
	}
	sb, err := parseEntitySnapshot(vb.SnapshotJson)
	if err != nil {
		return nil, fmt.Errorf("版本 %d 快照解析失败: %w", vb.Id, err)
	}

	// 3. 逐字段对比（name/comment/engine/rows_cnt/props_json/classification，纯函数见 snapshot.go）
	diffs := diffEntitySnapshot(sa, sb)
	data, err := json.Marshal(diffs)
	if err != nil {
		return nil, err
	}
	return &metadata.DiffReply{DiffJson: string(data)}, nil
}
