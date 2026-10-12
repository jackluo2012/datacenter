/**
 * 大屏图表底座：
 * - 统一 DataV 风深色色板与坐标轴/tooltip 样式常量
 * - 容器为 1920 设计稿固定像素（外层 transform scale 自适应），故 init 一次即可，
 *   无需监听容器尺寸；数据变化时 setOption(option, true) 整体重建序列
 * - 卸载时 dispose，防 ECharts 实例泄漏
 */
import * as echarts from 'echarts'
import { onBeforeUnmount, onMounted, shallowRef, type Ref } from 'vue'

/** 大屏统一色板（深色底 #0a1628 + 科技蓝青点缀，黄/绿/玫红点缀） */
export const BS = {
  /** 页面/卡片底色（同时也是饼图扇区描边色，形成 2px 分隔间隙） */
  surface: '#0a1628',
  cyan: '#00d4ff',
  green: '#00ff88',
  yellow: '#ffd855',
  violet: '#7b8cff',
  rose: '#ff5e7a',
  slate: '#8fa3bf',
  textMain: '#d8ecff',
  textSub: '#7f9bbf',
  axisLine: 'rgba(127,155,191,0.28)',
  splitLine: 'rgba(0,212,255,0.10)',
  tooltipBg: 'rgba(7,18,36,0.94)',
  tooltipBorder: 'rgba(0,212,255,0.35)',
} as const

/**
 * 多分类（环形图扇区）固定顺序，禁止按数值轮换：
 * 青 → 玫红 → 黄 → 蓝紫 → 绿 → 灰(其他)。已按相邻对 CVD 区分度校验排序，
 * 配合直接标签 + 扇区 2px 间隙做二次编码。
 */
export const BS_CATEGORICAL = [BS.cyan, BS.rose, BS.yellow, BS.violet, BS.green, BS.slate] as const

/** 深色 tooltip 公共片段（不 as const：避免只读元组与 ECharts padding 类型冲突） */
export const bsTooltip = {
  backgroundColor: BS.tooltipBg,
  borderColor: BS.tooltipBorder,
  borderWidth: 1,
  padding: [6, 10],
  textStyle: { color: BS.textMain, fontSize: 12 },
  extraCssText: 'box-shadow: 0 4px 16px rgba(0,0,0,0.45); border-radius: 4px;',
}

/** 深色类目/数值轴公共片段 */
export const bsAxisLabel = { color: BS.textSub, fontSize: 11 } as const

export function useChart(el: Ref<HTMLElement | undefined>, buildOption: () => echarts.EChartsOption) {
  const chart = shallowRef<echarts.ECharts>()

  function render(): void {
    if (!el.value) return
    if (!chart.value) chart.value = echarts.init(el.value)
    chart.value.setOption(buildOption(), true)
  }

  onMounted(render)
  onBeforeUnmount(() => {
    chart.value?.dispose()
    chart.value = undefined
  })

  return { chart, render }
}
