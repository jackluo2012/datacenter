import { get, post, cleanParams } from './http'

/** 对齐 catalog.api（Phase 1 数据资产目录） */
export interface AssetCategory {
  id?: number
  parent_id?: number
  name?: string
}

export interface Asset {
  id?: number
  entity_id?: number
  category_id?: number
  name?: string
  entity_type?: string
  comment?: string
  classification?: string
  asset_tags_json?: string
  /** public/private */
  visibility?: string
}

export interface AssetListResp {
  list: Asset[]
  total: number
}

export function createCategory(req: {
  parent_id?: number
  name: string
}): Promise<AssetCategory> {
  return post<AssetCategory>('/catalog/category/create', req)
}

export function categoryList(): Promise<{ list: AssetCategory[] }> {
  return get<{ list: AssetCategory[] }>('/catalog/category/list')
}

/** 登记资产（挂分类、打资产标签） */
export function tagAsset(req: {
  entity_id: number
  category_id?: number
  name: string
  entity_type?: string
  comment?: string
  classification?: string
  asset_tags_json?: string
  visibility?: string
}): Promise<Asset> {
  return post<Asset>('/catalog/asset/tag', req)
}

export function assetList(params: {
  category_id?: number
  keyword?: string
  page?: number
  size?: number
}): Promise<AssetListResp> {
  return get<AssetListResp>('/catalog/asset/list', cleanParams(params))
}
