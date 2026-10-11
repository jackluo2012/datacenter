package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"datacenter/datasvc/model"
)

// 网关对外执行入口契约（见 datasvc.api：POST /datasvc/execute，按 body.path 路由）
const (
	executeHttpPath = "/datasvc/execute"
	// signDesc 签名口径说明（文档与 OpenAPI securitySchemes 共用，与 credential.go 的
	// genSign/verifySign 实现一致）
	signDesc = "sign = hex(hmac-sha256(app_key + ts, app_secret))；ts 为秒级时间戳，与服务器偏差不得超过 300 秒"
)

// BuildApiMarkdown 生成 Markdown 文档（纯函数，不触 DB/网络）：
// API 名/path/请求方式 POST/参数表[name,type,required,default]/签名说明（HMAC 示例代码）/
// 成功响应示例（resp_sample_json）/错误码表
func BuildApiMarkdown(api *model.Api) string {
	var b strings.Builder
	//标题与元信息
	b.WriteString("# " + docText(api.Name) + "\n\n")
	b.WriteString("> 执行路径：`" + api.Path + "` ｜ 请求方式：`POST` ｜ 状态：`" + api.Status + "`")
	if api.CacheTtl > 0 {
		b.WriteString(" ｜ 结果缓存：`" + strconv.FormatInt(api.CacheTtl, 10) + "s`")
	} else {
		b.WriteString(" ｜ 结果缓存：关闭")
	}
	if owner := strings.TrimSpace(api.Owner); owner != "" {
		b.WriteString(" ｜ 负责人：" + docText(owner))
	}
	b.WriteString("\n\n")

	//调用方式
	b.WriteString("## 调用方式\n\n")
	b.WriteString("统一网关入口 `POST " + executeHttpPath + "`，请求体携带 `app_key`、`ts`、`sign`、")
	b.WriteString("`path`（取本 API 执行路径 `" + api.Path + "`）与 `params_json`（业务参数 JSON 串）。\n\n")

	//参数表[name,type,required,default]
	b.WriteString("## 参数表\n\n")
	defs, _ := ParseParamDefs(api.ParamsJson)
	if len(defs) == 0 {
		b.WriteString("该 API 无业务参数，可直接调用。\n\n")
	} else {
		b.WriteString("| 参数 | 类型 | 必填 | 默认值 |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, d := range defs {
			b.WriteString("| " + docText(d.Name) + " | " + docText(d.Type) + " | " +
				requiredText(d.Required) + " | " + docText(formatParamDefault(d)) + " |\n")
		}
		b.WriteString("\n")
	}

	//签名说明（含 HMAC 示例代码）
	b.WriteString("## 签名说明\n\n")
	b.WriteString(signDesc + "\n\n")
	b.WriteString("Go 签名示例：\n\n")
	b.WriteString("```go\n")
	b.WriteString("ts := strconv.FormatInt(time.Now().Unix(), 10)\n")
	b.WriteString("mac := hmac.New(sha256.New, []byte(appSecret))\n")
	b.WriteString("mac.Write([]byte(appKey + ts))\n")
	b.WriteString("sign := hex.EncodeToString(mac.Sum(nil))\n")
	b.WriteString("```\n\n")
	b.WriteString("curl 调用示例：\n\n")
	b.WriteString("```bash\n")
	b.WriteString("curl -X POST http://{host}" + executeHttpPath + " \\\n")
	b.WriteString("  -H \"Content-Type: application/json\" \\\n")
	b.WriteString("  -d '{\"app_key\":\"your_app_key\",\"ts\":\"<秒级时间戳>\",\"sign\":\"<hex签名>\"," +
		"\"path\":\"" + api.Path + "\",\"params_json\":\"" + sampleParamsJsonLiteral(defs) + "\"}'\n")
	b.WriteString("```\n\n")

	//成功响应示例
	b.WriteString("## 成功响应示例\n\n")
	b.WriteString("```json\n")
	b.WriteString(sampleRespJson(api) + "\n")
	b.WriteString("```\n\n")

	//错误码表
	b.WriteString("## 错误码表\n\n")
	b.WriteString("| 错误码 | 状态 | 说明 |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, spec := range callErrorSpecs {
		b.WriteString("| " + spec.Code + " | " + spec.Status + " | " + spec.Desc + " |\n")
	}
	b.WriteString("| internal_error | " + model.CallStatusError + " | 服务内部错误 |\n")
	return b.String()
}

// ---- OpenAPI 3.0 片段结构（struct 字段顺序即 JSON 输出顺序，生成结果稳定可测） ----

type openApiDoc struct {
	Openapi    string                     `json:"openapi"`
	Info       openApiInfo                `json:"info"`
	Paths      map[string]openApiPathItem `json:"paths"`
	Components openApiComponents          `json:"components"`
}

type openApiInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type openApiPathItem struct {
	Post openApiOperation `json:"post"`
}

type openApiOperation struct {
	Summary     string                     `json:"summary"`
	OperationId string                     `json:"operationId"`
	Description string                     `json:"description"`
	Parameters  []openApiParameter         `json:"parameters"`
	RequestBody openApiRequestBody         `json:"requestBody"`
	Responses   map[string]openApiResponse `json:"responses"`
	Security    []map[string][]string      `json:"security"`
}

type openApiParameter struct {
	Name        string        `json:"name"`
	In          string        `json:"in"`
	Required    bool          `json:"required"`
	Description string        `json:"description"`
	Schema      openApiSchema `json:"schema"`
}

// openApiSchema schema 节点（与 openApiProperty 同构，object 型参数可嵌套 properties/required）
type openApiSchema struct {
	Type        string                     `json:"type,omitempty"`
	Description string                     `json:"description,omitempty"`
	Default     interface{}                `json:"default,omitempty"`
	Properties  map[string]openApiProperty `json:"properties,omitempty"`
	Required    []string                   `json:"required,omitempty"`
}

// openApiProperty 属性节点：与 schema 同构（别名），嵌套对象复用同一套字段
type openApiProperty = openApiSchema

type openApiRequestBody struct {
	Required bool                    `json:"required"`
	Content  map[string]openApiMedia `json:"content"`
}

type openApiMedia struct {
	Schema  openApiSchema `json:"schema"`
	Example interface{}   `json:"example,omitempty"`
}

type openApiResponse struct {
	Description string                  `json:"description"`
	Content     map[string]openApiMedia `json:"content"`
}

type openApiComponents struct {
	SecuritySchemes map[string]openApiSecurityScheme `json:"securitySchemes"`
}

