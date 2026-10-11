import type { Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

/** 超级权限通配符（后端可在 perms 中下发 '*:*:*' 表示全部按钮权限） */
const SUPER_PERM = '*:*:*'

/**
 * 按钮权限判断：
 * - 角色 admin → 全量放行（对标 RuoYi 超管）
 * - perms 含 '*:*:*' → 全量放行
 * - value 为数组时命中任一即放行（or 语义）
 */
export function hasPerm(value: string | string[]): boolean {
  if (!value) return true
  const auth = useAuthStore()
  if (auth.roles.includes('admin')) return true
  if (auth.perms.includes(SUPER_PERM)) return true
  const need = Array.isArray(value) ? value : [value]
  if (!need.length) return true
  return need.some((p) => auth.perms.includes(p))
}

function check(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
  if (!binding.value) return
  if (!hasPerm(binding.value)) {
    // 无权限直接移除元素（与 RuoYi 一致；移除后不随权限变化恢复）
    el.parentNode?.removeChild(el)
  }
}

/**
 * v-perm="'sys:user:create'"        单个权限
 * v-perm="['sys:user:edit', ...]"   任一命中
 */
export const perm: Directive<HTMLElement, string | string[]> = {
  mounted: check,
}

export default perm
