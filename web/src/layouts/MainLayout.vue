<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { usePermissionStore } from '@/stores/permission'
import { useTagsViewStore } from '@/stores/tagsView'
import SidebarItem from '@/components/SidebarItem.vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import TagsView from '@/components/TagsView.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const permission = usePermissionStore()
const tagsView = useTagsViewStore()

// ===== 侧边栏：动态菜单（getRouters 注入；失败降级静态菜单） =====
const isCollapse = ref(false)
const activeMenu = computed(() => route.path)
/** 只渲染 visible 菜单（hidden 的仅注册路由不显示） */
const visibleMenus = computed(() => permission.menus.filter((m) => !m.hidden))

/** keep-alive 缓存名单：visitedViews 中各路由 name */
const cachedViews = computed(() => tagsView.cachedViews)

const displayName = computed(() => auth.nickname || auth.username || `用户${auth.uid || ''}`)

// ===== 全屏切换（手写 requestFullscreen，不引第三方库） =====
const isFullscreen = ref(false)
function syncFullscreen() {
  isFullscreen.value = !!document.fullscreenElement
}
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.().catch(() => undefined)
  } else {
    document.exitFullscreen?.().catch(() => undefined)
  }
}
onMounted(() => document.addEventListener('fullscreenchange', syncFullscreen))
onBeforeUnmount(() => document.removeEventListener('fullscreenchange', syncFullscreen))

function onLogout() {
  ElMessageBox.confirm('确定退出登录吗？', '提示', {
    confirmButtonText: '退出',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => {
      // logout 内部联动清理动态路由（permission.reset）与标签页（tagsView.delAllViews）
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
      <div class="menu-wrap">
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
          <SidebarItem v-for="m in visibleMenus" :key="m.path" :item="m" />
          <!-- 数据大屏：独立全屏页（/bigscreen 免登录白名单），固定菜单项置于动态菜单之后 -->
          <el-menu-item index="/bigscreen" class="bigscreen-entry">
            <el-icon :size="18"><DataBoard /></el-icon>
            <template #title>数据大屏</template>
          </el-menu-item>
        </el-menu>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <el-icon class="collapse-btn" :size="20" @click="isCollapse = !isCollapse">
          <Expand v-if="isCollapse" />
          <Fold v-else />
        </el-icon>
        <!-- 面包屑：菜单树层级优先，route.matched 兜底 -->
        <Breadcrumb class="header-breadcrumb" />
        <div class="header-right">
          <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="bottom">
            <el-icon class="header-icon" :size="18" @click="toggleFullscreen">
              <Aim v-if="isFullscreen" />
              <FullScreen v-else />
            </el-icon>
          </el-tooltip>
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

      <!-- 标签页：当前路由高亮 / 关闭 / 右键 关闭其他·全部 -->
      <TagsView />

      <el-main class="main">
        <router-view v-slot="{ Component }">
          <!--
            keep-alive 缓存依赖「页面组件 name === 路由 name」：
            TODO 各 views 页面（尤其占位 List.vue）尚未声明组件 name，
            需在各页面加 defineOptions({ name: '与路由name一致' }) 后缓存才会真正生效，
            否则 include 匹配不到（script-setup 默认以文件名推断，List.vue 全部同名 'List'）。
          -->
          <keep-alive :include="cachedViews">
            <component :is="Component" />
          </keep-alive>
        </router-view>
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
  display: flex;
  flex-direction: column;
}
.logo {
  height: 56px;
  flex-shrink: 0;
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
.menu-wrap {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
.menu {
  border-right: none;
}
/* 数据大屏固定入口：与动态菜单之间加分隔线，青色点缀 */
.bigscreen-entry {
  margin-top: 6px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 2px;
}
.header {
  background: #fff;
  display: flex;
  align-items: center;
  height: 56px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  position: relative;
  z-index: 2;
}
.collapse-btn {
  cursor: pointer;
  color: #303133;
  flex-shrink: 0;
}
.header-breadcrumb {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}
.header-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 18px;
  flex-shrink: 0;
}
.header-icon {
  cursor: pointer;
  color: #303133;
  outline: none;
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
