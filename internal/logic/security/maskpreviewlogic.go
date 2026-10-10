// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MaskPreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMaskPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MaskPreviewLogic {
	return &MaskPreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MaskPreviewLogic) MaskPreview(req *types.MaskPreviewReq) (resp *types.MaskPreviewResp, err error) {
	// todo: add your logic here and delete this line

	return
}
