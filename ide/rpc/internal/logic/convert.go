package logic

import (
	"datacenter/ide/model"
	"datacenter/ide/rpc/ide"
)

// scriptToPb 脚本 model -> pb（last_run_at 为毫秒时间戳，未运行为 0）
func scriptToPb(s *model.Script) *ide.Script {
	return &ide.Script{
		Id:           s.Id,
		Name:         s.Name,
		SqlText:      s.SqlText,
		Version:      s.Version,
		DatasourceId: s.DatasourceId,
		DbName:       s.DbName,
		LastRunAt:    s.LastRunAt,
		Beid:         s.Beid,
	}
}
