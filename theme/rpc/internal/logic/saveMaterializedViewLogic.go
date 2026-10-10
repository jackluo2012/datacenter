package logic

import (
	"context"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMaterializedViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMaterializedViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMaterializedViewLogic {
	return &SaveMaterializedViewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveMaterializedView 新增物化视图定义：name/view_name 必填，view_name 须通过小写
// 标识符正则（刷新时按此真实建视图，防注入）；select_sql 必须 SELECT 开头且不含分号
// （执行时与 create view 拼接）；refresh_mode 白名单（空缺省 manual）。
// 注：SaveMvReq 无 id 字段，本接口仅建档不更新（改定义走删除重建或后续扩展）
func (l *SaveMaterializedViewLogic) SaveMaterializedView(in *theme.SaveMvReq) (*theme.MvReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	name := trimSpace(in.Name)
	viewName := trimSpace(in.ViewName)
	if name == "" || viewName == "" || in.DatasourceId <= 0 {
		return nil, ErrInvalidParams
	}
	if !ValidIdentifier(viewName) || len(viewName) > identifierMaxLen {
		return nil, ErrInvalidViewName
	}
	//查询语句校验：SELECT 开头 + 无分号（防多语句注入）
	if err := ValidateSelectSql(in.SelectSql); err != nil {
		return nil, err
	}
	//刷新模式白名单（空缺省 manual）
	refreshMode := trimSpace(in.RefreshMode)
	if refreshMode == "" {
		refreshMode = RefreshModeManual
	}
	if !ValidRefreshMode(refreshMode) {
		return nil, ErrInvalidRefreshMode
	}
	//目标库名（可选）：非空须通过标识符校验（建视图时拼进 DDL，防注入）
	dbName := trimSpace(in.DbName)
	if dbName != "" && (!ValidIdentifier(dbName) || len(dbName) > identifierMaxLen) {
		return nil, ErrInvalidDbName
	}

	row := &model.MvRow{
		Name:         name,
		DatasourceId: in.DatasourceId,
		DbName:       dbName,
		ViewName:     viewName,
		SelectSql:    trimSpace(in.SelectSql),
		RefreshMode:  refreshMode,
		Beid:         in.Beid,
	}
	id, err := l.svcCtx.ThemeMvModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id = id
	l.Infof("[SaveMaterializedView] 新建物化视图 id=%d name=%s view=%s beid=%d", id, name, viewName, in.Beid)
	return &theme.MvReply{Mv: toPbMv(row)}, nil
}
