package logic

import (
	"context"
	"errors"
	"strings"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SaveDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveDictLogic {
	return &SaveDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存字典（整存整取：事务内 dict upsert + 字典项全删全插）。
// id>0 按 id 更新字典；id=0 按 uk_beid_code upsert（编码已存在等价于整体覆盖）。
// items 每次全量覆盖，以请求中的列表为准
func (l *SaveDictLogic) SaveDict(in *standard.SaveDictReq) (*standard.DictReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	dictCode := strings.TrimSpace(in.DictCode)
	if dictCode == "" {
		return nil, ErrDictCodeRequired
	}
	dictName := strings.TrimSpace(in.DictName)
	if dictName == "" {
		return nil, ErrDictNameRequired
	}
	// 字典项静态校验：编码/含义必填且同字典内编码不重复
	items := make([]*model.DictItem, 0, len(in.Items))
	if err := validateDictItems(in.Items); err != nil {
		return nil, err
	}
	for _, item := range in.Items {
		items = append(items, &model.DictItem{
			ItemCode:  strings.TrimSpace(item.ItemCode),
			ItemValue: item.ItemValue,
			Sort:      int64(item.Sort),
			Beid:      in.Beid,
		})
	}

	// id>0 先确认字典存在（含租户归属）
	if in.Id > 0 {
		if _, err := l.svcCtx.DictModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrDictNotFound
			}
			return nil, err
		}
	}

	var dictId int64
	// 事务保证"字典 + 全部字典项"原子落库，失败整体回滚
	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if in.Id > 0 {
			dictId = in.Id
			if err := l.svcCtx.DictModel.Update(txCtx, session, &model.Dict{
				Id:       in.Id,
				DictCode: dictCode,
				DictName: dictName,
				Comment:  in.Comment,
				Beid:     in.Beid,
			}); err != nil {
				return err
			}
		} else {
			// upsert：uk_beid_code 冲突时覆盖基础字段，等价整体保存
			if _, err := l.svcCtx.DictModel.InsertOnDuplicate(txCtx, session, &model.Dict{
				DictCode: dictCode,
				DictName: dictName,
				Comment:  in.Comment,
				Status:   model.DictStatusEnabled,
				Beid:     in.Beid,
			}); err != nil {
				l.Logger.Errorf("[定位] InsertOnDuplicate 失败: %v", err)
				return err
			}
			// on duplicate 走更新分支时 LastInsertId 不可靠，回查主键
			dict, err := l.svcCtx.DictModel.FindOneByCode(txCtx, session, in.Beid, dictCode)
			if err != nil {
				l.Logger.Errorf("[定位] 事务内回查 FindOneByCode 失败: %v", err)
				return err
			}
			dictId = dict.Id
		}

		// 字典项全删全插（整存整取的"取"半边）
		if err := l.svcCtx.DictItemModel.DeleteByDictId(txCtx, session, dictId); err != nil {
			l.Logger.Errorf("[定位] 删旧 items 失败 dictId=%d: %v", dictId, err)
			return err
		}
		for _, item := range items {
			item.DictId = dictId
		}
		if err := l.svcCtx.DictItemModel.BatchInsert(txCtx, session, items); err != nil {
			l.Logger.Errorf("[定位] 批量插 items 失败: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 回读保证返回与库内一致
	dict, err := l.svcCtx.DictModel.FindOne(l.ctx, dictId, in.Beid)
	if err != nil {
		return nil, err
	}
	savedItems, err := l.svcCtx.DictItemModel.FindByDictId(l.ctx, dictId)
	if err != nil {
		return nil, err
	}
	pbItems := make([]*standard.DictItem, 0, len(savedItems))
	for _, item := range savedItems {
		pbItems = append(pbItems, itemToPb(item))
	}
	return &standard.DictReply{Dict: dictToPb(dict), Items: pbItems}, nil
}

// validateDictItems 字典项静态校验（不查库）：编码/含义必填且编码不重复
func validateDictItems(items []*standard.DictItem) error {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.ItemCode)
		if code == "" || strings.TrimSpace(item.ItemValue) == "" {
			return ErrDictItemInvalid
		}
		if _, dup := seen[code]; dup {
			return ErrDictItemInvalid
		}
		seen[code] = struct{}{}
	}
	return nil
}