type openApiSecurityScheme struct {
	Type        string `json:"type"`
	In          string `json:"in"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// openApiExecuteExample 请求体示例（网关 body 契约：path + params 对象）
type openApiExecuteExample struct {
	Path   string                 `json:"path"`
	Params map[string]interface{} `json:"params"`
}

// openApiExecuteRespExample 响应体示例（网关 ExecuteApiResp 契约：rows_json 是 JSON 字符串）
type openApiExecuteRespExample struct {
	RowsJson string `json:"rows_json"`
	CostMs   int64  `json:"cost_ms"`
	CacheHit bool   `json:"cache_hit"`
	Affected int64  `json:"affected"`
}

// BuildOpenApiJSON 生成 OpenAPI 3.0 JSON 片段（纯函数，缩进序列化、合法性由构造保证）：
// paths/{path}/post + parameters（鉴权头）+ requestBody（业务参数）+ securitySchemes（apiKey）
func BuildOpenApiJSON(api *model.Api) (string, error) {
	defs, err := ParseParamDefs(api.ParamsJson)
	if err != nil {
		return "", fmt.Errorf("params_json 非法，无法生成 OpenAPI: %v", err)
	}
	//paths key：执行路径统一挂 /dsvc 前缀（proto 契约：执行路径 /dsvc/{path}）
	docPath := "/dsvc/" + strings.TrimPrefix(strings.TrimSpace(api.Path), "/")

	//业务参数 schema：按 params_json 定义生成类型/必填/缺省
	props := make(map[string]openApiProperty, len(defs))
	required := make([]string, 0, len(defs))
	for _, d := range defs {
		p := openApiProperty{Type: openApiType(d.Type), Description: "必填: " + requiredText(d.Required)}
		if d.Default != nil {
			p.Default = d.Default
		}
		props[d.Name] = p
		if d.Required {
			required = append(required, d.Name)
		}
	}

	doc := openApiDoc{
		Openapi: "3.0.3",
		Info: openApiInfo{
			Title:       "数据服务 API：" + api.Name,
			Description: "由 datasvc 文档门户自动生成（OpenAPI 3.0 片段）",
			Version:     "1.0.0",
		},
		Paths: map[string]openApiPathItem{
			docPath: {
				Post: openApiOperation{
					Summary:     api.Name,
					OperationId: "execute_" + strings.NewReplacer("/", "_").Replace(api.Path),
					Description: "统一网关入口 POST " + executeHttpPath +
						"（app_key/ts/sign 亦可通过 body 字段等价传递，见文档门户 Markdown）",
					Parameters: []openApiParameter{
						{Name: "X-App-Key", In: "header", Required: true,
							Description: "应用 app_key", Schema: openApiSchema{Type: "string"}},
						{Name: "X-Timestamp", In: "header", Required: true,
							Description: "签名时间戳 ts（秒级）", Schema: openApiSchema{Type: "string"}},
						{Name: "X-Sign", In: "header", Required: true,
							Description: "hex(hmac-sha256(app_key+ts, app_secret))",
							Schema:      openApiSchema{Type: "string"}},
					},
					RequestBody: openApiRequestBody{
						Required: true,
						Content: map[string]openApiMedia{
							"application/json": {
								Schema: openApiSchema{
									Type: "object",
									Properties: map[string]openApiProperty{
										"path": {Type: "string", Description: "API 执行路径", Default: api.Path},
										"params": {Type: "object", Description: "业务参数（见参数表）",
											Properties: props, Required: required},
									},
								},
								Example: openApiExecuteExample{Path: api.Path, Params: sampleParams(defs)},
							},
						},
					},
					Responses: map[string]openApiResponse{
						"200": {
							Description: "执行结果（rows_json 为查询结果 JSON 数组字符串）",
							Content: map[string]openApiMedia{
								"application/json": {
									Schema: openApiSchema{
										Type: "object",
										Properties: map[string]openApiProperty{
											"rows_json": {Type: "string", Description: "查询结果 JSON 数组字符串"},
											"cost_ms":   {Type: "integer", Description: "执行耗时（毫秒）"},
											"cache_hit": {Type: "boolean", Description: "是否命中结果缓存"},
											"affected":  {Type: "integer", Description: "写语句受影响行数"},
										},
									},
									Example: openApiExecuteRespExample{
										RowsJson: compactJsonString(sampleRespJson(api)),
										CostMs:   12,
										CacheHit: false,
										Affected: 0,
									},
								},
							},
						},
					},
					//apiKey 鉴权：app_key+ts+HMAC 签名（算法见 securitySchemes 描述）
					Security: []map[string][]string{{"AppKeyHmac": {}}},
				},
			},
		},
		Components: openApiComponents{
			SecuritySchemes: map[string]openApiSecurityScheme{
				"AppKeyHmac": {
					Type:        "apiKey",
					In:          "header",
					Name:        "X-App-Key",
					Description: "数据服务应用鉴权：" + signDesc,
				},
			},
		},
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ---- 生成器共用的小工具（均为纯函数） ----

// openApiType 参数类型 → OpenAPI schema 类型
func openApiType(typ string) string {
	switch typ {
	case TypeNumber:
		return "number"
	case TypeBool:
		return "boolean"
	default:
		return "string"
	}
}

// requiredText 必填 → 是/否（参数表与 OpenAPI 描述共用口径）
func requiredText(required bool) string {
	if required {
		return "是"
	}
	return "否"
}

// docText 文本转义：抹平竖线/换行，避免破坏 Markdown 表格结构
func docText(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

// formatParamDefault 参数默认值展示：未配置 → "-"；number/bool 用 JSON 字面量，string 原样
func formatParamDefault(d ParamDef) string {
	if d.Default == nil {
		return "-"
	}
	if d.Type == TypeString {
		return fmt.Sprintf("%v", d.Default)
	}
	if raw, err := json.Marshal(d.Default); err == nil {
		return string(raw)
	}
	return fmt.Sprintf("%v", d.Default)
}

// sampleParams 按参数定义生成最小请求参数示例：缺省优先 default，required 用类型占位值
func sampleParams(defs []ParamDef) map[string]interface{} {
	sample := make(map[string]interface{}, len(defs))
	for _, d := range defs {
		if d.Default != nil {
			sample[d.Name] = d.Default
			continue
		}
		if d.Required {
			switch d.Type {
			case TypeNumber:
				sample[d.Name] = 0
			case TypeBool:
				sample[d.Name] = false
			default:
				sample[d.Name] = "xxx"
			}
		}
	}
	return sample
}

// sampleParamsJsonLiteral curl 示例里 params_json 的内嵌字面量（JSON 对象字符串的转义形态）
func sampleParamsJsonLiteral(defs []ParamDef) string {
	raw, err := json.Marshal(sampleParams(defs))
	if err != nil {
		return "{}"
	}
	//curl -d 的 JSON 里再嵌一层 JSON 字符串：双引号需转义为 \"
	return strings.ReplaceAll(string(raw), `"`, `\"`)
}

// sampleRespJson 成功响应示例：优先 resp_sample_json（合法 JSON 时），否则通用占位数组
func sampleRespJson(api *model.Api) string {
	s := strings.TrimSpace(api.RespSampleJson)
	if s != "" && json.Valid([]byte(s)) {
		return s
	}
	return `[{"column1": "value1"}]`
}

// compactJsonString 压缩 JSON 文本（用于示例中的 rows_json 字符串值，去掉多余空白）
func compactJsonString(s string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s //理论不可达：入参已保证合法 JSON，兜底原样返回
	}
	return buf.String()
}
