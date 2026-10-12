/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}

interface ImportMetaEnv {
  /** 后端网关地址，默认 http://localhost:8887 */
  readonly VITE_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

interface ImportMetaEnv {
  /** DataEase 仪表板地址（数据可视化页 iframe 嵌入用），默认 http://localhost:8088 */
  readonly VITE_DATAEASE_URL?: string
  /** 后端网关地址，默认 http://localhost:8887 */
  readonly VITE_API_BASE?: string
}
