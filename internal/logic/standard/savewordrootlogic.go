// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/standard/rpc/standardClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveWordRootLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存词根
func NewSaveWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveWordRootLogic {
	return &SaveWordRootLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 保存词根: Type 缺省归一为 common, 逐字段映射为 rpc 请求后转发, 转发后把词根转为网关类型返回。
// 契约差异: 网关 SaveWordRootReq 无 Id 字段, 网关已透传 req.Id（0=新建,>0=更新）。
func (l *SaveWordRootLogic) SaveWordRoot(req *types.SaveWordRootReq) (resp *types.StdWordRoot, err error) {
	wordType := req.Type
	if wordType == "" {
		// 词根类型缺省为 common
		wordType = "common"
	}
	reply, err := l.svcCtx.StandardRpc.SaveWordRoot(l.ctx, &standardClient.SaveWordRootReq{
		Id:      req.Id,
		Name:    req.Name,
		CnName:  req.CnName,
		Type:    wordType,
		Comment: req.Comment,
		Beid:    req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return wordRootToTypes(reply.GetRoot()), nil
}
