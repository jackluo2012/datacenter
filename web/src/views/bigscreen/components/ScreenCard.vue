<script setup lang="ts">
/**
 * 大屏卡片壳：发光渐变边框 + 四角角标 + 标题装饰条
 * 高度由父级布局给定（body 区 flex/grid），内部纵向弹性
 */
defineProps<{
  title: string
  /** 标题右侧小徽标，如「演示数据」 */
  badge?: string
}>()
</script>

<template>
  <div class="sc-card">
    <i class="sc-corner sc-tl" />
    <i class="sc-corner sc-tr" />
    <i class="sc-corner sc-bl" />
    <i class="sc-corner sc-br" />
    <div class="sc-head">
      <span class="sc-tick" />
      <span class="sc-title">{{ title }}</span>
      <span v-if="badge" class="sc-badge">{{ badge }}</span>
      <span class="sc-extra"><slot name="extra" /></span>
    </div>
    <div class="sc-body">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.sc-card {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border: 1px solid rgba(0, 212, 255, 0.22);
  background: linear-gradient(180deg, rgba(0, 212, 255, 0.055) 0%, rgba(10, 22, 40, 0.55) 100%);
  box-shadow:
    inset 0 0 24px rgba(0, 212, 255, 0.06),
    0 0 14px rgba(0, 212, 255, 0.08);
  border-radius: 4px;
}
/* 四角角标 */
.sc-corner {
  position: absolute;
  width: 10px;
  height: 10px;
  border-color: #00d4ff;
  border-style: solid;
  border-width: 0;
  filter: drop-shadow(0 0 4px rgba(0, 212, 255, 0.8));
}
.sc-tl { top: -1px; left: -1px; border-top-width: 2px; border-left-width: 2px; }
.sc-tr { top: -1px; right: -1px; border-top-width: 2px; border-right-width: 2px; }
.sc-bl { bottom: -1px; left: -1px; border-bottom-width: 2px; border-left-width: 2px; }
.sc-br { bottom: -1px; right: -1px; border-bottom-width: 2px; border-right-width: 2px; }

.sc-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  height: 36px;
  padding: 0 12px;
  border-bottom: 1px solid rgba(0, 212, 255, 0.14);
  background: linear-gradient(90deg, rgba(0, 212, 255, 0.10), rgba(0, 212, 255, 0) 60%);
}
.sc-tick {
  width: 3px;
  height: 14px;
  border-radius: 2px;
  background: linear-gradient(180deg, #00d4ff, #0072ce);
  box-shadow: 0 0 6px rgba(0, 212, 255, 0.9);
}
.sc-title {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 2px;
  color: #cfeaff;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.45);
  white-space: nowrap;
}
.sc-badge {
  padding: 1px 7px;
  font-size: 10px;
  line-height: 16px;
  color: #ffd855;
  border: 1px solid rgba(255, 216, 85, 0.45);
  border-radius: 2px;
  background: rgba(255, 216, 85, 0.08);
  white-space: nowrap;
}
.sc-extra {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.sc-body {
  position: relative;
  flex: 1;
  min-height: 0;
  padding: 8px 10px;
}
</style>
