package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveDomainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDomainLogic {
	return &SaveDomainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveDomain 新增/更新主题域：code 必填且须通过小写标识符正则（会作为业务主键风格
// 的编码被引用，防注入），name 必填；同租户下 code 查重（uk_beid_code）
func (l *SaveDomainLogic) SaveDomain(in *theme.SaveDomainReq) (*theme.DomainReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	code := trimSpace(in.Code)
	name := trimSpace(in.Name)
	if code == "" || name == "" {
		return nil, ErrInvalidParams
	}
	if !ValidIdentifier(code) || len(code) > identifierMaxLen {
		return nil, ErrInvalidDomainCode
	}

	//id=0 新建，id>0 更新
	if in.Id == 0 {
		//查重：同租户下 code 唯一
		if _, err := l.svcCtx.DomainModel.FindByCode(l.ctx, in.Beid, code); err == nil {
			return nil, ErrDupDomainCode
		} else if !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		row := &model.DomainRow{Code: code, Name: name, Comment: trimSpace(in.Comment), Beid: in.Beid}
		id, err := l.svcCtx.DomainModel.Insert(l.ctx, row)
		if err != nil {
			if errors.Is(err, model.ErrDupDomainCode) {
				return nil, ErrDupDomainCode
			}
			return nil, err
		}
		row.Id = id
		l.Infof("[SaveDomain] 新建主题域 id=%d code=%s beid=%d", id, code, in.Beid)
		return &theme.DomainReply{Domain: toPbDomain(row)}, nil
	}

	//更新：须存在
	exist, err := l.svcCtx.DomainModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	//编码若变更，仍须查重
	if exist.Code != code {
		if dup, err := l.svcCtx.DomainModel.FindByCode(l.ctx, in.Beid, code); err == nil && dup.Id != exist.Id {
			return nil, ErrDupDomainCode
		} else if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
	}
	row := &model.DomainRow{Id: exist.Id, Code: code, Name: name, Comment: trimSpace(in.Comment), Beid: in.Beid}
	if err := l.svcCtx.DomainModel.Update(l.ctx, row); err != nil {
		if errors.Is(err, model.ErrDupDomainCode) {
			return nil, ErrDupDomainCode
		}
		return nil, err
	}
	l.Infof("[SaveDomain] 更新主题域 id=%d code=%s beid=%d", row.Id, code, in.Beid)
	return &theme.DomainReply{Domain: toPbDomain(row)}, nil
}
