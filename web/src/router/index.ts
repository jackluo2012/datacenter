import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from '@/layouts/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
  {
    path: '/',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' },
      },
      // ---- 完整功能页 ----
      {
        path: 'datasource',
        name: 'datasource',
        component: () => import('@/views/datasource/DataSourceList.vue'),
        meta: { title: '数据源' },
      },
      {
        path: 'metadata',
        name: 'metadata',
        component: () => import('@/views/metadata/MetadataList.vue'),
        meta: { title: '元数据' },
      },
      {
        path: 'ai',
        name: 'ai',
        component: () => import('@/views/ai/AiChat.vue'),
        meta: { title: 'AI 问数' },
      },
      // ---- 占位页（后续批次填充）----
      {
        path: 'datasvc',
        name: 'datasvc',
        component: () => import('@/views/datasvc/List.vue'),
        meta: { title: '数据服务' },
      },
      {
        path: 'catalog',
        name: 'catalog',
        component: () => import('@/views/catalog/List.vue'),
        meta: { title: '资产目录' },
      },
      {
        path: 'quality',
        name: 'quality',
        component: () => import('@/views/quality/List.vue'),
        meta: { title: '数据质量' },
      },
      {
        path: 'metric',
        name: 'metric',
        component: () => import('@/views/metric/List.vue'),
        meta: { title: '数据指标' },
      },
      {
        path: 'tag',
        name: 'tag',
        component: () => import('@/views/tag/List.vue'),
        meta: { title: '数据标签' },
      },
      {
        path: 'standard',
        name: 'standard',
        component: () => import('@/views/standard/List.vue'),
        meta: { title: '数据标准' },
      },
      {
        path: 'mdm',
        name: 'mdm',
        component: () => import('@/views/mdm/List.vue'),
        meta: { title: '主数据' },
      },
      {
        path: 'job',
        name: 'job',
        component: () => import('@/views/job/List.vue'),
        meta: { title: '调度集成' },
      },
      {
        path: 'ide',
        name: 'ide',
        component: () => import('@/views/ide/List.vue'),
        meta: { title: 'SQL IDE' },
      },
      {
        path: 'ops',
        name: 'ops',
        component: () => import('@/views/ops/List.vue'),
        meta: { title: '数据运维' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 路由守卫：无 token 一律回登录页
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path !== '/login' && !auth.isLogin) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.path === '/login' && auth.isLogin) {
    return { path: '/' }
  }
  document.title = to.meta.title ? `${to.meta.title} - 数据中台` : '数据中台'
  return true
})

export default router
