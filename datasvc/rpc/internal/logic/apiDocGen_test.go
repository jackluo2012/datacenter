package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"datacenter/datasvc/model"
)

// docSampleApi 构造一个参数/响应样例齐全的 API
func docSampleApi() *model.Api {
	return &model.Api{
		Id:       3,
		Name:     "用户查询",
		Path:     "user/query",
		Mode:     "sql",
		Status:   "published",
		CacheTtl: 60,
		ParamsJson: `[{"name":"id","type":"number","required":true},` +
			`{"name":"name","type":"string","required":false,"default":"tom"},` +
			`{"name":"vip","type":"bool","required":false}]`,
		RespSampleJson: `[{"id":1,"name":"tom","phone":"13800000000"}]`,
		Owner:          "ops",
		Beid:           1,
	}
}

// OpenAPI JSON 生成：合法性（可反序列化）+ 结构断言（paths/post/parameters/securitySchemes/requestBody）
func TestBuildOpenApiJSON(t *testing.T) {
	api := docSampleApi()
	out, err := BuildOpenApiJSON(api)
	if err != nil {
		t.Fatalf("生成 OpenAPI 失败: %v", err)
	}
	//json 合法性
	if !json.Valid([]byte(out)) {
		t.Fatalf("生成的 OpenAPI 不是合法 JSON")
	}
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("OpenAPI 反序列化失败: %v", err)
	}
	//根结构：openapi 版本 / paths / components
	if v, ok := doc["openapi"].(string); !ok || !strings.HasPrefix(v, "3.0") {
		t.Fatalf("openapi 版本应为 3.0.x, got %v", doc["openapi"])
	}
	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("paths 应为对象, got %T", doc["paths"])
	}
	//paths/{path}/post
	item, ok := paths["/dsvc/user/query"].(map[string]interface{})
	if !ok {
		t.Fatalf("paths 应包含 /dsvc/user/query, got %v", paths)
	}
	post, ok := item["post"].(map[string]interface{})
	if !ok {
		t.Fatalf("path item 应含 post 操作")
	}
	if post["summary"] != "用户查询" || post["operationId"] != "execute_user_query" {
		t.Fatalf("summary/operationId 不符: %v %v", post["summary"], post["operationId"])
	}
	//parameters：三个鉴权头
	params, ok := post["parameters"].([]interface{})
	if !ok || len(params) != 3 {
		t.Fatalf("parameters 应为 3 个鉴权头, got %v", post["parameters"])
	}
	first, _ := params[0].(map[string]interface{})
	if first["in"] != "header" || first["required"] != true {
		t.Fatalf("鉴权参数应为 required header, got %v", first)
	}
	//requestBody：业务参数 id 为 required 的 number
	body := post["requestBody"].(map[string]interface{})
	content := body["content"].(map[string]interface{})["application/json"].(map[string]interface{})
	schema := content["schema"].(map[string]interface{})
	props := schema["properties"].(map[string]interface{})
	paramsObj := props["params"].(map[string]interface{})
	if paramsObj["required"].([]interface{})[0] != "id" {
		t.Fatalf("params.required 应含 id, got %v", paramsObj["required"])
	}
	idProp := paramsObj["properties"].(map[string]interface{})["id"].(map[string]interface{})
	if idProp["type"] != "number" {
		t.Fatalf("参数 id 的 schema 类型应为 number, got %v", idProp["type"])
	}
	vipProp := paramsObj["properties"].(map[string]interface{})["vip"].(map[string]interface{})
	if vipProp["type"] != "boolean" {
		t.Fatalf("参数 vip 的 schema 类型应为 boolean, got %v", vipProp["type"])
	}
	nameProp := paramsObj["properties"].(map[string]interface{})["name"].(map[string]interface{})
	if nameProp["default"] != "tom" {
		t.Fatalf("参数 name 的 default 应为 tom, got %v", nameProp["default"])
	}
	//请求示例必须携带 path 与 params
	example := content["example"].(map[string]interface{})
	if example["path"] != "user/query" {
		t.Fatalf("示例 body.path 不符: %v", example["path"])
	}
	if _, ok := example["params"].(map[string]interface{}); !ok {
		t.Fatalf("示例 body.params 应为对象: %v", example["params"])
	}
	//responses：200 含 rows_json 示例（透传 resp_sample_json）
	resp200 := post["responses"].(map[string]interface{})["200"].(map[string]interface{})
	respContent := resp200["content"].(map[string]interface{})["application/json"].(map[string]interface{})
	respExample := respContent["example"].(map[string]interface{})
	if !strings.Contains(respExample["rows_json"].(string), "13800000000") {
		t.Fatalf("响应示例 rows_json 应含 resp_sample_json 内容: %v", respExample["rows_json"])
	}
	//security + securitySchemes（apiKey）
	security := post["security"].([]interface{})
	if len(security) != 1 {
		t.Fatalf("security 应有且仅有一项")
	}
	components := doc["components"].(map[string]interface{})
	schemes := components["securitySchemes"].(map[string]interface{})
	scheme := schemes["AppKeyHmac"].(map[string]interface{})
	if scheme["type"] != "apiKey" || scheme["in"] != "header" || scheme["name"] != "X-App-Key" {
		t.Fatalf("securitySchemes.AppKeyHmac 应为 header apiKey: %v", scheme)
	}
	if !strings.Contains(scheme["description"].(string), "hmac-sha256") {
		t.Fatalf("securitySchemes 描述应包含签名算法说明")
	}
}

