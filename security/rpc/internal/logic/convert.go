package logic

import (
	"database/sql"
	"time"

	"datacenter/security/model"
	"datacenter/security/rpc/security"
)

// levelToPb 分级行转 pb
func levelToPb(l *model.SecLevel) *security.SecLevel {
	return &security.SecLevel{
		Id:      l.Id,
		Code:    l.Code,
		Name:    l.Name,
		Comment: l.Comment,
		Beid:    l.Beid,
	}
}

// tagToPb 打标行转 pb
func tagToPb(t *model.EntityTag) *security.EntityTag {
	return &security.EntityTag{
		Id:           t.Id,
		EntityId:     t.EntityId,
		EntityName:   t.EntityName,
		LevelCode:    t.LevelCode,
		MaskRuleCode: t.MaskRuleCode,
		Beid:         t.Beid,
	}
}

// ruleToPb 脱敏规则行转 pb
func ruleToPb(r *model.MaskRule) *security.MaskRule {
	return &security.MaskRule{
		Id:         r.Id,
		RuleCode:   r.RuleCode,
		RuleName:   r.RuleName,
		Algorithm:  r.Algorithm,
		ParamsJson: r.ParamsJson,
		Beid:       r.Beid,
	}
}

// grantToPb 授权行转 pb（expire_at NULL → 0 表示永久）
func grantToPb(g *model.Grant) *security.SecGrant {
	pb := &security.SecGrant{
		Id:         g.Id,
		Subject:    g.Subject,
		Resource:   g.Resource,
		ResourceId: g.ResourceId,
		LevelCode:  g.LevelCode,
		Beid:       g.Beid,
	}
	if g.ExpireAt.Valid {
		pb.ExpireAt = g.ExpireAt.Time.Unix()
	}
	return pb
}

// expireAtToNullTime 秒级时间戳转 NullTime：0 → NULL（永久）
func expireAtToNullTime(expireAt int64) sql.NullTime {
	if expireAt <= 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: time.Unix(expireAt, 0), Valid: true}
}
