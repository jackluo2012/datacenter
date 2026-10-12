<script setup lang="ts">
/**
 * 表健康分 Top5：三色分级（≥90 绿 / ≥70 黄 / <70 红），
 * 分数条按健康分宽度填充，行背景微光。
 */
import BsEmpty from './BsEmpty.vue'

interface HealthRow {
  table: string
  score: number
  rows: number | null
  nullRate: number | null
}

withDefaults(
  defineProps<{
    rows: HealthRow[]
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

function level(score: number): { color: string; text: string } {
  if (score >= 90) return { color: '#00ff88', text: '优' }
  if (score >= 70) return { color: '#ffd855', text: '良' }
  return { color: '#ff5e7a', text: '差' }
}

const fmtNum = (n: number | null): string =>
  n === null || n === undefined ? '--' : Number(n).toLocaleString('zh-CN')
const fmtRate = (r: number | null): string =>
  r === null || r === undefined ? '--' : `${(r <= 1 ? r * 100 : r).toFixed(1)}%`
</script>

<template>
  <div class="ht-wrap">
    <template v-if="!loading && !failed && rows.length">
      <div class="ht-head">
        <span class="ht-table">表名</span>
        <span class="ht-rows">行数</span>
        <span class="ht-null">空值率</span>
        <span class="ht-score">健康分</span>
      </div>
      <div v-for="(r, i) in rows" :key="`${r.table}-${i}`" class="ht-row">
        <span class="ht-rank" :class="`rk-${i + 1}`">{{ i + 1 }}</span>
        <span class="ht-table" :title="r.table">{{ r.table }}</span>
        <span class="ht-rows">{{ fmtNum(r.rows) }}</span>
        <span class="ht-null">{{ fmtRate(r.nullRate) }}</span>
        <span class="ht-score">
          <span class="ht-score-num" :style="{ color: level(r.score).color }">
            {{ r.score.toFixed(1) }}
          </span>
          <span class="ht-bar">
            <i
              class="ht-bar-fill"
              :style="{ width: `${Math.max(0, Math.min(100, r.score))}%`, background: level(r.score).color }"
            />
          </span>
          <span class="ht-level" :style="{ color: level(r.score).color, borderColor: level(r.score).color }">
            {{ level(r.score).text }}
          </span>
        </span>
      </div>
    </template>
    <BsEmpty v-if="loading || failed || !rows.length" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.ht-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
  font-size: 12px;
}
.ht-head,
.ht-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 6px;
}
.ht-head {
  flex-shrink: 0;
  height: 26px;
  color: #5a7191;
  font-size: 11px;
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(0, 212, 255, 0.12);
}
.ht-row {
  flex: 1;
  min-height: 0;
  border-bottom: 1px dashed rgba(0, 212, 255, 0.06);
}
.ht-row:last-child { border-bottom: none; }
.ht-rank {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  border-radius: 2px;
  font-size: 11px;
  font-weight: 700;
  color: #0a1628;
  background: #2c4a6e;
}
.rk-1 { background: #ffd855; box-shadow: 0 0 6px rgba(255, 216, 85, 0.6); }
.rk-2 { background: #00d4ff; }
.rk-3 { background: #7b8cff; }
.ht-table {
  flex: 1;
  min-width: 0;
  color: #d8ecff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'Consolas', monospace;
}
.ht-rows, .ht-null {
  flex-shrink: 0;
  width: 64px;
  text-align: right;
  color: #a9c3e2;
  font-family: 'Consolas', monospace;
}
.ht-score {
  flex-shrink: 0;
  width: 128px;
  display: flex;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
}
.ht-score-num {
  width: 42px;
  text-align: right;
  font-family: 'Consolas', monospace;
  font-weight: 700;
}
.ht-bar {
  flex: 1;
  height: 5px;
  border-radius: 3px;
  background: rgba(0, 212, 255, 0.10);
  overflow: hidden;
}
.ht-bar-fill {
  display: block;
  height: 100%;
  border-radius: 3px;
  box-shadow: 0 0 6px currentColor;
  transition: width 0.8s ease;
}
.ht-level {
  flex-shrink: 0;
  width: 20px;
  height: 16px;
  line-height: 14px;
  text-align: center;
  font-size: 10px;
  border: 1px solid;
  border-radius: 2px;
  background: rgba(0, 0, 0, 0.25);
}
</style>
