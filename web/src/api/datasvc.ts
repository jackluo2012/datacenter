import { get, post, cleanParams } from './http'

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
