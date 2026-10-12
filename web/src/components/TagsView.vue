<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTagsViewStore, type TagView } from '@/stores/tagsView'

defineOptions({ name: 'TagsView' })

const route = useRoute()
const router = useRouter()
const store = useTagsViewStore()

/** 当前路由标签高亮 */
function isActive(v: TagView): boolean {
  return v.path === route.path
}

function closeTag(v: TagView) {
  if (v.affix) return
  const views = store.delView(v)
  // 关闭的是当前页 → 跳到最后一个标签（无标签回首页）
  if (isActive(v)) toLastView(views)
}

function toLastView(views: TagView[]) {
  const latest = views[views.length - 1]
  if (latest) router.push(latest.fullPath || latest.path)
  else router.push('/')
}

/**
 * 刷新当前页：从 tagsView store 定位当前 path 对应的标签（TagsView 仅在 MainLayout
 * 内渲染，正常恒命中）。keep-alive 已缓存实例且 router-view 未绑定 key，SPA 内无法
 * 强制重挂组件 → 选最小侵入方案 = 整页重载（hash 路由 URL 不变，token 存 localStorage 不丢）
 */
function refreshCurrent() {
  closeCtx()
  const cur = store.visitedViews.find((v) => v.path === route.path)
  if (!cur) return
  window.location.reload()
}

/* ===== 右键菜单（可选能力）：关闭 / 关闭其他 / 关闭全部 ===== */
const ctx = reactive({ visible: false, left: 0, top: 0, selected: null as TagView | null })

function openCtx(e: MouseEvent, v: TagView) {
  ctx.visible = true
  ctx.left = e.clientX
  ctx.top = e.clientY
  ctx.selected = v
}
function closeCtx() {
  ctx.visible = false
}
function closeOthers() {
  if (ctx.selected) store.delOthersViews(ctx.selected)
  closeCtx()
}
function closeAll() {
  toLastView(store.delAllViews())
  closeCtx()
}

// 点击页面任意处收起右键菜单
function hideOnGlobalClick() {
  closeCtx()
}
onMounted(() => document.addEventListener('click', hideOnGlobalClick))
onBeforeUnmount(() => document.removeEventListener('click', hideOnGlobalClick))
</script>

<template>
  <div class="tags-view">
    <div class="tags-scroll">
      <router-link
        v-for="v in store.visitedViews"
        :key="v.path"
        :to="v.fullPath || v.path"
        custom
        v-slot="{ navigate }"
      >
        <span
          class="tag"
          :class="{ active: isActive(v) }"
          @click="navigate"
          @contextmenu.prevent="openCtx($event, v)"
        >
          <span class="dot" />
          <span class="tag-title">{{ v.title }}</span>
          <el-icon v-if="!v.affix" class="close" @click.stop="closeTag(v)"><Close /></el-icon>
        </span>
      </router-link>
    </div>

    <!-- 刷新当前页 -->
    <el-tooltip content="刷新当前" placement="top">
      <el-icon class="refresh-btn" :size="14" @click="refreshCurrent"><RefreshRight /></el-icon>
    </el-tooltip>

    <teleport to="body">
      <ul
        v-if="ctx.visible"
        class="ctx-menu"
        :style="{ left: `${ctx.left}px`, top: `${ctx.top}px` }"
      >
        <li @click="refreshCurrent">刷新当前</li>
        <li v-if="ctx.selected && !ctx.selected.affix" @click="closeTag(ctx.selected); closeCtx()">
          关闭当前
        </li>
        <li @click="closeOthers">关闭其他</li>
        <li @click="closeAll">关闭全部</li>
      </ul>
    </teleport>
  </div>
</template>

<style scoped>
.tags-view {
  height: 34px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 3px rgba(0, 21, 41, 0.04);
  display: flex;
  align-items: center;
  padding: 0 10px;
  position: relative;
  z-index: 1;
}
.tags-scroll {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  white-space: nowrap;
}
.tags-scroll::-webkit-scrollbar {
  height: 4px;
}
.refresh-btn {
  margin-left: auto;
  flex-shrink: 0;
  padding: 3px;
  border-radius: 3px;
  color: #606266;
  cursor: pointer;
  outline: none;
}
.refresh-btn:hover {
  background: #ecf5ff;
  color: #409eff;
}
.tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 24px;
  padding: 0 8px;
  border: 1px solid #d8e1eb;
  border-radius: 3px;
  background: #fff;
  color: #495060;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
}
.tag:hover {
  border-color: #b3d8ff;
  color: #409eff;
}
.tag.active {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ddd;
}
.tag.active .dot {
  background: #fff;
}
.close {
  font-size: 12px;
  border-radius: 50%;
}
.close:hover {
  background: rgba(0, 0, 0, 0.16);
  color: #fff;
}
.ctx-menu {
  position: fixed;
  z-index: 3000;
  margin: 0;
  padding: 4px 0;
  list-style: none;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.12);
  font-size: 13px;
  color: #303133;
}
.ctx-menu li {
  padding: 6px 16px;
  cursor: pointer;
}
.ctx-menu li:hover {
  background: #ecf5ff;
  color: #409eff;
}
</style>
