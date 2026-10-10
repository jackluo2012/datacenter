import { get, post, cleanParams } from './http'

/** 对齐 datasource.api（Phase 1 数据源管理） */
export interface DataSourceInfo {
  id?: number
  name?: string
  /** mysql/clickhouse/postgresql/hive... */
  type?: string
  host?: string
  port?: number
  db_name?: string
  username?: string
  /** 1=启用 0=停用 */
  status?: number
  owner?: string
  created_at?: number
  updated_at?: number
}

export interface DataSourceListResp {
  list: DataSourceInfo[]
  total: number
}

export interface TestConnResp {
  ok: boolean
  latency_ms: number
  err_msg?: string
}

export function createDataSource(
  req: {
    name: string
    type: string
    host?: string
    port?: number
    db_name?: string
    username?: string
    password?: string
    params_json?: string
    owner?: string
    beid?: number
  },
): Promise<DataSourceInfo> {
  return post<DataSourceInfo>('/datasource/create', req)
}

export function updateDataSource(req: {
  id: number
  name?: string
  type?: string
  host?: string
  port?: number
  db_name?: string
  username?: string
  password?: string
  params_json?: string
  status?: number
  owner?: string
  beid?: number
}): Promise<void> {
  return post<void>('/datasource/update', req)
}

export function deleteDataSource(id: number): Promise<void> {
  return post<void>('/datasource/delete', { id })
}

export function dataSourceInfo(id: number): Promise<DataSourceInfo> {
  return get<DataSourceInfo>('/datasource/info', cleanParams({ id }))
}

export function dataSourceList(params: {
  type?: string
  page?: number
  size?: number
}): Promise<DataSourceListResp> {
  return get<DataSourceListResp>('/datasource/list', cleanParams(params))
}

/** 测试连接：返回 ok + latency_ms（或 err_msg） */
export function dataSourceTest(id: number): Promise<TestConnResp> {
  return get<TestConnResp>('/datasource/test', cleanParams({ id }))
}
