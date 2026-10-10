import { get, post, cleanParams } from './http'

/** 对齐 metric.api（Phase 2 数据指标） */
export interface MetricInfo {
  id?: number
  name_en?: string
  name_cn?: string
  /** atom/derive/composite */
  metric_type?: string
  biz_field?: string
  /** sum/count/avg ... */
  agg_type?: string
  expr?: string
  dimension_json?: string
  comment?: string
  owner?: string
  status?: number
}

export interface MetricListResp {
  list: MetricInfo[]
  total: number
}

export function createMetric(req: {
  name_en: string
  name_cn: string
  metric_type: string
  biz_field?: string
  agg_type?: string
  expr?: string
  dimension_json?: string
  comment?: string
  owner?: string
}): Promise<MetricInfo> {
  return post<MetricInfo>('/metric/create', req)
}

export function updateMetric(req: {
  id: number
  name_cn?: string
  expr?: string
  dimension_json?: string
  comment?: string
  status?: number
}): Promise<void> {
  return post<void>('/metric/update', req)
}

export function deleteMetric(id: number): Promise<void> {
  return post<void>('/metric/delete', { id })
}

export function metricInfo(id: number): Promise<MetricInfo> {
  return get<MetricInfo>('/metric/info', cleanParams({ id }))
}

export function metricList(params: {
  metric_type?: string
  keyword?: string
  page?: number
  size?: number
}): Promise<MetricListResp> {
  return get<MetricListResp>('/metric/list', cleanParams(params))
}

/** 派生指标绑定原子指标 */
export function bindAtoms(req: {
  derive_metric_id: number
  atom_metric_ids: number[]
}): Promise<void> {
  return post<void>('/metric/bind/atoms', req)
}

/** 派生指标引用的原子指标 */
export function metricAtoms(deriveMetricId: number): Promise<MetricListResp> {
  return get<MetricListResp>('/metric/atoms', cleanParams({ id: deriveMetricId }))
}
