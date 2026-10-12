<script setup lang="ts">
/** 表健康分环形仪表：进度弧渐变（青→绿）+ 中心大数字（valueAnimation 自带滚动） */
import * as echarts from 'echarts'
import { ref, watch } from 'vue'
import { BS, useChart } from '../useChart'
import BsEmpty from './BsEmpty.vue'

const props = withDefaults(
  defineProps<{
    score: number | null
    tableName?: string
    loading?: boolean
    failed?: boolean
  }>(),
  { loading: false, failed: false },
)

const el = ref<HTMLElement>()
const { render } = useChart(el, () => ({
  series: [
    {
      type: 'gauge',
      startAngle: 210,
      endAngle: -30,
      min: 0,
      max: 100,
      radius: '88%',
      center: ['50%', '58%'],
      progress: {
        show: true,
        width: 16,
        roundCap: true,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(1, 0, 0, 0, [
            { offset: 0, color: BS.green },
            { offset: 1, color: BS.cyan },
          ]),
          shadowColor: 'rgba(0,212,255,0.55)',
          shadowBlur: 10,
        },
      },
      axisLine: { roundCap: true, lineStyle: { width: 16, color: [[1, 'rgba(0,212,255,0.10)']] } },
      axisTick: { distance: -25, length: 4, lineStyle: { color: 'rgba(0,212,255,0.35)', width: 1 } },
      splitLine: { distance: -27, length: 8, lineStyle: { color: 'rgba(0,212,255,0.55)', width: 2 } },
      axisLabel: { distance: -12, color: BS.textSub, fontSize: 10 },
      pointer: { show: false },
      anchor: { show: false },
      title: { show: true, offsetCenter: [0, '36%'], color: BS.textSub, fontSize: 12 },
      detail: {
        valueAnimation: true,
        offsetCenter: [0, '-4%'],
        formatter: (v: number) => (v === null || v === undefined ? '--' : Number(v).toFixed(1)),
        color: BS.cyan,
        fontSize: 42,
        fontWeight: 700,
        fontFamily: 'DIN Alternate, Bahnschrift, Consolas, monospace',
      },
      data: [{ value: props.score ?? 0, name: props.tableName || '最新画像表健康分' }],
    },
  ],
}))

watch(() => [props.score, props.tableName], render, { flush: 'post' })
</script>

<template>
  <div class="gr-wrap">
    <div v-if="score !== null && !loading && !failed" ref="el" class="gr-chart" />
    <BsEmpty
      v-if="loading || failed || score === null"
      :loading="loading"
      :failed="failed"
      text="暂无健康画像 · 执行 /quality/profile/table 生成"
    />
  </div>
</template>

<style scoped>
.gr-wrap {
  position: relative;
  height: 100%;
}
.gr-chart {
  width: 100%;
  height: 100%;
}
</style>
