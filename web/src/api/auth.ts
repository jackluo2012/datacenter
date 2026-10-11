import { get, post, cleanParams } from './http'

/** 对齐 user.api 的 UserReply（登录/注册/用户信息统一返回） */
export interface JwtToken {
  access_token?: string
  access_expire?: number
  refresh_after?: number
}

export interface UserReply {
  auid?: number
  uid: number
  beid?: number
  ptyid?: number
  username?: string
  mobile?: string
  nickname?: string
  openid?: string
  avator?: string
  jwt_token?: JwtToken
}

export interface LoginReq {
  mobile: string
  /** 1.密码登陆，2.短信登陆 */
  type: number
  password: string
}

/** 账号密码登录（type=1），返回 UserReply；调用方从 jwt_token.access_token 取 token */
export function login(req: { mobile: string; password: string }): Promise<UserReply> {
  const body: LoginReq = { mobile: req.mobile, type: 1, password: req.password }
  return post<UserReply>('/user/login', body)
}

/** 注册（演示环境备用） */
export function register(req: {
  mobile: string
  password: string
  smscode?: string
}): Promise<UserReply> {
  return post<UserReply>('/user/register', req)
}

/** 当前用户信息（JWT + Usercheck 中间件） */
export function userInfo(params?: { uid?: number; auid?: number }): Promise<UserReply> {
  return get<UserReply>('/user/dc/info', cleanParams(params))
}

// ============ 管理台登录（sys_user，RBAC） ============
export interface SysLoginResp {
  token: string
  uid: number
  username: string
  nickname?: string
  roles: string[]
  perms: string[]
}

// 管理台登录（走 sys_user 表与 RBAC 角色；响应为裸对象，无 code 包装）
export function sysLogin(req: { username: string; password: string }): Promise<SysLoginResp> {
  return post<SysLoginResp>('/sys/login', req)
}
