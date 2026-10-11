<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import { useRouter } from 'vue-router'
import { apiList, getCallStats, type CallStatsResp } from '@/api/datasvc'
import { listProfiles, qualityRuleList, type TableProfile } from '@/api/quality'
import { dataSourceList } from '@/api/datasource'

const router = useRouter()

// ===== 概览指标卡：全部接真实接口，失败置 '-' 占位（不报错） =====
const stats = ref([
  { label: '数据源', value: '-' as string | number, icon: 'Coin', color: '#409eff', path: '/datasource' },
  { label: '数据服务 API', value: '-' as string | number, icon: 'Connection', color: '#e6a23c', path: '/datasvc' },
  { label: '质量规则', value: '-' as string | number, icon: 'CircleCheck', color: '#67c23a', path: '/quality' },
  { label: '近 7 日 API 调用', value: '-' as string | number, icon: 'TrendCharts', color: '#f56c6c', path: '/datasvc' },
])

const quickEntries = [
  { title: '登记数据源', desc: '接入 MySQL / ClickHouse 等', path: '/datasource', icon: 'Coin' },
  { title: '采集元数据', desc: '一键采集库表结构', path: '/metadata', icon: 'Collection' },
  { title: '发布数据 API', desc: 'SQL 一键生成服务', path: '/datasvc', icon: 'Connection' },
  { title: 'AI 问数', desc: '自然语言查询数据', path: '/ai', icon: 'ChatDotRound' },
  { title: '质量监控', desc: '规则化校验数据', path: '/quality', icon: 'CircleCheck' },
  { title: 'SQL IDE', desc: '在线查询与脚本管理', path: '/ide', icon: 'EditPen' },
]

// ---------- 指标卡：数据源 / API / 规则总数 ----------
async function loadCounts() {
  const [ds, api, rule] = await Promise.allSettled([
    dataSourceList({ page: 1, size: 1 }),
    apiList({ page: 1, size: 1 }),
    qualityRuleList({ page: 1, size: 1 }),
  ])
  if (ds.status === 'fulfilled') stats.value[0].value = ds.value.total ?? 0
  if (api.status === 'fulfilled') stats.value[1].value = api.value.total ?? 0
  if (rule.status === 'fulfilled') stats.value[2].value = rule.value.total ?? 0
  // 失败的卡片保持 '-' 占位
}

// ===== 近 7 日 API 调用趋势（/datasvc/callstats days=7） =====
const statsData = ref<CallStatsResp | null>(null)
const statsLoaded = ref(false)
const statsOffline = ref(false)
const hasDaily = computed(() => (statsData.value?.daily_dates?.length ?? 0) > 0)

const trendEl = ref<HTMLDivElement>()
let trendChart: echarts.ECharts | null = null

const trendEmptyText = computed(() => {
  if (!statsLoaded.value) return '加载中…'
  return statsOffline.value ? '调用统计接口上线中，趋势图暂不可用' : '暂无调用数据'
})

async function loadCallStats() {
  try {
    const resp = await getCallStats(7)
    statsData.value = resp
    stats.value[3].value = resp.overall?.total_calls ?? 0
    statsOffline.value = false
  } catch {
    // 统计接口未就绪 / 失败：占位展示，不报错
    statsOffline.value = true
  } finally {
    statsLoaded.value = true
    nextTick(renderTrend)
  }
}

function renderTrend() {
  const el = trendEl.value
  const dates = statsData.value?.daily_dates || []
  const calls = statsData.value?.daily_calls || []
  if (!el || !dates.length) return
  trendChart?.dispose()
  trendChart = echarts.init(el)
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 50, right: 20, bottom: 26, top: 30 },
    xAxis: { type: 'category', boundaryGap: false, data: dates },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      {
        name: '调用量',
        type: 'line',
        smooth: true,
        symbolSize: 6,
        data: calls,
        areaStyle: { opacity: 0.15 },
        itemStyle: { color: '#409eff' },
      },
    ],
  })
}

// ===== 表健康分 Top5（/quality/profile/list 最新 5 条） =====
const healthLoading = ref(true)
const healthOffline = ref(false)
const healthTop = ref<TableProfile[]>([])

async function loadHealthTop() {
  try {
    const resp = await listProfiles({ page: 1, size: 50 })
    const sorted = [...(resp.list || [])].sort((a, b) => (b.ran_at || 0) - (a.ran_at || 0))
    healthTop.value = sorted.slice(0, 5)
    healthOffline.value = false
  } catch {
    healthOffline.value = true
  } finally {
    healthLoading.value = false
  }
}

/** 健康分配色：<60 红 / 60-80 黄 / >=80 绿 */
function scoreTagType(score?: number): 'danger' | 'warning' | 'success' {
  const s = score ?? 0
  if (s < 60) return 'danger'
  if (s < 80) return 'warning'
  return 'success'
}

/** 空值率兼容 0~1 与 0~100 两种量纲 */
function nullRateText(r?: number): string {
  if (r === null || r === undefined) return '-'
  const pct = r <= 1 ? r * 100 : r
  return `${pct.toFixed(1)}%`
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleDateString('zh-CN')
}

function onResize() {
  trendChart?.resize()
}

onMounted(() => {
  loadCounts()
  loadCallStats()
  loadHealthTop()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  trendChart?.dispose()
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

    <!-- 图表区：近 7 日调用趋势 + 表健康分 Top5 -->
    <el-row :gutter="16" class="mt16">
      <el-col :xs="24" :lg="14">
        <el-card v-loading="!statsLoaded" shadow="never">
          <template #header><span class="card-title">近 7 日 API 调用趋势</span></template>
          <div v-if="statsLoaded && !statsOffline && hasDaily" ref="trendEl" class="chart" />
          <el-empty v-else :description="trendEmptyText" :image-size="90" />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="10">
        <el-card v-loading="healthLoading" shadow="never">
          <template #header><span class="card-title">表健康分 Top 5（最新分析）</span></template>
          <el-empty
            v-if="healthOffline"
            description="健康分接口上线中"
            :image-size="80"
          />
          <el-empty
            v-else-if="!healthTop.length"
            description="暂无分析记录，可在「数据质量 → 健康分」生成"
            :image-size="80"
          />
          <el-table v-else :data="healthTop" size="small" border>
            <el-table-column prop="table_name" label="表" min-width="110" show-overflow-tooltip />
            <el-table-column prop="rows_cnt" label="行数" width="80" align="right" />
            <el-table-column label="空值率" width="74" align="right">
              <template #default="{ row }">{{ nullRateText(row.null_rate) }}</template>
            </el-table-column>
            <el-table-column prop="dup_pk_cnt" label="重复" width="66" align="right" />
            <el-table-column label="分析日" width="90">
              <template #default="{ row }">{{ fmtTime(row.ran_at) }}</template>
            </el-table-column>
            <el-table-column label="健康分" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="scoreTagType(row.health_score)" size="small">
                  {{ row.health_score ?? '-' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
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
