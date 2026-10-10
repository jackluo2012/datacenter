import { get, post, cleanParams } from './http'

/** 对齐 mdm.api（Phase 5 主数据管理） */
export interface MdmModelInfo {
  id?: number
  model_code?: string
  model_name?: string
  table_name?: string
  fields_json?: string
  code_strategy?: string
}

export interface MdmModelListResp {
  list: MdmModelInfo[]
  total: number
}

export interface MdmRecordInfo {
  id?: number
  model_id?: number
  biz_code?: string
  data_json?: string
  /** draft/published ... */
  status?: string
  version?: number
}

export interface MdmRecordListResp {
  list: MdmRecordInfo[]
  total: number
}

export interface MdmDistribution {
  id?: number
  record_id?: number
  target_ds_id?: number
  target_db?: string
  target_table?: string
  status?: string
  msg?: string
  distributed_at?: number
}

export interface MdmDistListResp {
  list: MdmDistribution[]
  total: number
}

// ---------- 主数据模型 ----------
export function saveMdmModel(req: {
  id?: number
  model_code: string
  model_name: string
  table_name?: string
  fields_json: string
  code_strategy?: string
}): Promise<MdmModelInfo> {
  return post<MdmModelInfo>('/mdm/model/save', req)
}

export function mdmModelInfo(id: number): Promise<MdmModelInfo> {
  return get<MdmModelInfo>('/mdm/model/info', cleanParams({ id }))
}

export function deleteMdmModel(id: number): Promise<void> {
  return post<void>('/mdm/model/delete', { id })
}

export function mdmModelList(params: {
  keyword?: string
  page?: number
  size?: number
}): Promise<MdmModelListResp> {
  return get<MdmModelListResp>('/mdm/model/list', cleanParams(params))
}

/** 按模型编码策略取下一个业务编码 */
export function mdmNextCode(modelId: number): Promise<{ biz_code: string }> {
  return get<{ biz_code: string }>('/mdm/nextcode', cleanParams({ id: modelId }))
}

// ---------- 主数据记录 ----------
export function saveMdmRecord(req: {
  id?: number
  model_id: number
  biz_code?: string
  data_json: string
}): Promise<MdmRecordInfo> {
  return post<MdmRecordInfo>('/mdm/record/save', req)
}

export function mdmRecordInfo(id: number): Promise<MdmRecordInfo> {
  return get<MdmRecordInfo>('/mdm/record/info', cleanParams({ id }))
}

export function deleteMdmRecord(id: number): Promise<void> {
  return post<void>('/mdm/record/delete', { id })
}

export function mdmRecordList(params: {
  model_id?: number
  keyword?: string
  page?: number
  size?: number
}): Promise<MdmRecordListResp> {
  return get<MdmRecordListResp>('/mdm/record/list', cleanParams(params))
}

/** 发布主数据记录 */
export function publishMdmRecord(id: number): Promise<MdmRecordInfo> {
  return post<MdmRecordInfo>('/mdm/record/publish', { id })
}

/** 分发主数据记录到目标库表 */
export function distributeMdmRecord(req: {
  record_id: number
  target_ds_id: number
  target_db?: string
  target_table?: string
}): Promise<{ distribution: MdmDistribution }> {
  return post('/mdm/record/distribute', req)
}

export function mdmDistList(params: {
  record_id?: number
  page?: number
  size?: number
}): Promise<MdmDistListResp> {
  return get<MdmDistListResp>('/mdm/distribution/list', cleanParams(params))
}
