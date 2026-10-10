// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SecCheckPermLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecCheckPermLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecCheckPermLogic {
	return &SecCheckPermLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SecCheckPermLogic) SecCheckPerm(req *types.SecCheckPermReq) (resp *types.SecCheckPermResp, err error) {
	// todo: add your logic here and delete this line

	return
}
