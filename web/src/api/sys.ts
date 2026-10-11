import { get, post, cleanParams } from './http'

/**
 * sys RBAC 契约（后端 sys-rpc，对标 RuoYi 权限模型）
 * - GET /sys/getinfo    → 当前用户信息 + roles + perms + 菜单树（目录+菜单，不含按钮）
 * - GET /sys/getrouters → 前端动态路由树（component 为字符串，由前端映射到 @/views）
 * - 其余为系统管理 CRUD 预留：路径待 sys-rpc 定稿后对齐（go-zero 风格 list/create/update/delete）
 */

/** 菜单/目录（getinfo.menus 节点；按钮类型不下发到前端） */
export interface SysMenu {
  id: number
  parent_id: number
  /** 菜单/目录名称（getrouters 中同时用作路由 name 与侧边栏标题） */
  name: string
  /** 路由路径：目录如 /system，菜单如 user（相对，前端拼完整路径） */
  path: string
  /** 视图组件：相对 src/views 的路径（不含 .vue），如 system/user/index */
  component?: string
  sort?: number
  icon?: string
  /** 是否显示（false = 仅注册路由，不进侧边栏） */
  visible?: boolean
  children?: SysMenu[]
}

export interface SysUserInfo {
  uid: number
  username: string
  nickname?: string
}

export interface GetInfoResp {
  user: SysUserInfo
  roles: string[]
  /** 按钮权限标识集合，如 sys:user:create */
  perms: string[]
  /** 目录+菜单树（不含按钮） */
  menus?: SysMenu[]
}

/** getrouters 返回的动态路由节点 */
export interface RouterNode {
  name: string
  path: string
  component?: string
  icon?: string
  /** 隐藏：路由照常注册，仅不出现在侧边栏 */
  hidden?: boolean
  children?: RouterNode[]
}

export interface GetRoutersResp {
  routers: RouterNode[]
}

/** 当前登录用户信息（JWT 鉴权） */
export function getInfo(): Promise<GetInfoResp> {
  return get<GetInfoResp>('/sys/getinfo')
}

/** 前端动态路由（permission store：addRoute + 菜单树） */
export function getRouters(): Promise<GetRoutersResp> {
  return get<GetRoutersResp>('/sys/getrouters')
}

/* ==================== 系统管理 CRUD（预留，sys-rpc 就绪后对齐路径/字段） ==================== */

export interface SysUser {
  uid?: number
  username?: string
  nickname?: string
  mobile?: string
  role_ids?: number[]
  status?: number
  created_at?: number
}

export interface SysRole {
  id?: number
  name?: string
  /** 按钮权限标识集合 */
  perms?: string[]
  remark?: string
  status?: number
}

export interface SysMenuSaveReq {
  id?: number
  parent_id?: number
  /** 1 目录 / 2 菜单 / 3 按钮 */
  type?: number
  name?: string
  path?: string
  component?: string
  sort?: number
  icon?: string
  visible?: boolean
  /** 按钮权限标识，如 sys:user:create */
  perm?: string
}

export interface PageResp<T> {
  list: T[]
  total: number
}

// ---- 用户管理 ----
export function listUsers(params?: { page?: number; page_size?: number; keyword?: string }) {
  return get<PageResp<SysUser>>('/sys/user/list', cleanParams(params))
}
export function createUser(req: SysUser) {
  return post<SysUser>('/sys/user/create', req)
}
export function updateUser(req: SysUser) {
  return post<SysUser>('/sys/user/update', req)
}
export function deleteUser(req: { uid: number }) {
  return post<void>('/sys/user/delete', req)
}

// ---- 角色管理 ----
export function listRoles(params?: { page?: number; page_size?: number; keyword?: string }) {
  return get<PageResp<SysRole>>('/sys/role/list', cleanParams(params))
}
export function createRole(req: SysRole) {
  return post<SysRole>('/sys/role/create', req)
}
export function updateRole(req: SysRole) {
  return post<SysRole>('/sys/role/update', req)
}
export function deleteRole(req: { id: number }) {
  return post<void>('/sys/role/delete', req)
}

// ---- 菜单管理 ----
export function listMenus(params?: { keyword?: string }) {
  return get<{ list: SysMenu[] }>('/sys/menu/list', cleanParams(params))
}
export function createMenu(req: SysMenuSaveReq) {
  return post<SysMenuSaveReq>('/sys/menu/create', req)
}
export function updateMenu(req: SysMenuSaveReq) {
  return post<SysMenuSaveReq>('/sys/menu/update', req)
}
export function deleteMenu(req: { id: number }) {
  return post<void>('/sys/menu/delete', req)
}
