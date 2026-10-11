<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { usePermissionStore, type MenuNode } from '@/stores/permission'

defineOptions({ name: 'AppBreadcrumb' })

interface Crumb {
  title: string
  path?: string
}

const route = useRoute()
const permission = usePermissionStore()

/**
 * 面包屑层级：优先从菜单树还原（目录节点不注册路由，route.matched 拿不到父级），
 * 菜单树覆盖不到时（如降级静态菜单）退回 route.matched 的 meta.title。
 */
const crumbs = computed<Crumb[]>(() => {
  const out: Crumb[] = []
  const walk = (nodes: MenuNode[], parents: Crumb[]): boolean => {
    for (const node of nodes) {
      if (node.hidden) continue
      const cur: Crumb[] = [...parents, { title: node.title, path: node.path }]
      if (node.path === route.path) {
        out.push(...cur)
        return true
      }
      if (node.children?.length && walk(node.children, cur)) return true
    }
    return false
  }
  walk(permission.menus, [])

  if (!out.length) {
    for (const m of route.matched) {
      const title = m.meta?.title as string | undefined
      if (title) out.push({ title, path: m.path })
    }
  }
  return out
})
</script>

<template>
  <el-breadcrumb separator="/" class="breadcrumb">
    <el-breadcrumb-item
      v-for="(item, idx) in crumbs"
      :key="`${item.path ?? item.title}-${idx}`"
      :to="idx === crumbs.length - 1 ? undefined : item.path"
    >
      {{ item.title }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<style scoped>
.breadcrumb {
  margin-left: 16px;
  white-space: nowrap;
}
</style>
