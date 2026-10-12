<script setup lang="ts">
/**
 * 数字翻牌/滚动：value 变化时 rAF 缓动从旧值滚到新值（easeOutCubic），
 * 千分位格式化；value 为 null（加载失败/未就绪）显示 '--'
 */
import { onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    value: number | null
    /** 缓动时长 ms */
    duration?: number
  }>(),
  { duration: 900 },
)

const display = ref<number | null>(null)
let raf = 0

function tween(to: number): void {
  cancelAnimationFrame(raf)
  const from = display.value ?? 0
  const start = performance.now()
  const step = (t: number): void => {
    const p = Math.min(1, (t - start) / props.duration)
    const ease = 1 - Math.pow(1 - p, 3)
    display.value = Math.round(from + (to - from) * ease)
    if (p < 1) raf = requestAnimationFrame(step)
  }
  raf = requestAnimationFrame(step)
}

watch(
  () => props.value,
  (v) => {
    if (v === null || v === undefined) {
      cancelAnimationFrame(raf)
      display.value = null
      return
    }
    tween(v)
  },
  { immediate: true },
)

onBeforeUnmount(() => cancelAnimationFrame(raf))

const fmt = (n: number): string => n.toLocaleString('zh-CN')
</script>

<template>
  <span class="fc-num" :class="{ 'fc-empty': display === null }">
    {{ display === null ? '--' : fmt(display) }}
  </span>
</template>

<style scoped>
.fc-num {
  font-family: 'DIN Alternate', 'Bahnschrift', 'Rajdhani', 'Consolas', monospace;
  font-weight: 700;
  color: #00d4ff;
  text-shadow: 0 0 12px rgba(0, 212, 255, 0.55);
  letter-spacing: 1px;
  font-variant-numeric: tabular-nums;
}
.fc-empty {
  color: #5a7191;
  text-shadow: none;
}
</style>
