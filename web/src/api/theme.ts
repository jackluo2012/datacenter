import { get, post, cleanParams } from './http'

/**
 * 对齐 theme.api（Phase 6 数仓建模：主题域 / 主题 / 分层模型 / 物化视图）。
 * 响应为裸对象：theme 网关 handler 直接 httpx.OkJsonCtx 输出，无 {code,msg,data} 包装，
 * http.ts 拦截器对无 code 字段 body 原样透传，故此处类型即后端返回结构。
 *
 * 租户口径：theme rpc 全部 logic 要求 beid>0（缺省直接拒），管理台固定 beid=1
 * （与 security/bigscreen 等模块口径一致），在此统一注入，调用方无需关心。
 *
 * 契约差异（编辑能力）：网关 SaveThemeDomainReq / SaveThemeReq / SaveThemeModelReq
 * 均无 id 字段（saveThemeMv 的 rpc 请求同样无 id），rpc 侧恒按新建处理——即当前网关
 * 只支持「新建 + 删除」，不支持按 id 更新；页面不提供编辑入口。
 */

/** 管理台统一租户 ID */
const BEID = 1

// ---------- 类型 ----------

/** 主题域（dc_theme_domain；同租户下 code 唯一，小写标识符） */
export interface ThemeDomainInfo {
  id?: number
  code?: string
  name?: string
  comment?: string
}

/** 主题（dc_theme；挂靠主题域，同租户下 code 唯一） */
export interface ThemeInfo {
  id?: number
  domain_id?: number
  code?: string
  name?: string
  comment?: string
}

/** 分层模型（dc_theme_model；物理表名 <layer>_<table_name>，create_sql 由后端实时生成预览 DDL） */
export interface ThemeModelInfo {
  id?: number
  theme_id?: number
  /** ods/dwd/dws/ads/dim */
  layer?: string
  table_name?: string
  model_name?: string
  /** JSON 数组 [{name,type,comment?,pk?,dim?}]，pk 恰有一个 */
  fields_json?: string
  /** 可选，JSON 数组 [{dim_table,join_key}] */
  dim_json?: string
  create_sql?: string
  datasource_id?: number
  db_name?: string
}

/** 物化视图定义（dc_theme_mv；刷新为 create or replace view 真实执行） */
export interface ThemeMvInfo {
  id?: number
  name?: string
  datasource_id?: number
  db_name?: string
  view_name?: string
  select_sql?: string
  /** manual/schedule */
  refresh_mode?: string
  /** success/failed，空=尚未刷新 */
  last_status?: string
  last_refresh_at?: number
}

/** 模型部署 / 物化视图刷新共用结果：失败原因在 msg 随正常响应返回（不作为 error 抛出） */
export interface ThemeOpResp {
  ok: boolean
  msg?: string
  cost_ms?: number
}

// ---------- 响应 ----------

/** 主题域全量列表（无分页参数，域数量少不分页） */
export interface ThemeDomainListResp {
  list: ThemeDomainInfo[]
}

export interface ThemeListResp {
  list: ThemeInfo[]
  total: number
}

export interface ThemeModelListResp {
  list: ThemeModelInfo[]
  total: number
}

export interface ThemeMvListResp {
  list: ThemeMvInfo[]
  total: number
}

// ---------- 主题域 ----------

/** 新建主题域（网关无 id，恒新建）；code 须小写字母开头仅小写字母/数字/下划线，同租户查重 */
export function saveThemeDomain(req: {
  code: string
  name: string
  comment?: string
}): Promise<ThemeDomainInfo> {
  return post<ThemeDomainInfo>('/theme/domain/save', { ...req, beid: BEID })
}

/** 删除主题域：域下存在主题时后端拒绝（防悬空引用） */
export function deleteThemeDomain(id: number): Promise<void> {
  return post<void>('/theme/domain/delete', { id, beid: BEID })
}

