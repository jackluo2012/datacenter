<script setup lang="ts">
/**
 * 质量规则 6 类型分布环形图：固定顺序分类色板 + 扇区 2px 间隙 + 右侧直接标签图例；
 * 中心叠加规则总数。数据可为真实聚合或标注的演示数据（由父级传 badge）。
 */
import { computed, ref, watch } from 'vue'
import { BS, BS_CATEGORICAL, bsTooltip, useChart } from '../useChart'
import BsEmpty from './BsEmpty.vue'

interface DonutItem {
  name: string
  value: number
}

const props = withDefaults(
  defineProps<{
    items: DonutItem[]
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

const el = ref<HTMLElement>()
const { render } = useChart(el, () => ({
  tooltip: {
    ...bsTooltip,
    trigger: 'item',
    formatter: (p: unknown) => {
      const d = p as { name: string; value: number; percent: number }
      return `${d.name}<br/>规则数：<b style="color:${BS.cyan}">${d.value}</b>（${d.percent}%）`
    },
  },
  series: [
    {
      type: 'pie',
      radius: ['62%', '82%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: BS.surface, borderWidth: 2 },
      label: { show: false },
      labelLine: { show: false },
      emphasis: { scaleSize: 4 },
      data: props.items.map((it, i) => ({
        name: it.name,
        value: it.value,
        itemStyle: { color: BS_CATEGORICAL[i % BS_CATEGORICAL.length] },
      })),
    },
  ],
}))

watch(() => props.items, render, { flush: 'post' })

const total = computed(() => props.items.reduce((s, it) => s + it.value, 0))
const pct = (v: number): string => (total.value ? `${((v / total.value) * 100).toFixed(0)}%` : '0%')
</script>

<template>
  <div class="dt-wrap">
    <div v-if="items.length && !loading && !failed" ref="el" class="dt-chart" />
    <!-- 中心总数（纯展示，不遮挡交互） -->
    <div v-if="items.length && !loading && !failed" class="dt-center">
      <div class="dt-total">{{ total.toLocaleString('zh-CN') }}</div>
      <div class="dt-total-label">规则总数</div>
    </div>
    <div v-if="items.length && !loading && !failed" class="dt-legend">
      <div v-for="(it, i) in items" :key="it.name" class="dt-row">
        <span class="dt-dot" :style="{ background: BS_CATEGORICAL[i % BS_CATEGORICAL.length] }" />
        <span class="dt-name">{{ it.name }}</span>
        <span class="dt-val">{{ it.value }}</span>
        <span class="dt-pct">{{ pct(it.value) }}</span>
      </div>
    </div>
    <BsEmpty v-if="loading || failed || !items.length" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.dt-wrap {
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
}
.dt-chart {
  width: 46%;
  height: 100%;
  flex-shrink: 0;
}
.dt-center {
  position: absolute;
  left: 0;
  top: 50%;
  width: 46%;
  transform: translateY(-50%);
  text-align: center;
  pointer-events: none;
}
.dt-total {
  font-family: 'DIN Alternate', 'Bahnschrift', 'Consolas', monospace;
  font-size: 26px;
  font-weight: 700;
  color: #d8ecff;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
}
.dt-total-label {
  margin-top: 2px;
  font-size: 11px;
  letter-spacing: 2px;
  color: #7f9bbf;
}
.dt-legend {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 7px;
  padding-left: 10px;
}
.dt-row {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
}
.dt-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
  flex-shrink: 0;
  box-shadow: 0 0 5px currentColor;
}
.dt-name {
  color: #a9c3e2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dt-val {
  margin-left: auto;
  color: #d8ecff;
  font-family: 'Consolas', monospace;
  font-weight: 600;
}
.dt-pct {
  width: 36px;
  text-align: right;
  color: #7f9bbf;
  font-size: 11px;
}
</style>
