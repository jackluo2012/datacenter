package logic

import (
	"sort"
	"strings"
	"unicode"
)

// stopWords 问数停用词表（MVP 硬编码）：问句中的疑问词、量词、泛义动词等无检索价值词，
// 匹配时按字节长度降序（最长优先），避免 "有多少" 被 "有" 提前截断
var stopWords = []string{
	// 疑问/量词词组
	"有哪些", "有多少", "是什么", "怎么样", "有没有", "多少", "什么", "哪些", "几个", "怎么", "如何",
	// 泛义动作/统计词
	"查询", "统计", "汇总", "列出", "显示", "查看", "看看", "一下", "分别", "情况", "数量",
	// 单字虚词/代词/介词
	"的", "了", "吗", "呢", "吧", "是", "有", "在", "和", "与", "或", "及",
	"按", "对", "从", "都", "各", "中", "请", "帮", "我", "想", "要", "个",
}

// sortedStopWords 停用词按字节长度降序副本（包初始化时排好，避免每次请求重排）
var sortedStopWords = func() []string {
	s := append([]string(nil), stopWords...)
	sort.Slice(s, func(i, j int) bool { return len(s[i]) > len(s[j]) })
	return s
}()

const (
	maxKeywords  = 8 //关键词数上限（控制下游语义检索 RPC 调用次数）
	bigramMinLen = 4 //中文块达到该长度才降级切 bigram
	minChunkLen  = 2 //中文块最短保留长度（单字块噪音大，直接过滤）
)

// tokenKind 词元类型
type tokenKind int

const (
	tokCJK  tokenKind = iota //连续中文块
	tokWord                  //连续英文/数字词
)

// token 简易分词词元
type token struct {
	kind tokenKind
	text string
}

// ExtractKeywords 问句关键词提取（MVP 分词降级法，不依赖分词库）：
//  1. 去停用词：按硬编码停用词表（最长优先）剔除疑问词/泛义动词等；
//  2. 整句：清洗后剩余词元拼接为紧凑短语，作为首个关键词（精确优先）；
//  3. 分词：英文按连续字母数字切词，中文按连续字块切分，长度 >= 4 的中文块
//     再降级切为二元 bigram 提升模糊匹配命中率（单字中文块直接过滤）；
//  4. 去重保序、截断至上限；清洗后为空时兜底返回原问句。
func ExtractKeywords(question string) []string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return nil
	}

	//1. 去停用词（统一小写，英文停用词场景兼容）
	cleaned := removeStopWords(strings.ToLower(trimmed))
	tokens := splitTokens(cleaned)
	if len(tokens) == 0 {
		//整句皆为停用词：兜底返回原问句
		return []string{trimmed}
	}

	//2. 整句短语：词元直接拼接（中英混排不补空格，贴近表名字段命名习惯）
	var phrase strings.Builder
	for _, tk := range tokens {
		phrase.WriteString(tk.text)
	}

	//3. 分词：英文词原样保留，中文块按长度降级
	parts := make([]string, 0, len(tokens)+2)
	parts = append(parts, phrase.String())
	for _, tk := range tokens {
		if tk.kind == tokWord {
			parts = append(parts, tk.text)
			continue
		}
		parts = append(parts, expandChunk(tk.text)...)
	}

	//4. 去重保序 + 截断
	return dedupeKeepOrder(parts, maxKeywords)
}

// removeStopWords 按停用词表（最长优先）剔除停用词，命中处替换为空格分隔
func removeStopWords(text string) string {
	for _, sw := range sortedStopWords {
		if strings.Contains(text, sw) {
			text = strings.ReplaceAll(text, sw, " ")
		}
	}
	return strings.Join(strings.Fields(text), " ")
}

// splitTokens 简易切词：连续中文为一个块、连续字母数字为一个词，其余字符视作分隔符
func splitTokens(text string) []token {
	var toks []token
	var cur []rune
	kind := tokCJK
	flush := func() {
		if len(cur) > 0 {
			toks = append(toks, token{kind: kind, text: string(cur)})
			cur = cur[:0]
		}
	}
	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r): //中文块（先于 IsLetter 判断，Han 亦是 Letter）
			if kind != tokCJK {
				flush()
				kind = tokCJK
			}
			cur = append(cur, r)
		case unicode.IsLetter(r) || unicode.IsDigit(r): //英文/数字词
			if kind != tokWord {
				flush()
				kind = tokWord
			}
			cur = append(cur, r)
		default: //标点/空白等分隔符
			flush()
		}
	}
	flush()
	return toks
}

// expandChunk 中文块降级切分：>= 4 字切 bigram（提升模糊命中率），2~3 字整块保留，单字过滤
func expandChunk(s string) []string {
	runes := []rune(s)
	switch {
	case len(runes) >= bigramMinLen:
		out := make([]string, 0, len(runes)-1)
		for i := 0; i+1 < len(runes); i++ {
			out = append(out, string(runes[i:i+2]))
		}
		return out
	case len(runes) >= minChunkLen:
		return []string{s}
	default:
		return nil
	}
}

// dedupeKeepOrder 去重保序并截断至上限
func dedupeKeepOrder(in []string, limit int) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
		if len(out) >= limit {
			break
		}
	}
	return out
}
