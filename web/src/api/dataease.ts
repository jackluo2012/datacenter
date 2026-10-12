import { post } from './http'

/**
 * 数据可视化（DataEase 集成）前端辅助：
 * 1. DataEase 地址的读取/保存（优先级：localStorage 覆盖 > VITE_DATAEASE_URL > 默认 8088）
 * 2. DataEase 连通性探测
 * 3. 平台数据 API（datasvc /datasvc/execute）签名执行自测 ——
 *    供 DataEase「API 数据源」对接前验证 app_key 签名链路
 */

const DATAEASE_URL_KEY = 'dataease_embed_url'

/** env.d.ts 的 ImportMetaEnv 未声明 VITE_DATAEASE_URL（新页面自包含、不改公共类型），宽化读取 */
function envDataeaseUrl(): string {
  const env = import.meta.env as unknown as Record<string, string | undefined>
  return (env.VITE_DATAEASE_URL ?? '').trim()
}

/** 构建期默认地址（VITE_DATAEASE_URL，未配置时用本地默认端口） */
export function defaultDataeaseUrl(): string {
  return envDataeaseUrl() || 'http://localhost:8088'
}

/** 读当前生效地址：localStorage 覆盖优先，其次构建期默认 */
export function loadDataeaseUrl(): string {
  try {
    return (localStorage.getItem(DATAEASE_URL_KEY) || defaultDataeaseUrl()).trim()
  } catch {
    return defaultDataeaseUrl()
  }
}

/** 保存覆盖地址（存 localStorage；传空串等价于清除覆盖） */
export function saveDataeaseUrl(url: string): void {
  try {
    const clean = url.trim()
    if (clean) localStorage.setItem(DATAEASE_URL_KEY, clean)
    else localStorage.removeItem(DATAEASE_URL_KEY)
  } catch {
    /* 隐身模式等 localStorage 不可用时静默：仅本次会话生效 */
  }
}

/** 清除覆盖地址，回退构建期默认 */
export function clearDataeaseUrl(): void {
  try {
    localStorage.removeItem(DATAEASE_URL_KEY)
  } catch {
    /* 同上 */
  }
}

/** 归一化：去尾部 /（拼接探测路径用） */
export function normalizeUrl(url: string): string {
  return url.trim().replace(/\/+$/, '')
}

/**
 * 连通性探测：no-cors 模式下响应为 opaque 读不到状态码，
 * 但 fetch resolve 即代表主机/端口可达，reject（含超时）即不可达
 */
export async function probeDataease(url: string, timeoutMs = 4000): Promise<boolean> {
  const target = normalizeUrl(url)
  if (!target) return false
  const ctrl = new AbortController()
  const timer = setTimeout(() => ctrl.abort(), timeoutMs)
  try {
    await fetch(`${target}/login`, { mode: 'no-cors', signal: ctrl.signal })
    return true
  } catch {
    return false
  } finally {
    clearTimeout(timer)
  }
}

/** 对齐 datasvc.api 的 ExecuteApiResp（网关唯一免 JWT 入口，返回体未包 {code,msg,data} 壳） */
export interface ExecuteApiResp {
  rows_json: string
  cost_ms: number
  cache_hit: boolean
  affected: number
}

/**
 * 计算执行签名：hex(hmac-sha256(key=app_secret, msg=app_key+ts))
 * 与 datasvc/rpc/internal/logic/credential.go 的 genSign 完全一致（走 WebCrypto）
 */
export async function datasvcSign(appKey: string, appSecret: string, ts: string): Promise<string> {
  const enc = new TextEncoder()
  const key = await crypto.subtle.importKey(
    'raw',
    enc.encode(appSecret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const mac = await crypto.subtle.sign('HMAC', key, enc.encode(appKey + ts))
  return Array.from(new Uint8Array(mac))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

/**
 * 签名执行平台数据 API（ts 为当前 unix 秒，服务端容差 ±300s；
 * crypto.subtle 需要 https 或 localhost 的安全上下文）
 */
export async function executePlatformApi(
  appKey: string,
  appSecret: string,
  path: string,
  paramsJson: string,
): Promise<ExecuteApiResp> {
  const ts = Math.floor(Date.now() / 1000).toString()
  const sign = await datasvcSign(appKey, appSecret, ts)
  return post<ExecuteApiResp>('/datasvc/execute', {
    app_key: appKey,
    ts,
    sign,
    path,
    params_json: paramsJson,
  })
}
