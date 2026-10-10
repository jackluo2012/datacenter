import { post } from './http'

/** 对齐 ai.api（Phase 4 AI 问数：语义检索 + 可选 LLM，全程审计） */
export interface Candidate {
  /** entity/metric */
  type?: string
  id?: number
  name?: string
  detail?: string
  comment?: string
}

export interface AskResp {
  candidates: Candidate[]
  llm_enabled: boolean
  generated_sql?: string
  answer: string
  cost_ms: number
}

export interface AskAndRunResp extends AskResp {
  executed: boolean
  /** 列名数组 JSON */
  columns_json?: string
  /** 对象数组 JSON */
  rows_json?: string
  row_count: number
  run_error?: string
}

/** AI 问数（只检索+生成 SQL，不执行） */
export function ask(req: { question: string; max_candidates?: number }): Promise<AskResp> {
  return post<AskResp>('/ai/ask', req)
}

/**
 * AI 问数并执行：auto_run=true 时 LLM 生成 SQL 后经 ide 安全闸自动执行
 * @param datasourceId auto_run 时必填：执行目标库
 * @param maxRows 默认 20，上限 100
 */
export function askAndRun(req: {
  question: string
  auto_run?: boolean
  datasource_id?: number
  max_rows?: number
  max_candidates?: number
}): Promise<AskAndRunResp> {
  return post<AskAndRunResp>('/ai/ask_and_run', req)
}
