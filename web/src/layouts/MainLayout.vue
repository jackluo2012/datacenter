<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isCollapse = ref(false)

interface MenuItem {
  path: string
  title: string
  icon: string
}

/** 侧边栏菜单（icon 为全局注册的 ElementPlus 图标组件名） */
const menus: MenuItem[] = [
  { path: '/dashboard', title: '仪表盘', icon: 'Odometer' },
  { path: '/datasource', title: '数据源', icon: 'Coin' },
  { path: '/metadata', title: '元数据', icon: 'Collection' },
  { path: '/datasvc', title: '数据服务', icon: 'Connection' },
  { path: '/catalog', title: '资产目录', icon: 'FolderOpened' },
  { path: '/quality', title: '数据质量', icon: 'CircleCheck' },
  { path: '/metric', title: '数据指标', icon: 'DataAnalysis' },
  { path: '/tag', title: '数据标签', icon: 'PriceTag' },
  { path: '/standard', title: '数据标准', icon: 'Document' },
  { path: '/mdm', title: '主数据', icon: 'Box' },
  { path: '/job', title: '调度集成', icon: 'Timer' },
  { path: '/ide', title: 'SQL IDE', icon: 'EditPen' },
  { path: '/ai', title: 'AI 问数', icon: 'ChatDotRound' },
  { path: '/ops', title: '数据运维', icon: 'Setting' },
]

const activeMenu = computed(() => route.path)
const displayName = computed(() => auth.username || `用户${auth.uid || ''}`)

function onLogout() {
  ElMessageBox.confirm('确定退出登录吗？', '提示', {
    confirmButtonText: '退出',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => {
      auth.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    })
    .catch(() => undefined)
}
</script>

<template>
  <el-container class="layout">
    <el-aside :width="isCollapse ? '64px' : '210px'" class="aside">
      <div class="logo">
        <el-icon :size="22"><DataLine /></el-icon>
        <span v-show="!isCollapse" class="logo-text">数据中台</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        background-color="#001529"
        text-color="#a6adb4"
        active-text-color="#ffffff"
        router
        class="menu"
      >
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <template #title>{{ m.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <el-icon class="collapse-btn" :size="20" @click="isCollapse = !isCollapse">
          <Expand v-if="isCollapse" />
          <Fold v-else />
        </el-icon>
        <div class="header-title">{{ route.meta.title || '' }}</div>
        <div class="header-right">
          <el-dropdown @command="(cmd: string) => cmd === 'logout' && onLogout()">
            <span class="user-entry">
              <el-avatar :size="28" class="avatar">
                {{ displayName.slice(0, 1).toUpperCase() }}
              </el-avatar>
              <span class="user-name">{{ displayName }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout" data-testid="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100%;
}
.aside {
  background-color: #001529;
  transition: width 0.2s;
  overflow-x: hidden;
}
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  font-size: 17px;
  font-weight: 600;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}
.logo-text {
  white-space: nowrap;
}
.menu {
  border-right: none;
}
.header {
  background: #fff;
  display: flex;
  align-items: center;
  height: 56px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  position: relative;
  z-index: 1;
}
.collapse-btn {
  cursor: pointer;
  color: #303133;
}
.header-title {
  margin-left: 16px;
  font-size: 15px;
  font-weight: 500;
  color: #303133;
}
.header-right {
  margin-left: auto;
}
.user-entry {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #303133;
  outline: none;
}
.avatar {
  background: #409eff;
  color: #fff;
  font-size: 13px;
}
.user-name {
  font-size: 14px;
}
.main {
  padding: 16px;
  overflow-y: auto;
}
</style>
