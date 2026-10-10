package logic

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"datacenter/ai/rpc/ai"
	"datacenter/metadata/rpc/metadataClient"
	"datacenter/metric/rpc/metricsvc"
)

// 打分权重：name 命中 > comment 命中 > detail(fqn/expr) 命中
const (
	scoreNameHit    = 3 //名称命中关键词
	scoreCommentHit = 2 //注释/口径命中关键词
	scoreDetailHit  = 1 //fqn/表达式命中关键词
)

// candidateJson 候选的稳定 JSON 形态（prompt 与审计 candidates_json 共用，避免 pb 结构序列化噪音）
type candidateJson struct {
	Type    string `json:"type"`
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Detail  string `json:"detail"`
	Comment string `json:"comment"`
}

// searchCandidates 语义检索：用提取的关键词分别调元数据实体与指标列表（keyword 模糊，size=max_candidates），
// 合并去重为统一候选；单个关键词或单个下游失败只记日志降级，不阻断问数主流程
func (l *AskLogic) searchCandidates(ctx context.Context, beid int64, keywords []string, size int) []*ai.Candidate {
	if beid <= 0 || len(keywords) == 0 {
		return nil
	}
	//多租户必填校验提前：下游两个列表接口均要求 beid>0
	var entities []*metadataClient.MetaEntity
	var ms []*metricsvc.Metric
	for _, kw := range keywords {
		//元数据实体：name/fqn 模糊匹配
		entResp, err := l.svcCtx.MetadataRpc.ListEntities(ctx, &metadataClient.EntityListReq{
			Beid:    beid,
			Keyword: kw,
			Page:    1,
			Size:    int64(size),
		})
		if err != nil {
			l.Errorf("语义检索元数据实体失败(降级跳过): keyword=%s err=%v", kw, err)
		} else {
			entities = append(entities, entResp.List...)
		}

		//指标：name_cn/name_en 模糊匹配
		metricResp, err := l.svcCtx.MetricRpc.ListMetrics(ctx, &metricsvc.MetricListReq{
			Beid:    beid,
			Keyword: kw,
			Page:    1,
			Size:    int64(size),
		})
		if err != nil {
			l.Errorf("语义检索指标失败(降级跳过): keyword=%s err=%v", kw, err)
			continue
		}
		ms = append(ms, metricResp.List...)
	}
	return buildCandidates(entities, ms)
}

// buildCandidates 实体与指标检索结果转换为统一候选（纯函数）：
// entity→detail=fqn、comment 取注释；metric→detail=agg_type(biz_field)（派生指标用 expr）、comment 取口径说明；
// 并按 type+id 去重保序（多关键词命中同一实体/指标时只保留首个）
func buildCandidates(entities []*metadataClient.MetaEntity, ms []*metricsvc.Metric) []*ai.Candidate {
	out := make([]*ai.Candidate, 0, len(entities)+len(ms))
	for _, e := range entities {
		out = append(out, &ai.Candidate{
			Type:    "entity",
			Id:      e.Id,
			Name:    e.Name,
			Detail:  e.Fqn,
			Comment: e.Comment,
		})
	}
	for _, m := range ms {
		//口径表达式：原子指标用 agg_type(biz_field)，两者缺其一则回退派生表达式 expr
		detail := m.Expr
		if m.AggType != "" && m.BizField != "" {
			detail = m.AggType + "(" + m.BizField + ")"
		}
		out = append(out, &ai.Candidate{
			Type:    "metric",
			Id:      m.Id,
			Name:    metricName(m),
			Detail:  detail,
			Comment: m.Comment,
		})
	}
	return dedupeCandidates(out)
}

// metricName 指标展示名：中文名优先，为空回退英文名
func metricName(m *metricsvc.Metric) string {
	if m.NameCn != "" {
		return m.NameCn
	}
	return m.NameEn
}

// dedupeCandidates 按 type+id 去重（多关键词命中同一实体/指标时只保留首个），保序
func dedupeCandidates(cands []*ai.Candidate) []*ai.Candidate {
	seen := make(map[string]struct{}, len(cands))
	out := make([]*ai.Candidate, 0, len(cands))
	for _, c := range cands {
		key := c.Type + ":" + strconv.FormatInt(c.Id, 10)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, c)
	}
	return out
}

// RankCandidates 候选打分排序（纯函数）：
// 对每个关键词，name 含关键词计 3 分 > comment 含计 2 分 > detail(fqn/expr) 含计 1 分，命中数累加计分；
// 按总分降序（同分按命中关键词数降序，再按原始顺序稳定排序），截取 top n
func RankCandidates(cands []*ai.Candidate, keywords []string, n int) []*ai.Candidate {
	if len(cands) == 0 || n <= 0 {
		return nil
	}
	type scored struct {
		cand  *ai.Candidate
		score int
		hits  int
	}
	list := make([]scored, 0, len(cands))
	for _, c := range cands {
		s, h := scoreCandidate(c, keywords)
		list = append(list, scored{cand: c, score: s, hits: h})
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		return list[i].hits > list[j].hits
	})
	if n > len(list) {
		n = len(list)
	}
	out := make([]*ai.Candidate, 0, n)
	for _, s := range list[:n] {
		out = append(out, s.cand)
	}
	return out
}

// scoreCandidate 单候选打分（纯函数）：返回（总分, 命中的关键词数），
// 比较统一转小写，兼容英文关键词与驼峰/大写字段名
func scoreCandidate(c *ai.Candidate, keywords []string) (score, hits int) {
	name := strings.ToLower(c.Name)
	comment := strings.ToLower(c.Comment)
	detail := strings.ToLower(c.Detail)
	for _, kw := range keywords {
		switch {
		case kw != "" && strings.Contains(name, kw):
			score += scoreNameHit
			hits++
		case kw != "" && strings.Contains(comment, kw):
			score += scoreCommentHit
			hits++
		case kw != "" && strings.Contains(detail, kw):
			score += scoreDetailHit
			hits++
		}
	}
	return score, hits
}
