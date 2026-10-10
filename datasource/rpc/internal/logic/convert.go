package logic

import (
	"encoding/base64"
	"time"

	"datacenter/datasource/model"
	"datacenter/datasource/rpc/datasource"
)

// toPbDataSource 表模型转 pb 结构
// password_enc 为密文, 以 base64 传输; 明文密码任何情况下不回传
func toPbDataSource(m model.DataSource) *datasource.DataSourceInfo {
	return &datasource.DataSourceInfo{
		Id:          m.Id,
		Name:        m.Name,
		Type:        m.Type,
		Host:        m.Host,
		Port:        m.Port,
		DbName:      m.DbName,
		Username:    m.Username,
		PasswordEnc: base64.StdEncoding.EncodeToString(m.PasswordEnc),
		ParamsJson:  m.ParamsJson.String,
		Status:      m.Status,
		Owner:       m.Owner,
		Beid:        m.Beid,
		CreatedAt:   unixSeconds(m.CreatedAt),
		UpdatedAt:   unixSeconds(m.UpdatedAt),
	}
}

// unixSeconds 时间转秒级时间戳, 零值返回 0
func unixSeconds(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}