/** 主题域全量列表（按 id 倒序，不分页） */
export function themeDomainList(): Promise<ThemeDomainListResp> {
  return get<ThemeDomainListResp>('/theme/domain/list', cleanParams({ beid: BEID }))
}

// ---------- 主题 ----------

/** 新建主题（网关无 id，恒新建）；domain_id 须存在，code 同租户查重 */
export function saveTheme(req: {
  domain_id: number
  code: string
  name: string
  comment?: string
}): Promise<ThemeInfo> {
  return post<ThemeInfo>('/theme/save', { ...req, beid: BEID })
}

/** 删除主题：主题下存在模型时后端拒绝 */
export function deleteTheme(id: number): Promise<void> {
  return post<void>('/theme/delete', { id, beid: BEID })
}

/** 主题列表（domain_id 过滤 + 分页） */
export function themeList(params: {
  domain_id?: number
  page?: number
  size?: number
}): Promise<ThemeListResp> {
  return get<ThemeListResp>('/theme/list', cleanParams({ beid: BEID, ...params }))
}

// ---------- 分层模型 ----------

/** 新建分层模型（网关无 id，恒新建）；layer 白名单 ods/dwd/dws/ads/dim，
 *  table_name 小写标识符且 <layer>_<table> 整体 ≤64 字符，fields_json 校验
 *  （name 合法不重复 + type 白名单 + pk 恰有一个），同租户 table_name 查重 */
export function saveThemeModel(req: {
  theme_id: number
  layer: string
  table_name: string
  model_name: string
  fields_json: string
  dim_json?: string
  datasource_id: number
  db_name?: string
}): Promise<ThemeModelInfo> {
  return post<ThemeModelInfo>('/theme/model/save', { ...req, beid: BEID })
}

/** 删除分层模型 */
export function deleteThemeModel(id: number): Promise<void> {
  return post<void>('/theme/model/delete', { id, beid: BEID })
}

/** 模型列表（theme_id / layer 过滤 + 分页） */
export function themeModelList(params: {
  theme_id?: number
  layer?: string
  page?: number
  size?: number
}): Promise<ThemeModelListResp> {
  return get<ThemeModelListResp>('/theme/model/list', cleanParams({ beid: BEID, ...params }))
}

/** 部署模型：在目标数据源真实执行幂等建表 DDL（create table if not exists <layer>_<table>）。
 *  失败不抛错，以 ok=false + 失败原因 msg 回告；成功 msg 携带生成 SQL 摘要 */
export function deployThemeModel(id: number): Promise<ThemeOpResp> {
  return post<ThemeOpResp>('/theme/model/deploy', { id, beid: BEID })
}

// ---------- 物化视图 ----------

/** 新建物化视图定义（rpc 侧无 id，仅建档不更新）；view_name 小写标识符，
 *  select_sql 须 SELECT 开头且不含分号，refresh_mode 空缺省 manual */
export function saveThemeMv(req: {
  name: string
  datasource_id: number
  db_name?: string
  view_name: string
  select_sql: string
  refresh_mode?: string
}): Promise<ThemeMvInfo> {
  return post<ThemeMvInfo>('/theme/mv/save', { ...req, beid: BEID })
}

/** 删除物化视图定义 */
export function deleteThemeMv(id: number): Promise<void> {
  return post<void>('/theme/mv/delete', { id, beid: BEID })
}

/** 物化视图列表（分页） */
export function themeMvList(params: {
  page?: number
  size?: number
}): Promise<ThemeMvListResp> {
  return get<ThemeMvListResp>('/theme/mv/list', cleanParams({ beid: BEID, ...params }))
}

/** 刷新物化视图：create or replace view 真实执行并回写 last_status/last_refresh_at。
 *  成功/失败均以 ok+msg 回告（失败原因在 msg，不作为 error 抛出） */
export function refreshThemeMv(id: number): Promise<ThemeOpResp> {
  return post<ThemeOpResp>('/theme/mv/refresh', { id, beid: BEID })
}
