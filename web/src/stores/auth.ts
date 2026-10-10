import { defineStore } from 'pinia'

const TOKEN_KEY = 'dc_token'
const UID_KEY = 'dc_uid'
const NAME_KEY = 'dc_username'

/**
 * 登录态：token + uid 持久化到 localStorage
 * - token    → Authorization: Bearer <token>
 * - uid      → x-user-id: <uid>（防越权校验头）
 */
export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    uid: Number(localStorage.getItem(UID_KEY) || 0) || 0,
    username: localStorage.getItem(NAME_KEY) || '',
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
    logout() {
      this.token = ''
      this.uid = 0
      this.username = ''
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(UID_KEY)
      localStorage.removeItem(NAME_KEY)
    },
  },
})
