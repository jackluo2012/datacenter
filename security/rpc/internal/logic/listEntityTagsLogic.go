package logic

import (
	"context"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEntityTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEntityTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEntityTagsLogic {
	return &ListEntityTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 打标列表（分页，level_code 非空时按密级过滤）
func (l *ListEntityTagsLogic) ListEntityTags(in *security.EntityTagListReq) (*security.EntityTagListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.EntityTagModel.List(l.ctx, in.Beid, in.LevelCode, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*security.EntityTag, 0, len(list))
	for _, tag := range list {
		pbList = append(pbList, tagToPb(tag))
	}
	return &security.EntityTagListReply{List: pbList, Total: total}, nil
}
