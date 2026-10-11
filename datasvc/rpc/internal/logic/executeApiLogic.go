package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteApiLogic {
	return &ExecuteApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// callTrace 单次执行的埋点上下文：随执行链推进逐步补全，仅作埋点出参，不参与任何链路判断
type callTrace struct {
	apiId    int64  //API id（走到 d 步查询 API 后才有值）
	appId    int64  //应用 id（走到 a 步鉴权通过后才有值）
	path     string //执行路径（请求原文，鉴权前失败也保留）
	rowsCnt  int64  //返回行数/写语句受影响行数
	cacheHit bool   //是否命中结果缓存
}

// ExecuteApi 执行 API + 调用日志埋点：
// 原执行链（鉴权→限流→授权→执行→脱敏）语义完全不变，仅在其外包一层埋点——
// 成功/失败（denied/error）各出口统一异步落 dc_api_call_log（goroutine+recover，失败仅日志不阻断）
func (l *ExecuteApiLogic) ExecuteApi(in *datasvc.ExecuteReq) (*datasvc.ExecuteReply, error) {
	start := time.Now()
	trace := &callTrace{path: strings.TrimSpace(in.Path)}
	reply, execErr := l.executeCore(in, start, trace)
	//埋点：所有出口统一落调用日志（异步，不阻断返回）
	l.recordCallLog(in, trace, reply, execErr, time.Since(start).Milliseconds())
	return reply, execErr
}

// executeCore 原执行链（严格按序，任一步失败立即拒绝）：
// a 应用鉴权(app_key+签名+ts) → b IP 白名单 → c 应用限流 → d API 配置(status=published)+授权校验
// → e 参数校验 → f SQL 构建(:name→? 白名单替换) → g 结果缓存 → h 脱敏 → i 返回 rows/affected/cost
// trace 仅追加埋点赋值，不改变原有任何判断与返回
func (l *ExecuteApiLogic) executeCore(in *datasvc.ExecuteReq, start time.Time, trace *callTrace) (*datasvc.ExecuteReply, error) {
	//a. 应用鉴权：app_key 查应用（不存在→ErrAppNotFound）
	app, err := l.svcCtx.AppModel.FindByKey(strings.TrimSpace(in.AppKey))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}
	//多租户隔离：beid 不匹配按应用不存在处理，避免泄漏存在性
	if in.Beid != 0 && app.Beid != in.Beid {
		return nil, ErrAppNotFound
	}
	trace.appId = app.Id //埋点：鉴权通过，记录 app_id
	//比对签名：hex(hmac-sha256(app_key+ts))，密钥为解密后的 app_secret；ts 偏差>300s 拒绝
	secret, err := decryptAppSecret(l.svcCtx, app.AppSecretEnc)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSecretCrypto, err)
	}
	if err := verifySign(app.AppKey, in.Ts, in.Sign, secret, time.Now()); err != nil {
		return nil, err
	}

	//b. IP 白名单（空=不限；非空且 ip 不在列表→ErrIpDenied）
	if !ipAllowed(in.Ip, app.IpWhitelist) {
		return nil, ErrIpDenied
	}

	//c. 限流：每应用 rate_limit_qps（进程内令牌桶，超限→ErrRateLimited）
	if !l.svcCtx.Allow(app.Id, app.RateLimitQps) {
		return nil, ErrRateLimited
	}

	//d. 查 API 配置（path+beid）且 status=published；查授权（api_id+app_id，含过期时间校验）
	beid := in.Beid
	if beid == 0 {
		beid = app.Beid //未显式传租户时回落应用所属租户
	}
	api, err := l.svcCtx.ApiModel.FindByPath(beid, strings.TrimSpace(in.Path))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrApiNotFound
		}
		return nil, err
	}
	trace.apiId = api.Id //埋点：API 已解析，记录 api_id
	if api.Status != "published" {
		return nil, ErrNotPublished
	}
	grant, err := l.svcCtx.GrantModel.FindByApiAndApp(api.Id, app.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrNotGranted
		}
		return nil, err
	}
	if grant.ExpireAt.Valid && !time.Now().Before(grant.ExpireAt.Time) {
		return nil, fmt.Errorf("%w: 授权已过期", ErrNotGranted)
	}

	//e. 参数校验：按 params_json 定义收敛业务参数（缺 required 报错/多余拒绝/缺省补 default）
	defs, err := ParseParamDefs(api.ParamsJson)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
	}
	values, err := ValidateParams(defs, in.ParamsJson)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
	}

	//f. SQL 执行准备：:name 占位符按白名单顺序替换为 ? 并组装参数数组
	sqlText, args, err := BuildSql(api.SqlText, defs, values)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParamInvalid, err)
	}
	isSelect := IsSelectStmt(sqlText)
	//写语句（mode=sql 且非 SELECT 开头）不缓存
	cacheable := api.CacheTtl > 0 && isSelect
	var cacheKey string
	if cacheable {
		cacheKey = CacheKey(api.Id, values)
		var cached string
		if err := l.svcCtx.ResultCache.Get(cacheKey, &cached); err == nil {
			if rows, derr := decodeRows(cached); derr == nil {
				//h. 脱敏（缓存存原始结果，脱敏按授权在出口执行）
				MaskRows(rows, grant.FieldMask)
				if out, merr := json.Marshal(rows); merr == nil {
					trace.cacheHit = true //埋点：缓存命中
					trace.rowsCnt = int64(len(rows))
					return &datasvc.ExecuteReply{
						RowsJson: string(out),
						CostMs:   time.Since(start).Milliseconds(),
						CacheHit: true,
					}, nil
				}
			}
			l.Errorf("ExecuteApi 缓存值解析失败，降级直查 api=%d key=%s", api.Id, cacheKey)
		} else if !l.svcCtx.ResultCache.IsNotFound(err) {
			l.Errorf("ExecuteApi 读缓存失败，降级直查 api=%d err=%v", api.Id, err)
		}
	}

	//获取目标库 DSN：datasource rpc GetDataSource(id, beid, with_secret=true)
	dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         api.DatasourceId,
		Beid:       beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	if strings.TrimSpace(dsReply.GetDsn()) == "" {
		return nil, fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, api.DatasourceId)
	}
	//连接目标库：sqlx.NewMysql(dsn) 直连，连接池按 DSN 缓存复用
	conn, err := l.svcCtx.TargetConn(dsReply.Dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	db, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}

	if isSelect {
		//SELECT：结果集转 JSON
		rows, err := QueryRowsAsMaps(l.ctx, db, sqlText, args)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSqlExec, err)
		}
		rawJson, err := json.Marshal(rows)
		if err != nil {
			return nil, fmt.Errorf("%w: 结果集序列化失败: %v", ErrSqlExec, err)
		}
		//回填缓存（存原始结果，TTL=cache_ttl；写缓存失败不影响返回）
		if cacheable {
			if err := l.svcCtx.ResultCache.SetWithExpire(cacheKey, string(rawJson),
				time.Duration(api.CacheTtl)*time.Second); err != nil {
				l.Errorf("ExecuteApi 写缓存失败 api=%d err=%v", api.Id, err)
			}
		}
		//h. 脱敏：grant.field_mask 非空时命中列名的值替换为 ****
		MaskRows(rows, grant.FieldMask)
		out, err := json.Marshal(rows)
		if err != nil {
			return nil, fmt.Errorf("%w: 结果集序列化失败: %v", ErrSqlExec, err)
		}
		trace.rowsCnt = int64(len(rows)) //埋点：返回行数
		//i. 返回 rows_json/cost_ms
		return &datasvc.ExecuteReply{
			RowsJson: string(out),
			CostMs:   time.Since(start).Milliseconds(),
		}, nil
	}

	//写语句：返回受影响行数（不缓存）
	affected, err := ExecWrite(l.ctx, db, sqlText, args)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqlExec, err)
	}
	trace.rowsCnt = affected //埋点：受影响行数
	return &datasvc.ExecuteReply{
		Affected: affected,
		CostMs:   time.Since(start).Milliseconds(),
	}, nil
}

