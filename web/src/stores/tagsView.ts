import { defineStore } from 'pinia'
import type { RouteLocationNormalized } from 'vue-router'

/** 标签页（tags-view）条目 */
export interface TagView {
  /** 路由 name（keep-alive include 消费） */
  name?: string
  path: string
  fullPath: string
  title: string
  /** 固定标签（如仪表盘），不可关闭 */
  affix?: boolean
  /** 不参与 keep-alive 缓存 */
  noCache?: boolean
}

function toTagView(route: RouteLocationNormalized): TagView {
  return {
    name: typeof route.name === 'string' ? route.name : undefined,
    path: route.path,
    fullPath: route.fullPath,
    title: (route.meta.title as string) || route.path,
    affix: route.meta.affix === true,
    noCache: route.meta.noCache === true,
  }
}

/**
 * 标签页状态：router.afterEach 记录 visitedViews，TagsView.vue 渲染
 */
export const useTagsViewStore = defineStore('tagsView', {
  state: () => ({
    visitedViews: [] as TagView[],
  }),
  getters: {
    /** keep-alive include 名单（要求页面组件 name 与路由 name 一致，见 MainLayout TODO） */
    cachedViews: (state): string[] =>
      state.visitedViews.filter((v) => v.name && !v.noCache).map((v) => v.name as string),
  },
  actions: {
    /** router.afterEach 调用：登录页/404 不进标签 */
    addView(route: RouteLocationNormalized) {
      if (route.path === '/login' || route.path === '/404') return
      if (this.visitedViews.some((v) => v.path === route.path)) return
      this.visitedViews.push(toTagView(route))
    },
    /** 关闭单个（affix 固定标签不可关），返回剩余标签 */
    delView(view: TagView): TagView[] {
      const target = this.visitedViews.find((v) => v.path === view.path)
      if (target && !target.affix) {
        this.visitedViews = this.visitedViews.filter((v) => v.path !== view.path)
      }
      return [...this.visitedViews]
    },
    /** 关闭其他（保留固定标签与当前），返回剩余标签 */
    delOthersViews(view: TagView): TagView[] {
      this.visitedViews = this.visitedViews.filter((v) => v.affix || v.path === view.path)
      return [...this.visitedViews]
    },
    /** 关闭全部（保留固定标签，登出也走这里清空会话标签），返回剩余标签 */
    delAllViews(): TagView[] {
      this.visitedViews = this.visitedViews.filter((v) => v.affix)
      return [...this.visitedViews]
    },
  },
})
