import { get, post, cleanParams } from './http'

/** 对齐 job.api（Phase 3 调度集成：管控面，执行面推 DolphinScheduler） */
export interface SyncTask {
  id?: number
  name?: string
  source_ds_id?: number
  target_ds_id?: number
  /** full/incr */
  mode?: string
  mapping_json?: string
  cron?: string
  ds_workflow_code?: number
  status?: string
  last_run_at?: number
}

export interface SyncTaskListResp {
  list: SyncTask[]
  total: number
}

export interface DeployResp {
  ok: boolean
  ds_workflow_code?: number
  msg?: string
}

export interface TriggerResp {
  ok: boolean
  ds_process_instance_id?: number
  msg?: string
}

export interface TaskStatusResp {
  status?: string
  ds_instance_id?: number
  msg?: string
}

export interface SchedulerInfoResp {
  configured: boolean
  connected: boolean
  base_url?: string
  msg?: string
}

export function createSyncTask(req: {
  name: string
  source_ds_id: number
  target_ds_id?: number
  mode?: string
  mapping_json?: string
  cron?: string
}): Promise<SyncTask> {
  return post<SyncTask>('/job/task/create', req)
}

export function syncTaskList(params: {
  status?: string
  page?: number
  size?: number
}): Promise<SyncTaskListResp> {
  return get<SyncTaskListResp>('/job/task/list', cleanParams(params))
}

export function deleteSyncTask(id: number): Promise<void> {
  return post<void>('/job/task/delete', { id })
}

/** 部署任务到 DolphinScheduler */
export function deployTask(id: number): Promise<DeployResp> {
  return post<DeployResp>('/job/task/deploy', { id })
}

/** 手动触发任务 */
export function triggerTask(id: number): Promise<TriggerResp> {
  return post<TriggerResp>('/job/task/trigger', { id })
}

export function taskStatus(id: number): Promise<TaskStatusResp> {
  return get<TaskStatusResp>('/job/task/status', cleanParams({ id }))
}

/** 调度器连接信息 */
export function schedulerInfo(id?: number): Promise<SchedulerInfoResp> {
  return get<SchedulerInfoResp>('/job/scheduler/info', cleanParams({ id }))
}
