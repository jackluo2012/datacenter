package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveThemeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveThemeLogic {
	return &SaveThemeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveTheme 新增/更新主题：所属域须存在；code 必填且须通过小写标识符正则，name 必填；
// 同租户下 code 查重（uk_beid_code）
func (l *SaveThemeLogic) SaveTheme(in *theme.SaveThemeReq) (*theme.ThemeReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	code := trimSpace(in.Code)
	name := trimSpace(in.Name)
	if in.DomainId <= 0 || code == "" || name == "" {
		return nil, ErrInvalidParams
	}
	if !ValidIdentifier(code) || len(code) > identifierMaxLen {
		return nil, ErrInvalidThemeCode
	}
	//域存在性校验（多租户过滤防跨租户挂靠）
	if _, err := l.svcCtx.DomainModel.FindOne(l.ctx, in.DomainId, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}

	//id=0 新建，id>0 更新
	if in.Id == 0 {
		//查重：同租户下 code 唯一
		if _, err := l.svcCtx.ThemeModel.FindByCode(l.ctx, in.Beid, code); err == nil {
			return nil, ErrDupThemeCode
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		row := &model.ThemeRow{DomainId: in.DomainId, Code: code, Name: name,
			Comment: trimSpace(in.Comment), Beid: in.Beid}
		id, err := l.svcCtx.ThemeModel.Insert(l.ctx, row)
		if err != nil {
			if errors.Is(err, model.ErrDupThemeCode) {
				return nil, ErrDupThemeCode
			}
			return nil, err
		}
		row.Id = id
		l.Infof("[SaveTheme] 新建主题 id=%d code=%s domain=%d beid=%d", id, code, in.DomainId, in.Beid)
		return &theme.ThemeReply{Theme: toPbTheme(row)}, nil
	}

	//更新：须存在
	exist, err := l.svcCtx.ThemeModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrThemeNotFound
		}
		return nil, err
	}
	//编码若变更，仍须查重
	if exist.Code != code {
		if dup, err := l.svcCtx.ThemeModel.FindByCode(l.ctx, in.Beid, code); err == nil && dup.Id != exist.Id {
			return nil, ErrDupThemeCode
		} else if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
	}
	row := &model.ThemeRow{Id: exist.Id, DomainId: in.DomainId, Code: code, Name: name,
		Comment: trimSpace(in.Comment), Beid: in.Beid}
	if err := l.svcCtx.ThemeModel.Update(l.ctx, row); err != nil {
		if errors.Is(err, model.ErrDupThemeCode) {
			return nil, ErrDupThemeCode
		}
		return nil, err
	}
	l.Infof("[SaveTheme] 更新主题 id=%d code=%s domain=%d beid=%d", row.Id, code, in.DomainId, in.Beid)
	return &theme.ThemeReply{Theme: toPbTheme(row)}, nil
}
