package logic

import (
	"context"

	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type MaskPreviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMaskPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MaskPreviewLogic {
	return &MaskPreviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 脱敏预览
func (l *MaskPreviewLogic) MaskPreview(in *security.MaskPreviewReq) (*security.MaskPreviewReply, error) {
	// todo: add your logic here and delete this line

	return &security.MaskPreviewReply{}, nil
}
