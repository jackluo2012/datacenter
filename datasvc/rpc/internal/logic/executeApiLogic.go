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

// 执行 API（核心链路，严格按序，任一步失败立即拒绝）：
// a 应用鉴权(app_key+签名+ts) → b IP 白名单 → c 应用限流 → d API 配置(status=published)+授权校验
// → e 参数校验 → f SQL 构建(:name→? 白名单替换) → g 结果缓存 → h 脱敏 → i 返回 rows/affected/cost
func (l *ExecuteApiLogic) ExecuteApi(in *datasvc.ExecuteReq) (*datasvc.ExecuteReply, error) {
	start := time.Now()

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
	return &datasvc.ExecuteReply{
		Affected: affected,
		CostMs:   time.Since(start).Milliseconds(),
	}, nil
}
