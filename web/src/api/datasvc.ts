import { get, post, cleanParams } from './http'

/**
 * Phase 7 落地深化：API 可观测接口路径集中定义（契约先写，后端路由合入后仅需改这里）。
 * 后端按 proto 合成 .api 时若路径有出入，改这份常量即可，视图层不感知。
 */
export const DATASVC_API = {
  /** 调用日志（query: api_id / app_id / status / page / size） */
  callLogList: '/datasvc/calllog/list',
  /** 调用统计（query: days=7|14|30） */
  callStats: '/datasvc/callstats',
  /** API 文档（query: path） */
  apiDoc: '/datasvc/apidoc',
} as const

/** 对齐 datasvc.api（Phase 1 数据服务，参考 OneService） */
export interface ApiInfo {
  id?: number
  name?: string
  path?: string
  /** sql 模式等 */
  mode?: string
  datasource_id?: number
  sql_text?: string
  params_json?: string
  cache_ttl?: number
  /** offline/published */
  status?: string
  owner?: string
}

export interface ApiListResp {
  list: ApiInfo[]
  total: number
}

export interface AppResp {
  app_id: number
  app_key: string
  /** 仅注册时返回一次 */
  app_secret: string
}

export function createApi(req: {
  name: string
  path: string
  mode?: string
  datasource_id: number
  sql_text: string
  params_json?: string
  cache_ttl?: number
  owner?: string
}): Promise<ApiInfo> {
  return post<ApiInfo>('/datasvc/api/create', req)
}

export function updateApi(req: {
  id: number
  name?: string
  sql_text?: string
  params_json?: string
  cache_ttl?: number
}): Promise<void> {
  return post<void>('/datasvc/api/update', req)
}

export function deleteApi(id: number): Promise<void> {
  return post<void>('/datasvc/api/delete', { id })
}

export function apiInfo(path: string): Promise<ApiInfo> {
  return get<ApiInfo>('/datasvc/api/info', cleanParams({ path }))
}

export function apiList(params: {
  status?: string
  page?: number
  size?: number
}): Promise<ApiListResp> {
  return get<ApiListResp>('/datasvc/api/list', cleanParams(params))
}

export function publishApi(id: number): Promise<void> {
  return post<void>('/datasvc/api/publish', { id })
}

export function offlineApi(id: number): Promise<void> {
  return post<void>('/datasvc/api/offline', { id })
}

/** 注册调用应用：app_secret 仅此一次返回 */
export function registerApp(req: {
  name: string
  rate_limit_qps?: number
  ip_whitelist?: string
}): Promise<AppResp> {
  return post<AppResp>('/datasvc/app/register', req)
}

/** 授权 API 给应用（expire_at 为 unix 秒，0 永久） */
export function grantApi(req: {
  api_id: number
  app_id: number
  expire_at?: number
  field_mask?: string
}): Promise<void> {
  return post<void>('/datasvc/grant', req)
}

/** 对外执行入口（不走 JWT，app_key+签名鉴权）；管理台一般不直接用 */
export function executeApi(req: {
  app_key: string
  sign: string
  ts: string
  path: string
  params_json?: string
}): Promise<{ rows_json: string; cost_ms: number; cache_hit: boolean; affected: number }> {
  return post('/datasvc/execute', req)
}

// ==================== Phase 7：API 可观测（调用统计 / 调用日志 / API 文档） ====================

/** 单条调用日志 */
export interface ApiCallLog {
  id?: number
  api_id?: number
  app_id?: number
  path?: string
  /** ok/error/denied */
  status?: string
  error_code?: string
  rows_cnt?: number
  cost_ms?: number
  cache_hit?: boolean
  ip?: string
  created_at?: number
}

export interface CallLogListResp {
  list: ApiCallLog[]
  total: number
}

/** 汇总指标 */
export interface ApiCallStats {
  total_calls?: number
  ok_calls?: number
  error_calls?: number
  denied_calls?: number
  /** 0~1 */
  success_rate?: number
  avg_cost_ms?: number
  cache_hits?: number
}

/** 调用量 Top API */
export interface TopApi {
  api_id?: number
  path?: string
  calls?: number
  avg_cost_ms?: number
}

/** /datasvc/callstats 返回 */
export interface CallStatsResp {
  overall?: ApiCallStats
  top_apis?: TopApi[]
  /** 近 N 天调用量（升序，与 daily_dates 一一对应） */
  daily_calls?: number[]
  /** 对应日期 YYYY-MM-DD */
  daily_dates?: string[]
}

/** /datasvc/apidoc 返回 */
export interface ApiDocResp {
  api_name?: string
  path?: string
  /** Markdown 文档（参数表 / 调用示例 / 签名说明） */
  markdown?: string
  /** OpenAPI 3.0 片段 JSON 字符串 */
  openapi_json?: string
}

export function listCallLogs(params: {
  api_id?: number
  app_id?: number
  /** ok/error/denied */
  status?: string
  page?: number
  size?: number
}): Promise<CallLogListResp> {
  return get<CallLogListResp>(DATASVC_API.callLogList, cleanParams(params))
}

export function getCallStats(days?: number): Promise<CallStatsResp> {
  return get<CallStatsResp>(DATASVC_API.callStats, cleanParams({ days: days || 7 }))
}

export function getApiDoc(path: string): Promise<ApiDocResp> {
  return get<ApiDocResp>(DATASVC_API.apiDoc, cleanParams({ path }))
}
