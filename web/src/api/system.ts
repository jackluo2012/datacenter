import { get, post, cleanParams } from './http'

/**
 * 系统管理 RBAC API（对齐根目录 sys.api / sys-rpc 契约）
 * - web/src/api/sys.ts 保留 getInfo/getRouters 登录链路（stores 引用，勿动）
 * - 用户/角色/菜单管理 CRUD 以本文件为准：字段 role_key/role_name/menu_type/perm 等
 *   与 sys.proto 一致；GET 走 query（form 标签），POST 走 JSON body
 */

// ============ 类型 ============
export interface SysUserInfo {
  id?: number
  /** 登录名 */
  username?: string
  nickname?: string
  mobile?: string
  email?: string
  /** 1 启用 0 停用 */
  status?: number
  remark?: string
  /** 毫秒时间戳 */
  created_at?: number
  /** 用户角色 id 集 */
  role_ids?: number[]
}

export type SysUserListParams = {
  /** username/nickname/mobile 模糊匹配 */
  keyword?: string
  /** -1 全部 / 1 启用 / 0 停用（后端 >=0 才过滤） */
  status?: number
  page?: number
  size?: number
}

export type SysUserCreateReq = {
  username: string
  password: string
  nickname?: string
  mobile?: string
  email?: string
  remark?: string
  role_ids?: number[]
}

export type SysUserUpdateReq = {
  id: number
  nickname?: string
  mobile?: string
  email?: string
  status?: number
  remark?: string
  /** 空 = 不修改密码 */
  password?: string
  role_ids?: number[]
}

export interface SysUserListResp {
  list: SysUserInfo[]
  total: number
}

export interface SysUserResp {
  user: SysUserInfo
}

export interface SysRoleInfo {
  id?: number
  /** 权限字符（如 admin/operator） */
  role_key?: string
  role_name?: string
  /** 1 启用 0 停用 */
  status?: number
  remark?: string
  /** 角色菜单 id 集（详情接口返回） */
  menu_ids?: number[]
}

export type SysRoleSaveReq = {
  /** 0 = 新建 */
  id?: number
  role_key: string
  role_name: string
  status?: number
  remark?: string
  menu_ids?: number[]
}

export interface SysRoleListResp {
  list: SysRoleInfo[]
  total: number
}

export interface SysRoleResp {
  role: SysRoleInfo
}

/** 1 目录 / 2 菜单 / 3 按钮 */
export interface SysMenuInfo {
  id?: number
  /** 0 根 */
  parent_id?: number
  name?: string
  /** 前端路由路径（目录如 /system，菜单如 user） */
  path?: string
  /** 视图组件路径（如 system/UserList → web/src/views/system/UserList.vue） */
  component?: string
  sort?: number
  /** element 图标名 */
  icon?: string
  menu_type?: number
  /** 按钮权限标识（如 sys:user:create） */
  perm?: string
  /** 1 显示 0 隐藏 */
  visible?: number
  children?: SysMenuInfo[]
}

export type SysMenuSaveReq = {
  /** 0 = 新建；更新为整存（未传字段会被清空） */
  id?: number
  parent_id?: number
  name: string
  path?: string
  component?: string
  sort?: number
  icon?: string
  menu_type: number
  perm?: string
  visible?: number
}

export interface SysMenuResp {
  menu: SysMenuInfo
}

export interface SysMenuTreeResp {
  list: SysMenuInfo[]
}

// ============ 用户管理 ============
/** status 请传 -1 表示不过滤（后端 >=0 才参与过滤） */
export function listUsers(params?: SysUserListParams) {
  return get<SysUserListResp>('/sys/user/list', cleanParams(params as Record<string, unknown>))
}

export function getUser(id: number) {
  return get<SysUserResp>('/sys/user/info', cleanParams({ id }))
}

export function createUser(req: SysUserCreateReq) {
  return post<SysUserResp>('/sys/user/create', req)
}

export function updateUser(req: SysUserUpdateReq) {
  return post<SysUserResp>('/sys/user/update', req)
}

export function deleteUser(id: number) {
  return post<void>('/sys/user/delete', { id })
}

/** 管理员重置用户密码 */
export function resetUserPwd(id: number, password: string) {
  return post<void>('/sys/user/resetpwd', { id, password })
}

// ============ 角色管理 ============
export function listRoles(params?: { keyword?: string; page?: number; size?: number }) {
  return get<SysRoleListResp>('/sys/role/list', cleanParams(params))
}

/** 角色详情（含 menu_ids） */
export function getRole(id: number) {
  return get<SysRoleResp>('/sys/role/info', cleanParams({ id }))
}

export function createRole(req: SysRoleSaveReq) {
  return post<SysRoleResp>('/sys/role/create', req)
}

export function updateRole(req: SysRoleSaveReq) {
  return post<SysRoleResp>('/sys/role/update', req)
}

export function deleteRole(id: number) {
  return post<void>('/sys/role/delete', { id })
}

// ============ 菜单管理 ============
/** 全量菜单树（目录/菜单/按钮，管理用） */
export function menuTree() {
  return get<SysMenuTreeResp>('/sys/menu/tree')
}

/** 新建/保存菜单（id=0 新建） */
export function saveMenu(req: SysMenuSaveReq) {
  return post<SysMenuResp>('/sys/menu/save', req)
}

export function deleteMenu(id: number) {
  return post<void>('/sys/menu/delete', { id })
}
