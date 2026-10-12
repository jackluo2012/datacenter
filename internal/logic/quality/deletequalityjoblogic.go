// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteQualityJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除质量任务
func NewDeleteQualityJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteQualityJobLogic {
	return &DeleteQualityJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteQualityJob: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteJob 虽返回 IdReq, 网关无返回体, 仅透传错误
// （存在性校验在 rpc 侧，结果历史不级联清理）
func (l *DeleteQualityJobLogic) DeleteQualityJob(req *types.QualityIdReq) error {
	_, err := l.svcCtx.QualityRpc.DeleteJob(l.ctx, &qualityClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}
