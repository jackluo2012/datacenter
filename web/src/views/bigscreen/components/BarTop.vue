<script setup lang="ts">
/** TopN API 横向条形：单序列单色（青色渐变），条端 4px 圆角，右端直接标数值 */
import * as echarts from 'echarts'
import { ref, watch } from 'vue'
import { BS, bsTooltip, useChart } from '../useChart'
import BsEmpty from './BsEmpty.vue'

interface BarItem {
  name: string
  value: number
  cost?: number
}

const props = withDefaults(
  defineProps<{
    items: BarItem[]
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

const el = ref<HTMLElement>()
const { render } = useChart(el, () => ({
  grid: { left: 8, right: 56, top: 6, bottom: 6, containLabel: true },
  tooltip: {
    ...bsTooltip,
    trigger: 'item',
    formatter: (p: unknown) => {
      const d = p as { name: string; value: number; dataIndex: number }
      const cost = props.items[d.dataIndex]?.cost
      return `${d.name}<br/>调用量：<b style="color:${BS.cyan}">${Number(d.value).toLocaleString('zh-CN')}</b>${
        cost !== undefined ? `<br/>平均耗时：<b style="color:${BS.yellow}">${cost} ms</b>` : ''
      }`
    },
  },
  xAxis: {
    type: 'value',
    splitLine: { show: false },
    axisLabel: { show: false },
    axisLine: { show: false },
  },
  yAxis: {
    type: 'category',
    inverse: true,
    data: props.items.map((it) => (it.name.length > 22 ? `${it.name.slice(0, 21)}…` : it.name)),
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { color: BS.textMain, fontSize: 12 },
  },
  series: [
    {
      type: 'bar',
      barWidth: 10,
      data: props.items.map((it) => it.value),
      itemStyle: {
        borderRadius: [0, 4, 4, 0],
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: 'rgba(0,212,255,0.25)' },
          { offset: 1, color: BS.cyan },
        ]),
      },
      label: {
        show: true,
        position: 'right',
        color: BS.cyan,
        fontSize: 12,
        fontWeight: 600,
        fontFamily: 'Consolas, monospace',
        formatter: (d: unknown) => Number((d as { value: number }).value).toLocaleString('zh-CN'),
      },
    },
  ],
}))

watch(() => props.items, render, { flush: 'post' })
</script>

<template>
  <div class="bt-wrap">
    <div ref="el" class="bt-chart" />
    <BsEmpty v-if="loading || failed || !items.length" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.bt-wrap {
  position: relative;
  height: 100%;
}
.bt-chart {
  width: 100%;
  height: 100%;
}
</style>
