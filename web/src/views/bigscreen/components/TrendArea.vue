<script setup lang="ts">
/** 近 N 日 API 调用趋势：渐变填充面积图（单序列青色，末点直接标注） */
import * as echarts from 'echarts'
import { ref, watch } from 'vue'
import { BS, bsAxisLabel, bsTooltip, useChart } from '../useChart'
import BsEmpty from './BsEmpty.vue'

const props = withDefaults(
  defineProps<{
    dates: string[]
    values: number[]
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

const el = ref<HTMLElement>()
const { render } = useChart(el, () => ({
  grid: { left: 48, right: 22, top: 28, bottom: 26 },
  tooltip: {
    ...bsTooltip,
    trigger: 'axis',
    axisPointer: { type: 'line', lineStyle: { color: 'rgba(0,212,255,0.45)' } },
    formatter: (params: unknown) => {
      const p = (params as { name: string; value: number }[])[0]
      return `${p.name}<br/>调用量：<b style="color:${BS.cyan}">${Number(p.value).toLocaleString('zh-CN')}</b>`
    },
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: props.dates,
    axisLine: { lineStyle: { color: BS.axisLine } },
    axisTick: { show: false },
    axisLabel: { ...bsAxisLabel, formatter: (v: string) => v.slice(5) },
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: BS.splitLine, width: 1 } },
    axisLabel: bsAxisLabel,
  },
  series: [
    {
      name: '调用量',
      type: 'line',
      smooth: 0.4,
      showSymbol: false,
      lineStyle: { width: 2, color: BS.cyan, shadowColor: 'rgba(0,212,255,0.5)', shadowBlur: 8 },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(0,212,255,0.38)' },
          { offset: 1, color: 'rgba(0,212,255,0.02)' },
        ]),
      },
      data: props.values.map((v, i) =>
        i === props.values.length - 1 && Number.isFinite(v)
          ? {
              value: v,
              label: {
                show: true,
                position: 'top',
                distance: 6,
                color: BS.cyan,
                fontSize: 12,
                fontWeight: 700,
                formatter: (d: unknown) =>
                  Number((d as { value: number }).value).toLocaleString('zh-CN'),
              },
            }
          : v,
      ),
    },
  ],
}))

watch(() => [props.dates, props.values], render, { flush: 'post' })
</script>

<template>
  <div class="ta-wrap">
    <div ref="el" class="ta-chart" />
    <BsEmpty v-if="loading || failed || !values.length" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.ta-wrap {
  position: relative;
  height: 100%;
}
.ta-chart {
  width: 100%;
  height: 100%;
}
</style>
