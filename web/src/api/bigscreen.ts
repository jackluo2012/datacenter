/**
 * 数据大屏（/bigscreen）API 集中封装
 *
 * - 复用统一 http 实例：走既有 JWT 请求拦截器 + {code,data} 响应剥壳
 * - 大屏挂墙场景通常免登录：未携带 token 时不发请求（避免网关 401 →
 *   全局拦截器踢回 /login 导致大屏跳走），直接以本地 ApiError 失败，
 *   面板层捕获后走优雅占位（骨架 / '--' / 标注的演示数据）
 *
 * 类型全部复用各域 api 模块的既有定义，此处不做二次建模。
 */
import { get, cleanParams, ApiError } from './http'
import { useAuthStore } from '@/stores/auth'
import type { ApiListResp, CallLogListResp, CallStatsResp } from './datasvc'
import type { ProfileListResp, QualityRuleListResp } from './quality'
import type { DataSourceListResp } from './datasource'
import type { AssetListResp } from './catalog'
import type { TagDefListResp } from './tag'
import type { ListResp, StdWordRoot } from './standard'

/** 未登录（大屏演示模式）：不发请求，直接失败 → 面板占位，不触发 401 跳转 */
function requireToken(): void {
  if (!useAuthStore().token) {
    throw new ApiError(-1, '未登录 · 大屏演示模式（后端数据链路未接入）')
  }
}

/** 近 N 日调用统计：总量 / 成功率 / 日趋势 / Top API（30s 轮询主数据源） */
export function bsCallStats(days = 7): Promise<CallStatsResp> {
  requireToken()
  return get<CallStatsResp>('/datasvc/callstats', cleanParams({ days }))
}

/** 数据服务调用日志（大屏滚动列表，最新 N 条） */
export function bsCallLogs(size = 10): Promise<CallLogListResp> {
  requireToken()
  return get<CallLogListResp>('/datasvc/calllog/list', cleanParams({ page: 1, size }))
}

/** 表健康画像列表（环形仪表取最新 1 条；Top5 表格取健康分前 5） */
export function bsProfiles(size = 50): Promise<ProfileListResp> {
  requireToken()
  return get<ProfileListResp>('/quality/profile/list', cleanParams({ page: 1, size }))
}

/** 数据源总数（beid=1 与管理台口径一致） */
export function bsDataSources(): Promise<DataSourceListResp> {
  requireToken()
  return get<DataSourceListResp>('/datasource/list', cleanParams({ beid: 1, page: 1, size: 1 }))
}

/** 数据服务 API 总数 */
export function bsApis(): Promise<ApiListResp> {
  requireToken()
  return get<ApiListResp>('/datasvc/api/list', cleanParams({ beid: 1, page: 1, size: 1 }))
}

/** 质量规则列表（规则 6 类型分布按 rule_type 聚合） */
export function bsQualityRules(): Promise<QualityRuleListResp> {
  requireToken()
  return get<QualityRuleListResp>('/quality/rule/list', cleanParams({ beid: 1, page: 1, size: 100 }))
}

/** 资产总数（资产目录；beid 与管理台口径一致，缺省会被后端拒） */
export function bsAssetTotal(): Promise<AssetListResp> {
  requireToken()
  return get<AssetListResp>('/catalog/asset/list', cleanParams({ beid: 1, page: 1, size: 1 }))
}

/** 标签总数（标签定义） */
export function bsTagTotal(): Promise<TagDefListResp> {
  requireToken()
  return get<TagDefListResp>('/tag/def/list', cleanParams({ beid: 1, page: 1, size: 1 }))
}

/** 数据标准总数（词根） */
export function bsStdTotal(): Promise<ListResp<StdWordRoot>> {
  requireToken()
  return get<ListResp<StdWordRoot>>('/standard/wordroot/list', cleanParams({ beid: 1, page: 1, size: 1 }))
}
