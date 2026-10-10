package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateApiLogic {
	return &CreateApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建API
func (l *CreateApiLogic) CreateApi(in *datasvc.CreateApiReq) (*datasvc.ApiReply, error) {
	//基础字段校验
	name := strings.TrimSpace(in.Name)
	path := strings.TrimSpace(in.Path)
	if name == "" {
		return nil, fmt.Errorf("%w: name 不能为空", ErrParamInvalid)
	}
	if path == "" || strings.ContainsAny(path, " \t\r\n") {
		return nil, fmt.Errorf("%w: path 非法，不能为空且不含空白字符", ErrParamInvalid)
	}
	if in.DatasourceId <= 0 {
		return nil, fmt.Errorf("%w: datasource_id 必须大于 0", ErrParamInvalid)
	}
	if in.CacheTtl < 0 {
		return nil, fmt.Errorf("%w: cache_ttl 不能为负数", ErrParamInvalid)
	}
	//模式：缺省 sql；wizard=向导生成（仅 SELECT）、sql=自由 SQL（允许写语句）、register=登记的外部 API
	mode := strings.TrimSpace(in.Mode)
	if mode == "" {
		mode = "sql"
	}
	switch mode {
	case "wizard", "sql", "register":
	default:
		return nil, fmt.Errorf("%w: mode 非法，仅支持 wizard/sql/register", ErrParamInvalid)
	}
	//params_json 合法性校验
	if _, err := ParseParamDefs(in.ParamsJson); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
	}
	//wizard 模式强制 sql_text 以 SELECT 开头（防写）；sql 模式不限
	if mode == "wizard" {
		if strings.TrimSpace(in.SqlText) == "" {
			return nil, fmt.Errorf("%w: wizard 模式 sql_text 不能为空", ErrParamInvalid)
		}
		if !IsSelectStmt(in.SqlText) {
			return nil, fmt.Errorf("%w: wizard 模式仅允许 SELECT 查询模板", ErrParamInvalid)
		}
	}
	//path 唯一性（beid+path）
	if _, err := l.svcCtx.ApiModel.FindByPath(in.Beid, path); err == nil {
		return nil, ErrPathDuplicate
	} else if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}
	data := &model.Api{
		Name:           name,
		Path:           path,
		Mode:           mode,
		DatasourceId:   in.DatasourceId,
		SqlText:        in.SqlText,
		ParamsJson:     strings.TrimSpace(in.ParamsJson),
		CacheTtl:       in.CacheTtl,
		Status:         "draft", //新建即为草稿，需显式发布
		RespSampleJson: in.RespSampleJson,
		Owner:          strings.TrimSpace(in.Owner),
		Beid:           in.Beid,
	}
	id, err := l.svcCtx.ApiModel.Insert(data)
	if err != nil {
		//兜底：并发创建触发 uk_beid_path 唯一键冲突
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, ErrPathDuplicate
		}
		return nil, err
	}
	data.Id = id
	l.Infof("创建API成功 id=%d beid=%d path=%s mode=%s", id, data.Beid, data.Path, data.Mode)
	return &datasvc.ApiReply{Api: apiToPb(data)}, nil
}
