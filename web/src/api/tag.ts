import { get, post, cleanParams } from './http'

/** 对齐 tag.api（Phase 2 数据标签） */
export interface TagObject {
  id?: number
  /** entity/user/item ... */
  object_type?: string
  name?: string
  ds_id?: number
  db_name?: string
  table_name?: string
  pk_column?: string
}

export interface TagObjectListResp {
  list: TagObject[]
  total: number
}

export interface TagDef {
  id?: number
  object_id?: number
  tag_key?: string
  tag_name?: string
  /** 基础/规则/枚举 ... */
  tag_type?: string
  rule_sql?: string
  refresh_cron?: string
  status?: number
}

export interface TagDefListResp {
  list: TagDef[]
  total: number
}

export interface TaggingResp {
  tagged_rows: number
  tag_id: number
}

export interface TagValueItem {
  tag_key?: string
  tag_name?: string
  tag_value?: string
  tagged_at?: number
}

export function createTagObject(req: {
  object_type: string
  name: string
  ds_id?: number
  db_name?: string
  table_name?: string
  pk_column?: string
}): Promise<TagObject> {
  return post<TagObject>('/tag/object/create', req)
}

export function tagObjectList(params: {
  object_type?: string
  page?: number
  size?: number
}): Promise<TagObjectListResp> {
  return get<TagObjectListResp>('/tag/object/list', cleanParams(params))
}

export function createTagDef(req: {
  object_id: number
  tag_key: string
  tag_name: string
  tag_type?: string
  rule_sql?: string
  refresh_cron?: string
}): Promise<TagDef> {
  return post<TagDef>('/tag/def/create', req)
}

export function tagDefList(params: {
  object_id?: number
  page?: number
  size?: number
}): Promise<TagDefListResp> {
  return get<TagDefListResp>('/tag/def/list', cleanParams(params))
}

export function deleteTagDef(id: number): Promise<void> {
  return post<void>('/tag/def/delete', { id })
}

/** 执行打标，返回打标行数 */
export function tagging(req: { tag_id?: number; object_id?: number }): Promise<TaggingResp> {
  return post<TaggingResp>('/tag/tagging', req)
}

/** 对象标签画像 */
export function objectTags(params: {
  object_id?: number
  object_pk?: string
}): Promise<{ tags: TagValueItem[] }> {
  return get<{ tags: TagValueItem[] }>('/tag/object/tags', cleanParams(params))
}
