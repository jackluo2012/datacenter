package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/ide/model"
	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveScriptLogic {
	return &SaveScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存脚本（新建或更新，更新自动 version+1）：
// name 必填；id>0 时先校验脚本存在且归属当前租户（FindOne 已带多租户过滤），
// 再交由 model.Save 走更新链路（version 在库上 +1 并回读新版本号）
func (l *SaveScriptLogic) SaveScript(in *ide.SaveScriptReq) (*ide.ScriptReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrNameRequired
	}
	if in.Id > 0 {
		//更新：存在性 + beid 归属校验，不存在/越权统一按脚本不存在处理
		if _, err := l.svcCtx.ScriptModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrScriptNotFound
			}
			return nil, err
		}
	}
	saved, err := l.svcCtx.ScriptModel.Save(l.ctx, &model.Script{
		Id:           in.Id,
		Name:         strings.TrimSpace(in.Name),
		SqlText:      in.SqlText,
		DatasourceId: in.DatasourceId,
		DbName:       in.DbName,
		Beid:         in.Beid,
	})
	if err != nil {
		return nil, err
	}
	l.Infof("[SaveScript] id=%d version=%d beid=%d", saved.Id, saved.Version, in.Beid)
	return &ide.ScriptReply{Script: scriptToPb(saved)}, nil
}
