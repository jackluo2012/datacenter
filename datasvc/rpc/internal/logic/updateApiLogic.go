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

type UpdateApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApiLogic {
	return &UpdateApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新API（空字符串字段不更新，保持原值；cache_ttl 以传入值为准，0 即关闭缓存）
func (l *UpdateApiLogic) UpdateApi(in *datasvc.UpdateApiReq) (*datasvc.OkReply, error) {
	if in.Id <= 0 {
		return nil, fmt.Errorf("%w: id 非法", ErrParamInvalid)
	}
	if in.CacheTtl < 0 {
		return nil, fmt.Errorf("%w: cache_ttl 不能为负数", ErrParamInvalid)
	}
	api, err := l.svcCtx.ApiModel.FindOne(in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	//多租户校验
	if in.Beid != 0 && api.Beid != in.Beid {
		return nil, ErrApiNotFound
	}
	//params_json 合法性校验
	if strings.TrimSpace(in.ParamsJson) != "" {
		if _, err := ParseParamDefs(in.ParamsJson); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
		}
		api.ParamsJson = strings.TrimSpace(in.ParamsJson)
	}
	//sql_text：空串保持原值；wizard 模式仍强制 SELECT 开头
	if strings.TrimSpace(in.SqlText) != "" {
		if api.Mode == "wizard" && !IsSelectStmt(in.SqlText) {
			return nil, fmt.Errorf("%w: wizard 模式仅允许 SELECT 查询模板", ErrParamInvalid)
		}
		api.SqlText = in.SqlText
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		api.Name = name
	}
	if sample := strings.TrimSpace(in.RespSampleJson); sample != "" {
		api.RespSampleJson = in.RespSampleJson
	}
	api.CacheTtl = in.CacheTtl
	if err := l.svcCtx.ApiModel.Update(api); err != nil {
		return nil, err
	}
	l.Infof("更新API成功 id=%d beid=%d", api.Id, api.Beid)
	return &datasvc.OkReply{Ok: true}, nil
}
