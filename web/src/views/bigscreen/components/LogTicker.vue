<script setup lang="ts">
/**
 * 数据服务调用日志滚动列表：最新 N 条常驻，高亮行定时轮播（2.5s 下移循环）。
 * 状态三色（ok 绿 / error 红 / denied 黄），行进入时轻微滑入。
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ApiCallLog } from '@/api/datasvc'
import BsEmpty from './BsEmpty.vue'

const props = withDefaults(
  defineProps<{
    logs: ApiCallLog[]
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

const HL_MS = 2500
const active = ref(0)
let timer = 0

onMounted(() => {
  timer = window.setInterval(() => {
    if (props.logs.length) active.value = (active.value + 1) % props.logs.length
  }, HL_MS)
})
onBeforeUnmount(() => window.clearInterval(timer))
watch(
  () => props.logs,
  () => {
    active.value = 0
  },
)

const statusText: Record<string, string> = { ok: '成功', error: '失败', denied: '拒绝' }

const rows = computed(() =>
  props.logs.map((l) => ({
    time: l.created_at
      ? new Date(l.created_at * (l.created_at < 1e12 ? 1000 : 1)).toLocaleTimeString('zh-CN', {
          hour12: false,
        })
      : '--:--:--',
    path: l.path || `api#${l.api_id ?? '-'}`,
    status: l.status || 'ok',
    cost: l.cost_ms ?? null,
    rowsCnt: l.rows_cnt ?? null,
    cache: l.cache_hit === true,
  })),
)

const fmtCost = (c: number | null): string => (c === null ? '--' : `${c}ms`)
const fmtRows = (n: number | null): string =>
  n === null ? '' : `${n > 9999 ? `${(n / 1000).toFixed(1)}k` : n} 行`
</script>

<template>
  <div class="lt-wrap">
    <template v-if="!loading && !failed && rows.length">
      <div class="lt-head">
        <span class="lt-time">时间</span>
        <span class="lt-path">服务路径</span>
        <span class="lt-status">状态</span>
        <span class="lt-cost">耗时</span>
      </div>
      <div class="lt-body">
        <div
          v-for="(r, i) in rows"
          :key="`${r.time}-${i}`"
          class="lt-row"
          :class="{ active: i === active, slide: i === 0 }"
        >
          <span class="lt-time">{{ r.time }}</span>
          <span class="lt-path" :title="r.path">{{ r.path }}</span>
          <span class="lt-status">
            <i class="lt-dot" :class="r.status" />{{ statusText[r.status] || r.status }}
          </span>
          <span class="lt-cost">
            {{ fmtCost(r.cost) }}<em v-if="r.rowsCnt" class="lt-rows">{{ fmtRows(r.rowsCnt) }}</em>
            <em v-if="r.cache" class="lt-cache">缓存</em>
          </span>
        </div>
      </div>
    </template>
    <BsEmpty v-if="loading || failed || !rows.length" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.lt-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
  font-size: 12px;
}
.lt-head,
.lt-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 6px;
}
.lt-head {
  flex-shrink: 0;
  height: 24px;
  color: #5a7191;
  font-size: 11px;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(0, 212, 255, 0.12);
}
.lt-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.lt-row {
  flex: 1;
  min-height: 0;
  border-radius: 3px;
  border: 1px solid transparent;
  transition: background 0.4s, box-shadow 0.4s, border-color 0.4s;
}
.lt-row.active {
  background: linear-gradient(90deg, rgba(0, 212, 255, 0.14), rgba(0, 212, 255, 0.02));
  border-color: rgba(0, 212, 255, 0.35);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.12) inset;
}
.lt-row.slide {
  animation: lt-in 0.5s ease;
}
@keyframes lt-in {
  from { transform: translateY(-6px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}
.lt-time {
  flex-shrink: 0;
  width: 62px;
  color: #7f9bbf;
  font-family: 'Consolas', monospace;
}
.lt-path {
  flex: 1;
  min-width: 0;
  color: #d8ecff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'Consolas', monospace;
}
.lt-status {
  flex-shrink: 0;
  width: 54px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.lt-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.lt-dot.ok { background: #00ff88; box-shadow: 0 0 5px rgba(0, 255, 136, 0.8); }
.lt-dot.error { background: #ff5e7a; box-shadow: 0 0 5px rgba(255, 94, 122, 0.8); }
.lt-dot.denied { background: #ffd855; box-shadow: 0 0 5px rgba(255, 216, 85, 0.8); }
.lt-cost {
  flex-shrink: 0;
  width: 92px;
  text-align: right;
  color: #a9c3e2;
  font-family: 'Consolas', monospace;
  white-space: nowrap;
  overflow: hidden;
}
.lt-rows {
  margin-left: 5px;
  font-style: normal;
  font-size: 10px;
  color: #7b8cff;
}
.lt-cache {
  margin-left: 5px;
  font-style: normal;
  font-size: 10px;
  color: #00ff88;
  opacity: 0.85;
}
</style>
