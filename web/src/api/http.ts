import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

/** 后端统一响应包装：{code:0,msg:"",data:...}，code!=0 即业务错误 */
export interface ApiWrapper<T = unknown> {
  code: number
  msg: string
  data: T
}

/** 业务/网络错误的统一异常 */
export class ApiError extends Error {
  code: number
  constructor(code: number, msg: string) {
    super(msg || '请求失败')
    this.name = 'ApiError'
    this.code = code
  }
}

const http = axios.create({
  // 网关已开 CORS，前端直连后端（不经 vite 代理）；可用 .env 的 VITE_API_BASE 覆盖
  baseURL: import.meta.env.VITE_API_BASE || 'http://localhost:8887',
  timeout: 30000,
})

// 请求拦截：附加 JWT 与防越权 uid 头
http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) config.headers.Authorization = `Bearer ${auth.token}`
  if (auth.uid) config.headers['x-user-id'] = String(auth.uid)
  return config
})

// 响应拦截：剥包装、判业务码；401 清登录态回登录页
// 注：拦截器把 AxiosResponse 替换成了业务 data，调用方拿到的直接是 data 泛型
http.interceptors.response.use(
  (resp) => {
    const body = resp.data as ApiWrapper
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code !== 0) {
        return Promise.reject(new ApiError(body.code, body.msg))
      }
      return body.data as unknown as typeof resp
    }
    return body as unknown as typeof resp
  },
  (error: AxiosError<ApiWrapper>) => {
    if (error.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      if (router.currentRoute.value.path !== '/login') {
        router.push('/login')
      }
    }
    const msg = error.response?.data?.msg || error.message || '网络异常'
    return Promise.reject(new ApiError(error.response?.status ?? -1, msg))
  },
)

/** GET：params 走 query string（对应 .api 中 form 标签） */
export function get<T>(
  url: string,
  params?: Record<string, unknown>,
  config?: AxiosRequestConfig,
): Promise<T> {
  return http.get(url, { ...config, params }) as Promise<T>
}

/** POST：body 为 JSON */
export function post<T>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
): Promise<T> {
  return http.post(url, data ?? {}, config) as Promise<T>
}

/** 清洗请求参数：去掉 undefined/null/空串，避免 go-zero 解析出脏零值 */
export function cleanParams(params?: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  if (!params) return out
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== '') out[k] = v
  }
  return out
}

export default http
