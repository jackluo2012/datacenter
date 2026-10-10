<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import * as echarts from 'echarts'
import { useRouter } from 'vue-router'

const router = useRouter()

// ===== 概览指标：暂用 mock 占位，待网关聚合接口上线后替换 =====
// TODO: 接真实接口 —— 数据源数 /datasource/list、API 数 /datasvc/api/list、
//       表数 /metadata/entity/list(entity_type=table)、今日问数次 /ops/log/audit(action=ai_ask)
const stats = ref([
  { label: '数据源', value: 6, icon: 'Coin', color: '#409eff', path: '/datasource' },
  { label: '元数据表', value: 128, icon: 'Collection', color: '#67c23a', path: '/metadata' },
  { label: '数据服务 API', value: 23, icon: 'Connection', color: '#e6a23c', path: '/datasvc' },
  { label: '今日 AI 问数', value: 57, icon: 'ChatDotRound', color: '#f56c6c', path: '/ai' },
])

const quickEntries = [
  { title: '登记数据源', desc: '接入 MySQL / ClickHouse 等', path: '/datasource', icon: 'Coin' },
  { title: '采集元数据', desc: '一键采集库表结构', path: '/metadata', icon: 'Collection' },
  { title: '发布数据 API', desc: 'SQL 一键生成服务', path: '/datasvc', icon: 'Connection' },
  { title: 'AI 问数', desc: '自然语言查询数据', path: '/ai', icon: 'ChatDotRound' },
  { title: '质量监控', desc: '规则化校验数据', path: '/quality', icon: 'CircleCheck' },
  { title: 'SQL IDE', desc: '在线查询与脚本管理', path: '/ide', icon: 'EditPen' },
]

// ===== ECharts：服务数/表数 与 今日问数趋势（mock 数据） =====
const coreChartEl = ref<HTMLDivElement>()
const trendChartEl = ref<HTMLDivElement>()
let coreChart: echarts.ECharts | null = null
let trendChart: echarts.ECharts | null = null

function renderCore() {
  if (!coreChartEl.value) return
  coreChart = echarts.init(coreChartEl.value)
  // mock：各模块资源数量统计，待接 /metadata/entity/list 等真实接口
  coreChart.setOption({
    title: { text: '核心资源统计', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, bottom: 30, top: 50 },
    xAxis: {
      type: 'category',
      data: ['数据源', '元数据表', '数据 API', '质量规则', '指标', '标签'],
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '数量',
        type: 'bar',
        barWidth: 36,
        itemStyle: { color: '#409eff', borderRadius: [4, 4, 0, 0] },
        data: [6, 128, 23, 45, 62, 38],
      },
    ],
  })
}

function renderTrend() {
  if (!trendChartEl.value) return
  trendChart = echarts.init(trendChartEl.value)
  // mock：近 7 日 AI 问数次数趋势，待接审计日志统计接口
  const days = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(Date.now() - (6 - i) * 86400000)
    return `${d.getMonth() + 1}/${d.getDate()}`
  })
  trendChart.setOption({
    title: { text: '近 7 日 AI 问数趋势', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, bottom: 30, top: 50 },
    xAxis: { type: 'category', boundaryGap: false, data: days },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      {
        name: '问数次',
        type: 'line',
        smooth: true,
        symbolSize: 7,
        areaStyle: { opacity: 0.15 },
        itemStyle: { color: '#67c23a' },
        data: [21, 35, 28, 44, 39, 48, 57],
      },
    ],
  })
}

function onResize() {
  coreChart?.resize()
  trendChart?.resize()
}

onMounted(() => {
  renderCore()
  renderTrend()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  coreChart?.dispose()
  trendChart?.dispose()
  coreChart = null
  trendChart = null
})
</script>

<template>
  <div class="dashboard">
    <!-- 概览指标卡 -->
    <el-row :gutter="16">
      <el-col v-for="s in stats" :key="s.label" :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card" @click="router.push(s.path)">
          <div class="stat-body">
            <el-icon :size="34" :color="s.color"><component :is="s.icon" /></el-icon>
            <div class="stat-meta">
              <div class="stat-value">{{ s.value }}</div>
              <div class="stat-label">{{ s.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区 -->
    <el-row :gutter="16" class="mt16">
      <el-col :xs="24" :lg="12">
        <el-card shadow="never">
          <div ref="coreChartEl" class="chart" />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card shadow="never">
          <div ref="trendChartEl" class="chart" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷入口 -->
    <el-card shadow="never" class="mt16">
      <template #header><span class="card-title">快捷入口</span></template>
      <el-row :gutter="16">
        <el-col v-for="e in quickEntries" :key="e.title" :xs="12" :sm="8" :md="4">
          <div class="entry" @click="router.push(e.path)">
            <el-icon :size="24" color="#409eff"><component :is="e.icon" /></el-icon>
            <div class="entry-title">{{ e.title }}</div>
            <div class="entry-desc">{{ e.desc }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<style scoped>
.stat-card {
  cursor: pointer;
  margin-bottom: 4px;
}
.stat-body {
  display: flex;
  align-items: center;
  gap: 14px;
}
.stat-value {
  font-size: 26px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: #909399;
}
.mt16 {
  margin-top: 16px;
}
.chart {
  width: 100%;
  height: 320px;
}
.card-title {
  font-weight: 600;
  font-size: 14px;
}
.entry {
  text-align: center;
  padding: 16px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}
.entry:hover {
  background: #ecf5ff;
}
.entry-title {
  margin-top: 8px;
  font-size: 14px;
  color: #303133;
}
.entry-desc {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
</style>
