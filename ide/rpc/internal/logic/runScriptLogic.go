package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"datacenter/datasource/rpc/dataSourceClient"
	"datacenter/ide/model"
	"datacenter/ide/rpc/ide"
	"datacenter/ide/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// runDefaultMaxRows 行数保护缺省上限（入参 max_rows<=0 时生效）
const runDefaultMaxRows = 100

type RunScriptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunScriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunScriptLogic {
	return &RunScriptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行脚本（核心链路，任一步失败立即拒绝）：
// 取脚本（存在性+beid 归属）→ 安全闸（剥注释与前导括号后必须以 select 开头，禁止分号多语句）→
// datasource rpc 换目标库 DSN（with_secret=true）→ 目标库连接池按 DSN 缓存 →
// 手工扫描结果集（max_rows<=0 默认 100，超出截断 truncated=true）→
// 回填 last_run_at → 返回 columns_json/rows_json/row_count/cost_ms/truncated
func (l *RunScriptLogic) RunScript(in *ide.RunScriptReq) (*ide.RunScriptReply, error) {
	start := time.Now()
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrScriptNotFound
	}

	//1. 取脚本（FindOne 已带多租户过滤，越权/不存在统一按脚本不存在处理）
	script, err := l.svcCtx.ScriptModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrScriptNotFound
		}
		return nil, err
	}

	//2. 安全闸：仅放行单条 SELECT（写语句与多语句注入片段一律拒绝）
	if err := ValidateSelectSql(script.SqlText); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSql, err)
	}

	//3. datasource rpc 换目标库 DSN（with_secret=true 返回解密密码与拼好的连接串）
	if script.DatasourceId <= 0 {
		return nil, fmt.Errorf("%w: 脚本未配置数据源", ErrDatasource)
	}
	dsReply, err := l.svcCtx.DataSourceRpc.GetDataSource(l.ctx, &dataSourceClient.GetDataSourceReq{
		Id:         script.DatasourceId,
		Beid:       in.Beid,
		WithSecret: true,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: 数据源 %d: %v", ErrDatasource, script.DatasourceId, err)
	}
	if strings.TrimSpace(dsReply.GetDsn()) == "" {
		return nil, fmt.Errorf("%w: 数据源 %d 未返回 DSN", ErrDatasource, script.DatasourceId)
	}

	//4. 目标库连接（ServiceContext 按 DSN 缓存连接池）；RawDB 不可 Close，底层连接池需复用
	conn, err := l.svcCtx.TargetConn(dsReply.GetDsn())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}
	db, err := conn.RawDB()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasource, err)
	}

	//5. 行数保护：max_rows<=0 默认 100，超出截断
	maxRows := int(in.MaxRows)
	if maxRows <= 0 {
		maxRows = runDefaultMaxRows
	}
	columns, rows, truncated, err := QueryRowsAsMaps(l.ctx, db, script.SqlText, maxRows)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScriptExec, err)
	}

	//6. 回填最近执行时间（失败仅记日志，不影响结果返回）
	if uerr := l.svcCtx.ScriptModel.UpdateLastRunAt(l.ctx, script.Id, in.Beid, time.Now()); uerr != nil {
		l.Errorf("[RunScript] 回填 last_run_at 失败 script=%d err=%v", script.Id, uerr)
	}

	//7. 结果序列化：columns_json=列名数组，rows_json=通用行集数组（空结果输出 []）
	columnsJson, err := json.Marshal(columns)
	if err != nil {
		return nil, fmt.Errorf("%w: 列信息序列化失败: %v", ErrScriptExec, err)
	}
	rowsJson, err := json.Marshal(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: 结果集序列化失败: %v", ErrScriptExec, err)
	}
	l.Infof("[RunScript] script=%d datasource=%d rows=%d truncated=%v cost=%dms",
		script.Id, script.DatasourceId, len(rows), truncated, time.Since(start).Milliseconds())
	return &ide.RunScriptReply{
		ColumnsJson: string(columnsJson),
		RowsJson:    string(rowsJson),
		RowCount:    int64(len(rows)),
		CostMs:      time.Since(start).Milliseconds(),
		Truncated:   truncated,
	}, nil
}
