import { get, post, cleanParams } from './http'

/**
 * 对齐 security.api（Phase 6 数据安全）。
 * 响应为裸对象：sys/security 网关 handler 直接 httpx.OkJsonCtx 输出，无 {code,msg,data} 包装，
 * http.ts 拦截器对无 code 字段 body 原样透传，故此处类型即后端返回结构。
 *
 * 租户口径：security rpc 全部 logic 要求 beid>0（缺省直接拒），管理台固定 beid=1
 * （与 bigscreen/catalog 等模块口径一致），在此统一注入，调用方无需关心。
 */

/** 管理台统一租户 ID */
const BEID = 1

// ---------- 类型 ----------

/** 分级目录条目（内置 L1 公开 / L2 内部 / L3 敏感 / L4 机密，L5-L9 租户自定义） */
export interface SecLevelInfo {
  id?: number
  code?: string
  name?: string
  comment?: string
}

export interface SecLevelListResp {
  list: SecLevelInfo[]
}

/** 实体打标条目（同实体重复打标等价覆盖） */
export interface SecEntityTag {
  id?: number
  entity_id?: number
  entity_name?: string
  level_code?: string
  mask_rule_code?: string
}

/** 脱敏规则（algorithm ∈ mask|hash|replace|null） */
export interface SecMaskRule {
  id?: number
  rule_code?: string
  rule_name?: string
  algorithm?: string
  /** JSON 串：mask→{prefix,suffix,mask_char?}，replace→{replace}，hash/null 无参数 */
  params_json?: string
}

/** 授权条目（subject 格式 user:1 / app:key / role:admin；expire_at=0 表示永久） */
export interface SecGrant {
  id?: number
  subject?: string
  resource?: string
  resource_id?: string
  level_code?: string
  expire_at?: number
}

/** 脱敏预览结果（rule_code 便于前端核对命中规则） */
export interface MaskPreviewResp {
  masked: string
  rule_code?: string
}

/** 权限检查结果（数据密级 ≤ 许可密级才放行） */
export interface SecCheckPermResp {
  allowed: boolean
  reason?: string
}

export interface SecListResp<T> {
  list: T[]
  total: number
}

// ---------- 分级目录 ----------

/** 新建/编辑分级（按 code upsert）；code 须匹配 ^L[1-9]$ */
export function saveSecLevel(req: {
  code: string
  name: string
  comment?: string
}): Promise<SecLevelInfo> {
  return post<SecLevelInfo>('/security/level/save', { ...req, beid: BEID })
}

/** 分级目录全量列表；目录为空时后端自动播种内置四级 */
export function secLevelList(): Promise<SecLevelListResp> {
  return get<SecLevelListResp>('/security/level/list', cleanParams({ beid: BEID }))
}

// ---------- 实体打标 ----------

/** 实体打标（uk_beid_entity upsert）；level_code 须在分级目录，mask_rule_code 非空时规则须存在 */
export function tagSecEntity(req: {
  entity_id: number
  entity_name?: string
  level_code: string
  mask_rule_code?: string
}): Promise<SecEntityTag> {
  return post<SecEntityTag>('/security/entity/tag', { ...req, beid: BEID })
}

/** 打标列表（level_code 过滤 + 分页） */
export function secEntityTagList(params: {
  level_code?: string
  page?: number
  size?: number
}): Promise<SecListResp<SecEntityTag>> {
  return get<SecListResp<SecEntityTag>>(
    '/security/entity/tag/list',
    cleanParams({ beid: BEID, ...params }),
  )
}

// ---------- 脱敏规则 ----------

/** 新建/编辑脱敏规则（按 code upsert）；params_json 由后端按算法校验（mask/replace 必填） */
export function saveSecMaskRule(req: {
  rule_code: string
  rule_name: string
  algorithm: string
  params_json?: string
}): Promise<SecMaskRule> {
  return post<SecMaskRule>('/security/maskrule/save', { ...req, beid: BEID })
}

/** 删除规则：被实体打标引用时后端拒绝 */
export function deleteSecMaskRule(id: number): Promise<void> {
  return post<void>('/security/maskrule/delete', { id, beid: BEID })
}

export function secMaskRuleList(params: {
  page?: number
  size?: number
}): Promise<SecListResp<SecMaskRule>> {
  return get<SecListResp<SecMaskRule>>(
    '/security/maskrule/list',
    cleanParams({ beid: BEID, ...params }),
  )
}

/** 脱敏预览：按规则主键查询（不做租户过滤），对样例值执行算法返回 masked */
export function maskPreview(params: { rule_id: number; sample: string }): Promise<MaskPreviewResp> {
  return get<MaskPreviewResp>('/security/mask/preview', cleanParams(params))
}

// ---------- 授权矩阵 ----------

/** 新建/编辑授权；subject 须为 type:id（type ∈ user/app/role） */
export function saveSecGrant(req: {
  subject: string
  resource: string
  resource_id?: string
  level_code: string
  expire_at?: number
}): Promise<SecGrant> {
  return post<SecGrant>('/security/grant/save', { ...req, beid: BEID })
}

export function deleteSecGrant(id: number): Promise<void> {
  return post<void>('/security/grant/delete', { id, beid: BEID })
}

export function secGrantList(params: {
  subject?: string
  resource?: string
  page?: number
  size?: number
}): Promise<SecListResp<SecGrant>> {
  return get<SecListResp<SecGrant>>(
    '/security/grant/list',
    cleanParams({ beid: BEID, ...params }),
  )
}

// ---------- 权限检查 ----------

/** 权限检查：任一参数缺省时按该维度不限制处理（后端 DecidePerm） */
export function secCheckPerm(params: {
  subject?: string
  resource?: string
  resource_id?: string
  level_code?: string
}): Promise<SecCheckPermResp> {
  return get<SecCheckPermResp>(
    '/security/checkperm',
    cleanParams({ beid: BEID, ...params }),
  )
}
