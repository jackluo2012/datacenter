import { get, post, cleanParams } from './http'

/** 对齐 metadata.api（Phase 1 元数据管理） */
export interface MetaEntity {
  id?: number
  /** table/column 等 */
  entity_type?: string
  fqn?: string
  name?: string
  comment?: string
  engine?: string
  rows_cnt?: number
  props_json?: string
  classification?: string
  owner?: string
  ds_id?: number
}

export interface MetaEntityListResp {
  list: MetaEntity[]
  total: number
}

export interface CollectResp {
  tables_created: number
  columns_created: number
}

export interface LineageNode {
  id?: number
  fqn?: string
  name?: string
}

export interface LineageEdge {
  from_id?: number
  to_id?: number
  props_json?: string
}

export interface LineageResp {
  nodes: LineageNode[]
  edges: LineageEdge[]
}

export interface MetaVersion {
  id?: number
  entity_id?: number
  version?: number
  snapshot_json?: string
  operator?: string
  published_at?: number
}

/** 采集库表元数据：dsn 为空则按 ds_id 从数据源服务取 */
export function collectDatabase(req: {
  ds_id: number
  db_name?: string
  db_type?: string
  dsn?: string
}): Promise<CollectResp> {
  return post<CollectResp>('/metadata/collect', req)
}

export function metaEntityInfo(id: number): Promise<MetaEntity> {
  return get<MetaEntity>('/metadata/entity/info', cleanParams({ id }))
}

export function metaEntityList(params: {
  entity_type?: string
  keyword?: string
  ds_id?: number
  page?: number
  size?: number
}): Promise<MetaEntityListResp> {
  return get<MetaEntityListResp>('/metadata/entity/list', cleanParams(params))
}

/** 登记血缘：from → to */
export function addLineage(req: {
  from_id: number
  to_id: number
  props_json?: string
}): Promise<void> {
  return post<void>('/metadata/lineage/add', req)
}

/** 血缘图 */
export function lineage(params: {
  entity_id: number
  direction?: string
  depth?: number
}): Promise<LineageResp> {
  return get<LineageResp>('/metadata/lineage', cleanParams(params))
}

/** 定版快照 */
export function createVersion(entityId: number): Promise<{ version: MetaVersion }> {
  return post<{ version: MetaVersion }>('/metadata/version/create', { id: entityId })
}

export function versionList(entityId: number): Promise<{ list: MetaVersion[] }> {
  return get<{ list: MetaVersion[] }>('/metadata/version/list', cleanParams({ id: entityId }))
}

/** 版本对比，返回 diff_json */
export function versionCompare(
  versionA: number,
  versionB: number,
): Promise<{ diff_json: string }> {
  return get<{ diff_json: string }>(
    '/metadata/version/compare',
    cleanParams({ version_a: versionA, version_b: versionB }),
  )
}
