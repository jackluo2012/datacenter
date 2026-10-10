import { get, post, cleanParams } from './http'

/** 对齐 ide.api（Phase 3 SQL IDE，仅 SELECT + 行数保护） */
export interface Script {
  id?: number
  name?: string
  sql_text?: string
  version?: number
  datasource_id?: number
  db_name?: string
  last_run_at?: number
}

export interface ScriptListResp {
  list: Script[]
  total: number
}

/** 执行结果：columns_json=列名数组，rows_json=对象数组 */
export interface RunScriptResp {
  columns_json: string
  rows_json: string
  row_count: number
  cost_ms: number
  truncated: boolean
}

export function saveScript(req: {
  id?: number
  name: string
  sql_text: string
  datasource_id: number
  db_name?: string
}): Promise<Script> {
  return post<Script>('/ide/script/save', req)
}

export function scriptList(params: {
  keyword?: string
  page?: number
  size?: number
}): Promise<ScriptListResp> {
  return get<ScriptListResp>('/ide/script/list', cleanParams(params))
}

export function scriptInfo(id: number): Promise<Script> {
  return get<Script>('/ide/script/info', cleanParams({ id }))
}

export function deleteScript(id: number): Promise<void> {
  return post<void>('/ide/script/delete', { id })
}

export function runScript(req: {
  id: number
  max_rows?: number
}): Promise<RunScriptResp> {
  return post<RunScriptResp>('/ide/script/run', req)
}