// recordCallLog 调用日志埋点（异步落库）：
// 1) goroutine+recover 兜底，埋点属旁路逻辑，panic/落库失败仅记日志，绝不阻断执行链返回；
// 2) ip 取请求透传的 ExecuteReq.Ip（网关当前不透传则落空串）；
// 3) 成功路径 cost_ms 以执行链统计值为准，与返回给调用方的一致
func (l *ExecuteApiLogic) recordCallLog(in *datasvc.ExecuteReq, trace *callTrace,
	reply *datasvc.ExecuteReply, execErr error, costMs int64) {
	if reply != nil && reply.CostMs > 0 {
		costMs = reply.CostMs
	}
	status, errorCode := classifyCall(execErr)
	entry := &model.ApiCallLog{
		ApiId:     trace.apiId,
		AppId:     trace.appId,
		Path:      trace.path,
		Status:    status,
		ErrorCode: errorCode,
		RowsCnt:   trace.rowsCnt,
		CostMs:    costMs,
		CacheHit:  cacheHitFlag(trace.cacheHit),
		Ip:        strings.TrimSpace(in.Ip),
	}
	go func() {
		//panic 兜底：埋点任何异常只记日志，不外溢
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("ExecuteApi 调用日志埋点 panic: %v", r)
			}
		}()
		if _, err := l.svcCtx.CallLogModel.Insert(entry); err != nil {
			logx.Errorf("ExecuteApi 调用日志落库失败 path=%s status=%s err=%v",
				entry.Path, entry.Status, err)
		}
	}()
}

