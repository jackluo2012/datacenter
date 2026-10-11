import { defineStore } from 'pinia'
import type { RouteComponent, RouteRecordRaw } from 'vue-router'
import router, { registerFallbackRoutes, staticMenuTree } from '@/router'
import { getRouters, type RouterNode } from '@/api/sys'
import NotFound from '@/components/NotFound.vue'

/** 侧边栏菜单节点（getRouters 树 → el-menu 渲染用） */
export interface MenuNode {
  /** 路由 name（后端 name；同时用作 keep-alive include 匹配） */
  name: string
  /** 完整路径：el-menu index / router.push 使用 */
  path: string
  /** 菜单标题 */
  title: string
  icon?: string
  /** 隐藏菜单：路由照常注册，仅不出现在侧边栏 */
  hidden?: boolean
  children?: MenuNode[]
}

/**
 * 预注册全部视图组件：vite 下动态 import 必须可静态分析，
 * 用 import.meta.glob 一次性收集 @/views 目录下所有 .vue，再按后端 component 字符串取用
 */
const viewModules = import.meta.glob('@/views/**/*.vue') as Record<
  string,
  () => Promise<RouteComponent>
>

/** 后端 component 字符串（如 system/user/index）→ glob 键（兼容不同 glob 键形式） */
function resolveView(component?: string): (() => Promise<RouteComponent>) | undefined {
  if (!component) return undefined
  const clean = component.replace(/^\/+/, '').replace(/\.vue$/, '')
  return viewModules[`/src/views/${clean}.vue`] ?? viewModules[`@/views/${clean}.vue`]
}

/** 拼接父子路径；子级为绝对路径时直接采用 */
function joinPath(parent: string, child: string): string {
  if (!child) return parent || '/'
  if (child.startsWith('/')) return child
  return `${parent.replace(/\/+$/, '')}/${child}`
}

function toMenu(node: RouterNode, path: string, hidden: boolean): MenuNode {
  return { name: node.name, path, title: node.name, icon: node.icon, hidden, children: [] }
}

/**
 * 递归转换 getRouters 节点：
 * - 目录（有 children）→ 只进菜单树，不注册路由
 * - 叶子 → 以完整路径 addRoute 挂到 Layout（MainLayout）下，并进菜单树
 * - 与静态路由同名（如 dashboard）或同路径已存在 → 跳过注册仅保留菜单（去重）
 * - component 解析失败 → 回退 404 页并 console.warn
 */
function transformNode(node: RouterNode, parentPath: string, menus: MenuNode[], added: string[]) {
  const fullPath = joinPath(parentPath, node.path)
  const hidden = node.hidden === true

  if (node.children?.length) {
    const childMenus: MenuNode[] = []
    for (const child of node.children) transformNode(child, fullPath, childMenus, added)
    if (childMenus.length) {
      menus.push({
        name: node.name,
        path: fullPath,
        title: node.name,
        icon: node.icon,
        hidden,
        children: childMenus,
      })
    }
    return
  }

  const duplicated =
    (!!node.name && router.hasRoute(node.name)) ||
    router.getRoutes().some((r) => r.path === fullPath)
  if (duplicated) {
    // 静态路由已存在（如 dashboard）：跳过 addRoute，菜单项指向既有静态路由
    menus.push(toMenu(node, fullPath, hidden))
    return
  }

  const view = resolveView(node.component)
  if (!view) {
    console.warn(
      `[permission] 组件解析失败: "${node.component ?? '(空)'}"（路由 ${fullPath}），回退 404 页`,
    )
  }

  router.addRoute('Layout', {
    path: fullPath,
    name: node.name || undefined,
    component: view || NotFound,
    meta: { title: node.name, icon: node.icon },
  } satisfies RouteRecordRaw)

  if (node.name) added.push(node.name)
  menus.push(toMenu(node, fullPath, hidden))
}

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    /** 动态路由/菜单是否已注入（登录后首次导航完成置 true，防重复拉取） */
    loaded: false,
    /** 侧边栏菜单树 */
    menus: [] as MenuNode[],
    /** 本会话动态注册的路由名（登出时移除） */
    addedRouteNames: [] as string[],
  }),
  actions: {
    /**
     * 拉取动态路由并注册（守卫首次导航调用）。
     * 降级：getRouters 失败或返回空 → 走 staticMenuTree()（静态菜单），绝不白屏
     */
    async generateRoutes() {
      try {
        const resp = await getRouters()
        const routers = resp?.routers ?? []
        const menus: MenuNode[] = []
        const added: string[] = []
        for (const node of routers) transformNode(node, '', menus, added)
        this.menus = menus
        this.addedRouteNames = added
        if (!menus.length) {
          console.warn('[permission] getRouters 返回空菜单，降级为静态菜单')
          this.fallbackToStatic()
        } else {
          console.info(`[permission] 动态路由注册完成：${added.length} 条，菜单 ${menus.length} 组`)
        }
      } catch (e) {
        console.warn('[permission] getRouters 失败，降级为静态菜单', e)
        this.fallbackToStatic()
      } finally {
        this.loaded = true
      }
    },
    /** 降级：注册内置回退路由（改造前的 14 条业务路由）并生成静态菜单，保证可用性 */
    fallbackToStatic() {
      registerFallbackRoutes()
      this.menus = staticMenuTree()
    },
    /** 登出清理：移除动态路由、清空菜单，下次登录重新注入 */
    reset() {
      for (const name of this.addedRouteNames) {
        if (router.hasRoute(name)) router.removeRoute(name)
      }
      this.addedRouteNames = []
      this.menus = []
      this.loaded = false
    },
  },
})
