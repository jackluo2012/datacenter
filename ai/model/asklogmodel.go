package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// AskLog dc_ask_log 行（AI 问数审计表）
// 注意坑：字段顺序与 askLogRows 列清单一一对应（9 列对 9 字段），
// 若新增查询必须按 askLogRows 的列清单扫描，少列查询扫宽 struct 会报错
type AskLog struct {
	Id             int64     `db:"id"`
	Question       string    `db:"question"`
	CandidatesJson string    `db:"candidates_json"` //语义检索命中的候选(json)
	GeneratedSql   string    `db:"generated_sql"`   //LLM 生成的 SQL(未启用为空)
	Answer         string    `db:"answer"`
	LlmEnabled     int64     `db:"llm_enabled"` //1=本次 LLM 生成成功 0=未启用或已降级
	CostMs         int64     `db:"cost_ms"`
	Beid           int64     `db:"beid"`
	CreatedAt      time.Time `db:"created_at"`
}

// askLogRows 查询列（与 AskLog 字段顺序一致；text 列可空，扫描需 ifnull 兜底避免空值扫描报错）
const askLogRows = "`id`,`question`,ifnull(`candidates_json`,'') as `candidates_json`,ifnull(`generated_sql`,'') as `generated_sql`,ifnull(`answer`,'') as `answer`,`llm_enabled`,`cost_ms`,`beid`,`created_at`"

const askLogTable = "`dc_ask_log`"

type (
	// AskLogModel AI 问数审计表 model
	AskLogModel interface {
		// Insert 写入一条问数审计记录，返回自增 id
		Insert(ctx context.Context, data *AskLog) (int64, error)
	}

	defaultAskLogModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewAskLogModel 创建问数审计 model
func NewAskLogModel(conn sqlx.SqlConn) AskLogModel {
	return &defaultAskLogModel{
		conn:  conn,
		table: askLogTable,
	}
}

// Insert 写审计记录：显式列清单插入，created_at 由库默认 CURRENT_TIMESTAMP(3) 填充
func (m *defaultAskLogModel) Insert(ctx context.Context, data *AskLog) (int64, error) {
	query := "insert into " + m.table +
		" (`question`,`candidates_json`,`generated_sql`,`answer`,`llm_enabled`,`cost_ms`,`beid`) values (?,?,?,?,?,?,?)"
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Question, data.CandidatesJson, data.GeneratedSql, data.Answer,
		data.LlmEnabled, data.CostMs, data.Beid)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}
