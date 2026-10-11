import { defineStore } from 'pinia'
import { getInfo } from '@/api/sys'

const TOKEN_KEY = 'dc_token'
const UID_KEY = 'dc_uid'
const NAME_KEY = 'dc_username'

/**
 * 登录态：token + uid 持久化到 localStorage
 * - token    → Authorization: Bearer <token>
 * - uid      → x-user-id: <uid>（防越权校验头）
 * - nickname/roles/perms：会话级（登录后由路由守卫调 getInfo 注入，不落 localStorage）
 */
export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    uid: Number(localStorage.getItem(UID_KEY) || 0) || 0,
    username: localStorage.getItem(NAME_KEY) || '',
    /** 昵称（getinfo.user.nickname，MainLayout 头像旁展示） */
    nickname: '',
    /** 角色编码集合，如 ['admin']；admin 拥有全部按钮权限 */
    roles: [] as string[],
    /** 按钮权限标识集合，如 ['sys:user:create']；v-perm 指令消费 */
    perms: [] as string[],
  }),
  getters: {
    isLogin: (state) => !!state.token,
  },
  actions: {
    setAuth(token: string, uid: number, username = '') {
      this.token = token
      this.uid = uid || 0
      this.username = username
      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(UID_KEY, String(this.uid))
      localStorage.setItem(NAME_KEY, username)
    },
    /** 登录后守卫调用：GET /sys/getinfo → 注入昵称/角色/按钮权限 */
    async fetchInfo() {
      const resp = await getInfo()
      if (resp.user) {
        this.uid = resp.user.uid || this.uid
        this.username = resp.user.username || this.username
        this.nickname = resp.user.nickname || ''
      }
      this.roles = resp.roles || []
      this.perms = resp.perms || []
    },
    logout() {
      this.token = ''
      this.uid = 0
      this.username = ''
      this.nickname = ''
      this.roles = []
      this.perms = []
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(UID_KEY)
      localStorage.removeItem(NAME_KEY)
      // 联动清理动态路由与标签页（延迟引入，避免 stores ↔ router 的 ESM 循环依赖在初始化期展开）
      void import('@/stores/permission')
        .then(({ usePermissionStore }) => usePermissionStore().reset())
        .catch(() => undefined)
      void import('@/stores/tagsView')
        .then(({ useTagsViewStore }) => useTagsViewStore().delAllViews())
        .catch(() => undefined)
    },
  },
})
