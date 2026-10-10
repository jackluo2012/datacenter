import { get, post, cleanParams } from './http'

/** 对齐 ops.api（Phase 5 数据运维与系统日志） */
export interface OpsArchiveTask {
  id?: number
  name?: string
  datasource_id?: number
  db_name?: string
  table_name?: string
  condition_sql?: string
  archive_table?: string
  delete_source?: boolean
  batch_size?: number
  cron?: string
  last_status?: string
  last_run_at?: number
}

export interface ArchiveTaskListResp {
  list: OpsArchiveTask[]
  total: number
}

export interface RunArchiveResp {
  archived_rows: number
  deleted_rows: number
  status: string
}

export interface OpsArchiveLog {
  id?: number
  task_id?: number
  archived_rows?: number
  deleted_rows?: number
  status?: string
  msg?: string
  run_at?: number
}

export interface OpsLoginLog {
  id?: number
  username?: string
  uid?: number
  ip?: string
  user_agent?: string
  success: boolean
  msg?: string
  created_at?: number
}

export interface OpsAuditLog {
  id?: number
  uid?: number
  username?: string
  action?: string
  resource?: string
  resource_id?: string
  detail_json?: string
  ip?: string
  created_at?: number
}

export interface ListResp<T> {
  list: T[]
  total: number
}

// ---------- 数据归档 ----------
export function createArchiveTask(req: {
  name: string
  datasource_id: number
  db_name?: string
  table_name: string
  condition_sql?: string
  archive_table: string
  delete_source?: boolean
  batch_size?: number
  cron?: string
}): Promise<OpsArchiveTask> {
  return post<OpsArchiveTask>('/ops/archive/create', req)
}

export function archiveTaskList(params: {
  page?: number
  size?: number
}): Promise<ArchiveTaskListResp> {
  return get<ArchiveTaskListResp>('/ops/archive/list', cleanParams(params))
}

export function deleteArchiveTask(id: number): Promise<void> {
  return post<void>('/ops/archive/delete', { id })
}

export function runArchive(id: number): Promise<RunArchiveResp> {
  return post<RunArchiveResp>('/ops/archive/run', { id })
}

export function archiveLogList(params: {
  task_id?: number
  page?: number
  size?: number
}): Promise<ListResp<OpsArchiveLog>> {
  return get<ListResp<OpsArchiveLog>>('/ops/archive/log/list', cleanParams(params))
}

// ---------- 系统日志 ----------
/** 登录日志（success_filter: 0 全部 1 成功 2 失败） */
export function loginLogList(params: {
  username?: string
  success_filter?: number
  page?: number
  size?: number
}): Promise<ListResp<OpsLoginLog>> {
  return get<ListResp<OpsLoginLog>>('/ops/log/login', cleanParams(params))
}

/** 审计日志 */
export function auditLogList(params: {
  resource?: string
  uid?: number
  page?: number
  size?: number
}): Promise<ListResp<OpsAuditLog>> {
  return get<ListResp<OpsAuditLog>>('/ops/log/audit', cleanParams(params))
}
