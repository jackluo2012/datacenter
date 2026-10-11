import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePermissionStore, type MenuNode } from '@/stores/permission'
import { useTagsViewStore } from '@/stores/tagsView'
import MainLayout from '@/layouts/MainLayout.vue'
import NotFound from '@/components/NotFound.vue'

/**
 * 静态路由白名单：仅保留 /login、/404、/（MainLayout + dashboard）与兜底 404。
 * 其余业务路由由 stores/permission.ts 调 getRouters 动态 addRoute 注册；
 * dashboard 保留静态（菜单种子同样会下发，动态注册时按 name/路径 去重跳过）。
 */
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/404',
    name: 'not-found',
    component: NotFound,
    meta: { title: '404' },
  },
  {
    path: '/',
    name: 'Layout',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'Odometer', affix: true },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/404' },
]

/**
 * 降级路由数据（= 改造前的静态业务路由，含手写菜单的 title/icon）：
 * 平时【不】注册进路由表（常驻白名单只有 /login、/404、/ + dashboard），
 * 仅当 getRouters 失败（sys-rpc 未部署/网络异常）时由 permission.fallbackToStatic()
 * 调 registerFallbackRoutes() 注册，菜单走 staticMenuTree()，保证框架不白屏且功能入口齐全。
 */
export const fallbackChildren: RouteRecordRaw[] = [
  { path: 'dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '仪表盘', icon: 'Odometer' } },
  { path: 'datasource', name: 'datasource', component: () => import('@/views/datasource/DataSourceList.vue'), meta: { title: '数据源', icon: 'Coin' } },
  { path: 'metadata', name: 'metadata', component: () => import('@/views/metadata/MetadataList.vue'), meta: { title: '元数据', icon: 'Collection' } },
  { path: 'ai', name: 'ai', component: () => import('@/views/ai/AiChat.vue'), meta: { title: 'AI 问数', icon: 'ChatDotRound' } },
  { path: 'datasvc', name: 'datasvc', component: () => import('@/views/datasvc/List.vue'), meta: { title: '数据服务', icon: 'Connection' } },
  { path: 'catalog', name: 'catalog', component: () => import('@/views/catalog/List.vue'), meta: { title: '资产目录', icon: 'FolderOpened' } },
  { path: 'quality', name: 'quality', component: () => import('@/views/quality/List.vue'), meta: { title: '数据质量', icon: 'CircleCheck' } },
  { path: 'metric', name: 'metric', component: () => import('@/views/metric/List.vue'), meta: { title: '数据指标', icon: 'DataAnalysis' } },
  { path: 'tag', name: 'tag', component: () => import('@/views/tag/List.vue'), meta: { title: '数据标签', icon: 'PriceTag' } },
  { path: 'standard', name: 'standard', component: () => import('@/views/standard/List.vue'), meta: { title: '数据标准', icon: 'Document' } },
  { path: 'mdm', name: 'mdm', component: () => import('@/views/mdm/List.vue'), meta: { title: '主数据', icon: 'Box' } },
  { path: 'job', name: 'job', component: () => import('@/views/job/List.vue'), meta: { title: '调度集成', icon: 'Timer' } },
  { path: 'ide', name: 'ide', component: () => import('@/views/ide/List.vue'), meta: { title: 'SQL IDE', icon: 'EditPen' } },
  { path: 'ops', name: 'ops', component: () => import('@/views/ops/List.vue'), meta: { title: '数据运维', icon: 'Setting' } },
]

/** 降级菜单树：基于 fallbackChildren（meta.title/icon）生成 */
export function staticMenuTree(): MenuNode[] {
  return fallbackChildren.map((c) => ({
    name: String(c.name ?? ''),
    path: typeof c.path === 'string' && c.path.startsWith('/') ? c.path : `/${c.path}`,
    title: c.meta?.title as string,
    icon: c.meta?.icon as string | undefined,
    children: [],
  }))
}

/** 降级时把 fallbackChildren 注册为 Layout 子路由（幂等：同名/同路径已存在则跳过） */
export function registerFallbackRoutes(): void {
  for (const child of fallbackChildren) {
    const fullPath = typeof child.path === 'string' && child.path.startsWith('/') ? child.path : `/${child.path}`
    if (child.name && router.hasRoute(child.name)) continue
    if (router.getRoutes().some((r) => r.path === fullPath)) continue
    router.addRoute('Layout', { ...child, path: fullPath })
  }
}

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

/**
 * 路由守卫：
 * - 无 token → /login（携带 redirect）
 * - 已登录进 /login → 回首页
 * - 登录后首次导航：await getInfo + getRouters → addRoute → 重放当前导航（{...to, replace}）
 *   两步接口失败均降级放行（静态菜单兜底），不允许白屏
 */
router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!auth.isLogin) {
    if (to.path === '/login') return true
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/login') return { path: '/' }

  const permission = usePermissionStore()
  if (!permission.loaded) {
    try {
      await auth.fetchInfo()
    } catch (e) {
      // 降级：getInfo 失败不阻塞进入系统（roles/perms 为空 → v-perm 隐藏受控按钮）
      console.warn('[guard] getInfo 失败，已降级（无角色/按钮权限）', e)
    }
    await permission.generateRoutes() // 内部已兜底：getRouters 失败 → 静态菜单
    return { ...to, replace: true } // 重放导航，命中刚注册的动态路由
  }
  return true
})

/** 页面标题 + 标签页（tags-view）记录 */
router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - 数据中台` : '数据中台'
  useTagsViewStore().addView(to)
})

export default router
