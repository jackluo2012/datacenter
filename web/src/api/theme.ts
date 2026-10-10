/**
 * 主题数仓（theme，Phase 6 骨架）—— 目前只有 gRPC 服务（theme/rpc/theme.proto），
 * 管理面 HTTP 接口尚未在网关暴露，先占位。
 * 待网关追加 theme.api 后，在此按 .api 文件对齐路由与类型。
 */
export interface ThemeBoard {
  id?: number
  code?: string
  name?: string
  comment?: string
}

/** 占位：主题看板列表（接口未上线） */
export async function themeBoardList(): Promise<ThemeBoard[]> {
  console.warn('[theme] 管理面 HTTP 接口未上线，返回空列表')
  return []
}
