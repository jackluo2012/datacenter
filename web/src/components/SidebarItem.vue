<script setup lang="ts">
import { computed } from 'vue'
import type { MenuNode } from '@/stores/permission'

defineOptions({ name: 'SidebarItem' })

const props = defineProps<{ item: MenuNode }>()

/** 只渲染 visible 子菜单（hidden 的仅注册了路由，不显示） */
const visibleChildren = computed(() => (props.item.children || []).filter((c) => !c.hidden))
</script>

<template>
  <!-- 叶子菜单（无子级）：index 即路由完整路径，配合 el-menu 的 router 模式跳转 -->
  <el-menu-item v-if="!visibleChildren.length" :index="item.path">
    <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
    <template #title>{{ item.title }}</template>
  </el-menu-item>

  <!-- 目录：递归渲染子级 -->
  <el-sub-menu v-else :index="item.path">
    <template #title>
      <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
      <span>{{ item.title }}</span>
    </template>
    <SidebarItem v-for="child in visibleChildren" :key="child.path" :item="child" />
  </el-sub-menu>
</template>
