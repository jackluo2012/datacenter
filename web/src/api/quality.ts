import { get, post, cleanParams } from './http'

/**
 * Phase 7 落地深化：新增接口路径集中定义（契约先写，后端路由合入后仅需改这里）。
 * 后端按 proto 合成 .api 时若路径有出入，改这份常量即可，视图层不感知。
 */
export const QUALITY_API = {
  /** 保存告警通道（id=0 新建，id>0 更新） */
  alertChannelSave: '/quality/alertchannel/save',
  alertChannelDelete: '/quality/alertchannel/delete',
  alertChannelList: '/quality/alertchannel/list',
  /** 发送测试消息，返回 {ok,msg} */
  alertChannelTest: '/quality/alertchannel/test',
  /** 错误数据样本（query: job_id / rule_id） */
  errorDataList: '/quality/errordata/list',
  /** 表健康分析（POST） */
  profileTable: '/quality/profile/table',
  /** 历史健康分列表 */
  profileList: '/quality/profile/list',
} as const

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

// ==================== Phase 7：告警通道 / 错误数据 / 表健康分 ====================

/** 告警通道（钉钉 / 飞书机器人） */
export interface AlertChannel {
  id?: number
  name?: string
  /** dingtalk/feishu */
  type?: string
  webhook_url?: string
  /** 加签密钥，钉钉机器人可选 */
  secret?: string
  /** failed/partial：触发告警的最低结果等级 */
  min_level?: string
  /** 1=启用 0=停用 */
  status?: number
}

export interface AlertChannelListResp {
  list: AlertChannel[]
  total: number
}

/** 通道测试结果 */
export interface TestAlertResp {
  ok: boolean
  msg: string
}

/** 错误数据样本：rows_json 为 JSON 数组字符串 */
export interface ErrorData {
  rule_id?: number
  job_id?: number
  rows_json?: string
  sample_count?: number
  run_at?: number
}

export interface ErrorDataListResp {
  list: ErrorData[]
  total: number
}

/** 表健康画像结果 */
export interface TableProfile {
  id?: number
  datasource_id?: number
  /** 库.表 或表名（以后端返回为准） */
  table_name?: string
  rows_cnt?: number
  cols_cnt?: number
  /** 空值率 0~1 */
  null_rate?: number
  dup_pk_cnt?: number
  /** 健康分 0~100 */
  health_score?: number
  /** 逐列画像 JSON（数组字符串） */
  profile_json?: string
  ran_at?: number
}

/** /quality/profile/table 返回包装；部分实现可能直接回 TableProfile，视图层做兼容 */
export interface ProfileReply {
  profile?: TableProfile
}

export interface ProfileListResp {
  list: TableProfile[]
  total: number
}

/** 保存告警通道：id 传 0/缺省即新建 */
export function saveAlertChannel(req: {
  id?: number
  name: string
  /** dingtalk/feishu */
  type: string
  webhook_url: string
  secret?: string
  /** failed/partial */
  min_level: string
  /** 启停状态（后端契约确认中，未启用时后端会忽略） */
  status?: number
}): Promise<AlertChannel> {
  return post<AlertChannel>(QUALITY_API.alertChannelSave, req)
}

export function deleteAlertChannel(id: number): Promise<void> {
  return post<void>(QUALITY_API.alertChannelDelete, { id })
}

export function listAlertChannels(params?: {
  page?: number
  size?: number
}): Promise<AlertChannelListResp> {
  return get<AlertChannelListResp>(QUALITY_API.alertChannelList, cleanParams(params))
}

/** 向通道发送测试消息 */
export function testAlertChannel(id: number): Promise<TestAlertResp> {
  return post<TestAlertResp>(QUALITY_API.alertChannelTest, { id })
}

export function listErrorData(params: {
  job_id?: number
  rule_id?: number
  page?: number
  size?: number
}): Promise<ErrorDataListResp> {
  return get<ErrorDataListResp>(QUALITY_API.errorDataList, cleanParams(params))
}

/** 表健康分析：对指定表做画像并返回健康分 */
export function profileTable(req: {
  datasource_id: number
  db_name?: string
  table_name: string
  pk_column?: string
}): Promise<ProfileReply> {
  return post<ProfileReply>(QUALITY_API.profileTable, req)
}

export function listProfiles(params: {
  datasource_id?: number
  table_name?: string
  page?: number
  size?: number
}): Promise<ProfileListResp> {
  return get<ProfileListResp>(QUALITY_API.profileList, cleanParams(params))
}
