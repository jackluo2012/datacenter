package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveWordRootLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveWordRootLogic {
	return &SaveWordRootLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存词根（id>0 更新，id=0 新建；name 同租户内查重）
func (l *SaveWordRootLogic) SaveWordRoot(in *standard.SaveWordRootReq) (*standard.WordRootReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrWordRootNameRequired
	}
	// 类型缺省 common，提供时须在白名单内
	rootType := strings.TrimSpace(in.Type)
	if rootType == "" {
		rootType = wordRootTypeDefault
	}
	if !IsValidWordRootType(rootType) {
		return nil, ErrWordRootTypeInvalid
	}

	data := &model.WordRoot{
		Name:    name,
		CnName:  strings.TrimSpace(in.CnName),
		Type:    rootType,
		Comment: in.Comment,
		Beid:    in.Beid,
	}

	if in.Id > 0 {
		// 更新：先确认存在（含租户归属），再对 name 查重（排除自身）
		old, err := l.svcCtx.WordRootModel.FindOne(l.ctx, in.Id, in.Beid)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrWordRootNotFound
			}
			return nil, err
		}
		if dup, derr := l.svcCtx.WordRootModel.FindOneByName(l.ctx, in.Beid, name); derr == nil && dup.Id != old.Id {
			return nil, ErrWordRootDuplicate
		} else if derr != nil && !errors.Is(derr, model.ErrNotFound) {
			return nil, derr
		}
		data.Id = old.Id
		if err := l.svcCtx.WordRootModel.Update(l.ctx, data); err != nil {
			return nil, err
		}
	} else {
		// 新建：name 查重后插入
		if _, err := l.svcCtx.WordRootModel.FindOneByName(l.ctx, in.Beid, name); err == nil {
			return nil, ErrWordRootDuplicate
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		id, err := l.svcCtx.WordRootModel.Insert(l.ctx, data)
		if err != nil {
			return nil, err
		}
		data.Id = id
	}

	// 回读保证返回与库内一致
	root, err := l.svcCtx.WordRootModel.FindOne(l.ctx, data.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &standard.WordRootReply{Root: rootToPb(root)}, nil
}
