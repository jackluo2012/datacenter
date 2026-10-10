package logic

import (
	"encoding/json"
	"strconv"

	"datacenter/metadata/model"
)

// entitySnapshot 实体定版快照结构（dc_meta_version.snapshot_json）
type entitySnapshot struct {
	EntityType     string `json:"entity_type"`
	Fqn            string `json:"fqn"`
	Name           string `json:"name"`
	Comment        string `json:"comment"`
	Engine         string `json:"engine"`
	RowsCnt        int64  `json:"rows_cnt"`
	PropsJson      string `json:"props_json"`
	Classification string `json:"classification"`
	Owner          string `json:"owner"`
	DsId           int64  `json:"ds_id"`
}

// newEntitySnapshot 从实体行构造定版快照（记录当前字段 + props）
func newEntitySnapshot(e *model.Entity) *entitySnapshot {
	return &entitySnapshot{
		EntityType:     e.EntityType,
		Fqn:            e.Fqn,
		Name:           e.Name,
		Comment:        e.Comment,
		Engine:         e.Engine,
		RowsCnt:        e.RowsCnt,
		PropsJson:      e.PropsJson,
		Classification: e.Classification,
		Owner:          e.Owner,
		DsId:           e.DsId,
	}
}

// parseEntitySnapshot 解析快照 JSON
func parseEntitySnapshot(data string) (*entitySnapshot, error) {
	var s entitySnapshot
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// FieldDiff 单字段差异
type FieldDiff struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// diffEntitySnapshot 逐字段对比两版快照，返回差异字段数组（before=旧版，after=新版）。
// 对比范围：name/comment/engine/rows_cnt/props_json/classification，字段顺序固定，便于前端渲染。
func diffEntitySnapshot(before, after *entitySnapshot) []FieldDiff {
	diffs := make([]FieldDiff, 0)
	add := func(field, beforeVal, afterVal string) {
		if beforeVal != afterVal {
			diffs = append(diffs, FieldDiff{Field: field, Before: beforeVal, After: afterVal})
		}
	}
	add("name", before.Name, after.Name)
	add("comment", before.Comment, after.Comment)
	add("engine", before.Engine, after.Engine)
	add("rows_cnt", strconv.FormatInt(before.RowsCnt, 10), strconv.FormatInt(after.RowsCnt, 10))
	add("props_json", before.PropsJson, after.PropsJson)
	add("classification", before.Classification, after.Classification)
	return diffs
}
