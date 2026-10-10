import { get, post, cleanParams } from './http'

/** 对齐 quality.api（Phase 2 数据质量） */
export interface QualityRule {
  id?: number
  name?: string
  /** not_null/unique/range/sql ... */
  rule_type?: string
  target_table?: string
  target_column?: string
  params_json?: string
  threshold_json?: string
  status?: number
  owner?: string
}

export interface QualityRuleListResp {
  list: QualityRule[]
  total: number
}

export interface QualityJob {
  id?: number
  name?: string
  datasource_id?: number
  db_name?: string
  rule_ids?: string
  cron?: string
  last_status?: string
  last_run_at?: number
}

export interface JobListResp {
  list: QualityJob[]
  total: number
}

export interface RunJobResp {
  job_id: number
  total_rules: number
  passed: number
  last_status: string
}

export interface QualityResult {
  id?: number
  job_id?: number
  rule_id?: number
  rule_type?: string
  check_value?: string
  pass_flag?: boolean
  detail_json?: string
  run_at?: number
}

export interface ResultListResp {
  list: QualityResult[]
  total: number
}

export function createQualityRule(req: {
  name: string
  rule_type: string
  target_table: string
  target_column?: string
  params_json?: string
  threshold_json?: string
  owner?: string
}): Promise<QualityRule> {
  return post<QualityRule>('/quality/rule/create', req)
}

export function deleteQualityRule(id: number): Promise<void> {
  return post<void>('/quality/rule/delete', { id })
}

export function qualityRuleList(params: {
  rule_type?: string
  page?: number
  size?: number
}): Promise<QualityRuleListResp> {
  return get<QualityRuleListResp>('/quality/rule/list', cleanParams(params))
}

export function createQualityJob(req: {
  name: string
  datasource_id: number
  db_name?: string
  /** 逗号分隔的规则 id */
  rule_ids: string
  cron?: string
}): Promise<QualityJob> {
  return post<QualityJob>('/quality/job/create', req)
}

export function qualityJobList(params: {
  page?: number
  size?: number
}): Promise<JobListResp> {
  return get<JobListResp>('/quality/job/list', cleanParams(params))
}

export function runQualityJob(id: number): Promise<RunJobResp> {
  return post<RunJobResp>('/quality/job/run', { id })
}

export function qualityResultList(params: {
  job_id?: number
  pass_filter?: number
  page?: number
  size?: number
}): Promise<ResultListResp> {
  return get<ResultListResp>('/quality/result/list', cleanParams(params))
}
