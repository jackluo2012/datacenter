<script setup lang="ts">
/**
 * 大屏占位三态：loading → 骨架 shimmer；fail → 「链路未连接 · 演示占位」；
 * 其他 → 「暂无数据」。失败不弹错误，静默占位。
 */
withDefaults(
  defineProps<{
    loading?: boolean
    failed?: boolean
    text?: string
  }>(),
  { loading: false, failed: false, text: '暂无数据' },
)
</script>

<template>
  <div class="be-wrap">
    <template v-if="loading">
      <div class="be-skeleton">
        <span class="be-bar w60" />
        <span class="be-bar w90" />
        <span class="be-bar w75" />
      </div>
    </template>
    <template v-else>
      <div class="be-empty">
        <span class="be-line" />
        <span class="be-text" :class="{ 'be-failed': failed }">
          {{ failed ? '数据链路未连接 · 演示占位' : text }}
        </span>
        <span class="be-line" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.be-wrap {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3;
}
.be-skeleton {
  width: 82%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.be-bar {
  height: 12px;
  border-radius: 3px;
  background: linear-gradient(
    90deg,
    rgba(0, 212, 255, 0.06) 25%,
    rgba(0, 212, 255, 0.18) 50%,
    rgba(0, 212, 255, 0.06) 75%
  );
  background-size: 200% 100%;
  animation: be-shimmer 1.4s ease-in-out infinite;
}
.w60 { width: 60%; }
.w75 { width: 75%; }
.w90 { width: 90%; }
@keyframes be-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
.be-empty {
  display: flex;
  align-items: center;
  gap: 10px;
  opacity: 0.85;
}
.be-line {
  width: 34px;
  height: 1px;
  background: linear-gradient(90deg, rgba(0, 212, 255, 0), rgba(0, 212, 255, 0.4));
}
.be-line:last-child {
  background: linear-gradient(90deg, rgba(0, 212, 255, 0.4), rgba(0, 212, 255, 0));
}
.be-text {
  font-size: 12px;
  letter-spacing: 1px;
  color: #7f9bbf;
  white-space: nowrap;
}
.be-failed {
  color: #8fa3bf;
}
</style>
