package logic

import (
	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
)

// apiToPb model.Api → pb Api
func apiToPb(a *model.Api) *datasvc.Api {
	if a == nil {
		return nil
	}
	return &datasvc.Api{
		Id:             a.Id,
		Name:           a.Name,
		Path:           a.Path,
		Mode:           a.Mode,
		DatasourceId:   a.DatasourceId,
		SqlText:        a.SqlText,
		ParamsJson:     a.ParamsJson,
		CacheTtl:       a.CacheTtl,
		Status:         a.Status,
		RespSampleJson: a.RespSampleJson,
		Owner:          a.Owner,
		Beid:           a.Beid,
	}
}

// appToPb model.ApiApp → pb ApiApp（不回传加密密钥）
func appToPb(a *model.ApiApp) *datasvc.ApiApp {
	if a == nil {
		return nil
	}
	return &datasvc.ApiApp{
		Id:           a.Id,
		AppKey:       a.AppKey,
		Name:         a.Name,
		RateLimitQps: a.RateLimitQps,
		IpWhitelist:  a.IpWhitelist,
		Beid:         a.Beid,
	}
}
