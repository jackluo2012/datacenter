import { get, post, cleanParams } from './http'

/** 对齐 standard.api（Phase 5 数据标准管理） */
export interface StdWordRoot {
  id?: number
  name?: string
  cn_name?: string
  type?: string
  comment?: string
}

export interface StdDictItem {
  id?: number
  dict_id?: number
  item_code?: string
  item_value?: string
  sort?: number
}

export interface StdDict {
  id?: number
  dict_code?: string
  dict_name?: string
  comment?: string
  status?: number
  items: StdDictItem[]
}

export interface StdElement {
  id?: number
  element_code?: string
  element_name?: string
  data_type?: string
  data_length?: number
  data_precision?: number
  dict_code?: string
  comment?: string
}

export interface StdModel {
  id?: number
  model_code?: string
  model_name?: string
  elements_json?: string
  status?: string
  version?: number
  snapshot_json?: string
  published_at?: number
}

export interface StdCheckReport {
  id?: number
  model_id?: number
  table_name?: string
  diff_json?: string
  pass_flag: boolean
  run_at?: number
}

export interface ListResp<T> {
  list: T[]
  total: number
}

// ---------- 词根 ----------
export function saveWordRoot(req: {
  id?: number
  name: string
  cn_name?: string
  type?: string
  comment?: string
}): Promise<StdWordRoot> {
  return post<StdWordRoot>('/standard/wordroot/save', req)
}

export function deleteWordRoot(id: number): Promise<void> {
  return post<void>('/standard/wordroot/delete', { id })
}

export function wordRootList(params: {
  keyword?: string
  page?: number
  size?: number
}): Promise<ListResp<StdWordRoot>> {
  return get<ListResp<StdWordRoot>>('/standard/wordroot/list', cleanParams(params))
}

// ---------- 标准字典 ----------
export function saveDict(req: {
  id?: number
  dict_code: string
  dict_name: string
  comment?: string
  items: StdDictItem[]
}): Promise<StdDict> {
  return post<StdDict>('/standard/dict/save', req)
}

export function dictInfo(id: number): Promise<StdDict> {
  return get<StdDict>('/standard/dict/info', cleanParams({ id }))
}

export function deleteDict(id: number): Promise<void> {
  return post<void>('/standard/dict/delete', { id })
}

export function dictList(params: {
  keyword?: string
  page?: number
  size?: number
}): Promise<ListResp<StdDict>> {
  return get<ListResp<StdDict>>('/standard/dict/list', cleanParams(params))
}

// ---------- 数据元 ----------
export function saveElement(req: {
  id?: number
  element_code: string
  element_name: string
  data_type: string
  data_length?: number
  data_precision?: number
  dict_code?: string
  comment?: string
}): Promise<StdElement> {
  return post<StdElement>('/standard/element/save', req)
}

export function deleteElement(id: number): Promise<void> {
  return post<void>('/standard/element/delete', { id })
}

export function elementList(params: {
  keyword?: string
  page?: number
  size?: number
}): Promise<ListResp<StdElement>> {
  return get<ListResp<StdElement>>('/standard/element/list', cleanParams(params))
}

// ---------- 标准模型 ----------
export function saveStdModel(req: {
  id?: number
  model_code: string
  model_name: string
  elements_json: string
}): Promise<StdModel> {
  return post<StdModel>('/standard/model/save', req)
}

export function stdModelInfo(id: number): Promise<StdModel> {
  return get<StdModel>('/standard/model/info', cleanParams({ id }))
}

export function deleteStdModel(id: number): Promise<void> {
  return post<void>('/standard/model/delete', { id })
}

export function stdModelList(params: {
  status?: string
  keyword?: string
  page?: number
  size?: number
}): Promise<ListResp<StdModel>> {
  return get<ListResp<StdModel>>('/standard/model/list', cleanParams(params))
}

/** 发布标准模型，返回版本号与快照 */
export function publishStdModel(req: {
  id: number
  operator?: string
}): Promise<{ version: number; snapshot_json: string }> {
  return post('/standard/model/publish', req)
}

/** 落标检查：模型 vs 实际表结构 */
export function stdCheck(params: {
  model_id: number
  datasource_id?: number
  db_name?: string
  table_name?: string
}): Promise<{ report: StdCheckReport }> {
  return get<{ report: StdCheckReport }>('/standard/check', cleanParams(params))
}

export function stdCheckList(params: {
  model_id?: number
  page?: number
  size?: number
}): Promise<ListResp<StdCheckReport>> {
  return get<ListResp<StdCheckReport>>('/standard/check/list', cleanParams(params))
}