// OpenAPI 生成边界：非法 params_json 报错；resp_sample_json 非法 JSON 时示例仍合法；无参数可生成
func TestBuildOpenApiJsonEdge(t *testing.T) {
	api := docSampleApi()
	//非法 params_json → 报错且不产出
	api.ParamsJson = `not-json`
	if out, err := BuildOpenApiJSON(api); err == nil || out != "" {
		t.Fatalf("非法 params_json 应报错, got %q %v", out, err)
	}
	//resp_sample_json 非法 JSON：示例回退占位，整体仍合法
	api = docSampleApi()
	api.RespSampleJson = `broken json`
	out, err := BuildOpenApiJSON(api)
	if err != nil || !json.Valid([]byte(out)) {
		t.Fatalf("非法响应样例应回退占位且输出合法 JSON: %v", err)
	}
	if !strings.Contains(out, "column1") {
		t.Fatalf("回退占位应出现在输出中")
	}
	//无业务参数
	api = docSampleApi()
	api.ParamsJson = ""
	out, err = BuildOpenApiJSON(api)
	if err != nil || !json.Valid([]byte(out)) {
		t.Fatalf("无参数 API 也应可生成合法 OpenAPI: %v", err)
	}
}

// Markdown 生成：标题/元信息/参数表/签名示例/响应示例/错误码表逐段断言
func TestBuildApiMarkdown(t *testing.T) {
	api := docSampleApi()
	md := BuildApiMarkdown(api)
	checks := []string{
		"# 用户查询",
		"`user/query`", //执行路径
		"请求方式：`POST`",  //请求方式
		"`published`",  //状态
		"结果缓存：`60s`",   //缓存
		"负责人：ops",
		"POST " + executeHttpPath, //网关统一入口
		"| 参数 | 类型 | 必填 | 默认值 |",
		"| id | number | 是 | - |",
		"| name | string | 否 | tom |",
		"| vip | bool | 否 | - |",
		"hmac-sha256(app_key + ts, app_secret)",   //签名说明
		"hmac.New(sha256.New, []byte(appSecret))", //Go 示例代码
		"curl -X POST",
		`"path":"user/query"`, //curl 内嵌 path
		"{\\\"id\\\":0",       //curl 内嵌 params_json（内层引号转义）
		"## 成功响应示例",
		"13800000000", //resp_sample_json 内容
		"| 错误码 | 状态 | 说明 |",
		"| app_not_found | denied |",
		"| sql_exec_error | error |",
		"| internal_error | error |",
	}
	for _, want := range checks {
		if !strings.Contains(md, want) {
			t.Fatalf("Markdown 缺少片段 %q\n%s", want, md)
		}
	}
	//代码围栏配对（```数量为偶数，文档可正常渲染）
	if n := strings.Count(md, "```"); n%2 != 0 {
		t.Fatalf("代码围栏应配对, got %d 个 ```", n)
	}
	//无参数 + 无缓存 + 非法响应样例的兜底渲染
	api2 := docSampleApi()
	api2.ParamsJson = ""
	api2.CacheTtl = 0
	api2.RespSampleJson = ""
	api2.Owner = ""
	md2 := BuildApiMarkdown(api2)
	for _, want := range []string{"该 API 无业务参数", "结果缓存：关闭", `[{"column1": "value1"}]`} {
		if !strings.Contains(md2, want) {
			t.Fatalf("兜底渲染缺少 %q\n%s", want, md2)
		}
	}
	if strings.Contains(md2, "负责人") {
		t.Fatal("空 owner 不应输出负责人")
	}
	//表格转义：name/default 含竖线时不得破坏表格
	api3 := docSampleApi()
	api3.ParamsJson = `[{"name":"a|b","type":"string","default":"x|y"}]`
	md3 := BuildApiMarkdown(api3)
	if !strings.Contains(md3, `a\|b`) || !strings.Contains(md3, `x\|y`) {
		t.Fatalf("竖线应被转义: %s", md3)
	}
}

// classifyCall：成功 / 各哨兵错误的 denied、error 归类 / 未识别错误兜底
func TestClassifyCall(t *testing.T) {
	//成功
	if status, code := classifyCall(nil); status != model.CallStatusOk || code != "" {
		t.Fatalf("nil 应归类 ok/空码, got %s %s", status, code)
	}
	cases := []struct {
		err        error
		wantStatus string
		wantCode   string
	}{
		{ErrAppNotFound, model.CallStatusDenied, "app_not_found"},
		{fmt.Errorf("%w: 签名不匹配", ErrSignInvalid), model.CallStatusDenied, "sign_invalid"},
		{ErrIpDenied, model.CallStatusDenied, "ip_denied"},
		{ErrRateLimited, model.CallStatusDenied, "rate_limited"},
		{ErrNotPublished, model.CallStatusDenied, "not_published"},
		{fmt.Errorf("%w: 授权已过期", ErrNotGranted), model.CallStatusDenied, "not_granted"},
		{ErrParamInvalid, model.CallStatusError, "param_invalid"},
		{fmt.Errorf("%w: %v", ErrDatasource, errors.New("conn refused")), model.CallStatusError, "datasource_error"},
		{ErrSqlExec, model.CallStatusError, "sql_exec_error"},
		{ErrSecretCrypto, model.CallStatusError, "secret_crypto_error"},
		{errors.New("boom"), model.CallStatusError, "internal_error"},
	}
	for _, c := range cases {
		status, code := classifyCall(c.err)
		if status != c.wantStatus || code != c.wantCode {
			t.Fatalf("classifyCall(%v) = (%s,%s), want (%s,%s)",
				c.err, status, code, c.wantStatus, c.wantCode)
		}
	}
	//错误码注册表：denied/error 之外不得出现未知状态；ok 不入表
	for _, spec := range callErrorSpecs {
		if spec.Status != model.CallStatusDenied && spec.Status != model.CallStatusError {
			t.Fatalf("错误码 %s 状态非法: %s", spec.Code, spec.Status)
		}
	}
}