// cacheHitFlag bool → tinyint(0/1)
func cacheHitFlag(hit bool) int64 {
	if hit {
		return 1
	}
	return 0
}

// callErrorSpec 执行结果错误码规格：埋点分类（classifyCall）与文档错误码表（BuildApiMarkdown）
// 共用同一份注册表，保证落库 error_code 与对外文档口径一致
type callErrorSpec struct {
	Code   string //落库 error_code / 文档错误码
	Status string //ok/error/denied
	Err    error  //执行链哨兵错误（error.go）
	Desc   string //文档中的语义说明
}

// callErrorSpecs 错误码注册表：前 6 项属鉴权/授权类拒绝（denied），后 4 项属执行类错误（error）
var callErrorSpecs = []callErrorSpec{
	{Code: "app_not_found", Status: model.CallStatusDenied, Err: ErrAppNotFound, Desc: "应用不存在或 app_key 错误"},
	{Code: "sign_invalid", Status: model.CallStatusDenied, Err: ErrSignInvalid, Desc: "签名校验失败或 ts 超出时间窗口"},
	{Code: "ip_denied", Status: model.CallStatusDenied, Err: ErrIpDenied, Desc: "来源 IP 不在应用白名单内"},
	{Code: "rate_limited", Status: model.CallStatusDenied, Err: ErrRateLimited, Desc: "超过应用限流阈值，请稍后重试"},
	{Code: "not_published", Status: model.CallStatusDenied, Err: ErrNotPublished, Desc: "API 未发布，拒绝执行"},
	{Code: "not_granted", Status: model.CallStatusDenied, Err: ErrNotGranted, Desc: "应用未获得该 API 授权或授权已过期"},
	{Code: "param_invalid", Status: model.CallStatusError, Err: ErrParamInvalid, Desc: "参数校验失败（缺必填/多余/类型错误）"},
	{Code: "datasource_error", Status: model.CallStatusError, Err: ErrDatasource, Desc: "获取数据源连接失败"},
	{Code: "sql_exec_error", Status: model.CallStatusError, Err: ErrSqlExec, Desc: "目标库 SQL 执行失败"},
	{Code: "secret_crypto_error", Status: model.CallStatusError, Err: ErrSecretCrypto, Desc: "应用密钥加解密失败"},
}

// classifyCall 把执行结果 error 映射为（status, error_code）（纯函数）：
// nil → (ok, "")；命中哨兵 → 注册表口径；未识别错误按 error/internal_error 兜底
func classifyCall(err error) (string, string) {
	if err == nil {
		return model.CallStatusOk, ""
	}
	for _, spec := range callErrorSpecs {
		if errors.Is(err, spec.Err) {
			return spec.Status, spec.Code
		}
	}
	return model.CallStatusError, "internal_error"
}
