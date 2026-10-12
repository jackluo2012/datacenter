<script setup lang="ts">
/**
 * 数据中台运营驾驶舱（数据可视化大屏 /bigscreen）
 *
 * - 1920×1080 设计稿，CSS transform scale 等比/拉伸自适应任意分辨率
 * - 深色科技蓝 DataV 风：发光边框卡片 / 数字翻牌 / 面积图 / 环形仪表 / 轮播日志
 * - 数据：api/bigscreen.ts 集中封装；30s 轮询 callstats + calllog
 * - 失败一律优雅占位（骨架 / '--' / 标注演示数据），不弹错误；
 *   全部失败时顶部出「演示模式」横幅，页面不白屏
 * - 免登录白名单：未带 token 时本地直接失败（不发请求，避免 401 踢回登录页）
 */
import { computed, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  bsApis,
  bsAssetTotal,
  bsCallLogs,
  bsCallStats,
  bsDataSources,
  bsProfiles,
  bsQualityRules,
  bsStdTotal,
  bsTagTotal,
} from '@/api/bigscreen'
import type { ApiCallLog, CallStatsResp } from '@/api/datasvc'
import type { QualityRule, TableProfile } from '@/api/quality'
import ScreenCard from './components/ScreenCard.vue'
import FlipCounter from './components/FlipCounter.vue'
import TrendArea from './components/TrendArea.vue'
import BarTop from './components/BarTop.vue'
import GaugeRing from './components/GaugeRing.vue'
import DonutTypes from './components/DonutTypes.vue'
import ServiceGrid from './components/ServiceGrid.vue'
import HealthTop from './components/HealthTop.vue'
import LogTicker from './components/LogTicker.vue'
import AssetCounts from './components/AssetCounts.vue'

type PanelState = 'loading' | 'ok' | 'fail'

const router = useRouter()

// ==================== 自适应缩放（1920×1080 设计稿 → 任意分辨率） ====================
const DESIGN_W = 1920
const DESIGN_H = 1080
const scale = ref({ x: 1, y: 1 })
function fitScreen(): void {
  scale.value = { x: window.innerWidth / DESIGN_W, y: window.innerHeight / DESIGN_H }
}
const screenStyle = computed(() => ({
  width: `${DESIGN_W}px`,
  height: `${DESIGN_H}px`,
  transform: `translate(-50%, -50%) scale(${scale.value.x}, ${scale.value.y})`,
}))

// ==================== 实时时钟 ====================
const now = ref(new Date())
let clockTimer = 0
function tick(): void {
  now.value = new Date()
}
const dateText = computed(() => {
  const d = now.value
  const p = (n: number): string => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
})
const weekText = computed(() => `星期${'日一二三四五六'[now.value.getDay()]}`)
const clockText = computed(() =>
  now.value.toLocaleTimeString('zh-CN', { hour12: false }),
)

// ==================== 全屏 ====================
const isFull = ref(false)
function syncFullscreen(): void {
  isFull.value = !!document.fullscreenElement
}
function toggleFullscreen(): void {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.().catch(() => undefined)
  } else {
    document.exitFullscreen?.().catch(() => undefined)
  }
}
function backConsole(): void {
  router.push('/')
}

// ==================== 数据编排（失败静默占位，不弹错误） ====================
const callStats = ref<CallStatsResp | null>(null)
const callStatsState = ref<PanelState>('loading')
const profiles = ref<TableProfile[]>([])
const profilesState = ref<PanelState>('loading')
const rules = ref<QualityRule[]>([])
const ruleState = ref<PanelState>('loading')
const logs = ref<ApiCallLog[]>([])
const logsState = ref<PanelState>('loading')
const dsTotal = ref<number | null>(null)
const dsState = ref<PanelState>('loading')
const apiTotal = ref<number | null>(null)
const apiState = ref<PanelState>('loading')
const ruleTotal = ref<number | null>(null)
const assets = ref({ asset: null as number | null, tag: null as number | null, std: null as number | null })
const assetsState = ref<PanelState>('loading')
const booted = ref(false)

async function load<T>(state: Ref<PanelState>, fn: () => Promise<T>, done: (d: T) => void): Promise<void> {
  try {
    done(await fn())
    state.value = 'ok'
  } catch {
    state.value = 'fail'
  }
}

/** silent：轮询静默刷新，不回退 loading（避免已渲染图表闪骨架） */
function loadStats(silent = false): Promise<void> {
  if (!silent) callStatsState.value = 'loading'
  return load(callStatsState, () => bsCallStats(7), (d) => {
    callStats.value = d
  })
}

function loadProfiles(): Promise<void> {
  profilesState.value = 'loading'
  return load(profilesState, () => bsProfiles(), (d) => {
    profiles.value = d.list ?? []
  })
}

function loadLogs(silent = false): Promise<void> {
  if (!silent) logsState.value = 'loading'
  return load(logsState, () => bsCallLogs(10), (d) => {
    logs.value = d.list ?? []
  })
}

function loadCounts(): Promise<void> {
  dsState.value = 'loading'
  apiState.value = 'loading'
  ruleState.value = 'loading'
  return Promise.all([
    load(dsState, () => bsDataSources(), (d) => {
      dsTotal.value = d.total ?? 0
    }),
    load(apiState, () => bsApis(), (d) => {
      apiTotal.value = d.total ?? 0
    }),
    load(ruleState, () => bsQualityRules(), (d) => {
      rules.value = d.list ?? []
      ruleTotal.value = d.total ?? rules.value.length
    }),
  ]).then(() => undefined)
}

function loadAssets(): Promise<void> {
  assetsState.value = 'loading'
  const one = (fn: () => Promise<{ total?: number }>): Promise<number | null> =>
    fn()
      .then((d) => d.total ?? 0)
      .catch(() => null)
  return Promise.all([one(bsAssetTotal), one(bsTagTotal), one(bsStdTotal)]).then(([a, t, s]) => {
    assets.value = { asset: a, tag: t, std: s }
    assetsState.value = a === null && t === null && s === null ? 'fail' : 'ok'
  })
}

async function loadAll(): Promise<void> {
  await Promise.all([loadCounts(), loadStats(), loadProfiles(), loadLogs(), loadAssets()])
}

let pollTimer = 0
onMounted(() => {
  fitScreen()
  window.addEventListener('resize', fitScreen)
  document.addEventListener('fullscreenchange', syncFullscreen)
  clockTimer = window.setInterval(tick, 1000)
  tick()
  // 30s 定时轮询：调用统计 + 调用日志（静默刷新，不闪骨架；其余低频数据进页面时加载一次）
  pollTimer = window.setInterval(
    () => {
      loadStats(true)
      loadLogs(true)
    },
    30_000,
  )
  loadAll().finally(() => {
    booted.value = true
  })
})
onBeforeUnmount(() => {
  window.clearInterval(clockTimer)
  window.clearInterval(pollTimer)
  window.removeEventListener('resize', fitScreen)
  document.removeEventListener('fullscreenchange', syncFullscreen)
})

const anyOk = computed(() =>
  [dsState, apiState, ruleState, callStatsState, profilesState, logsState, assetsState].some(
    (s) => s.value === 'ok',
  ),
)

// ==================== 左列：指标卡 / 趋势 / TopN ====================
const totalCalls = computed<number | null>(() => {
  const o = callStats.value?.overall
  if (o && o.total_calls !== undefined && o.total_calls !== null) return o.total_calls
  const daily = callStats.value?.daily_calls
  if (daily && daily.length) return daily.reduce((s, v) => s + v, 0)
  return null
})

/** 成功率兼容 0~1 与 0~100 两种量纲 */
function ratePct(rate?: number): string {
  if (rate === null || rate === undefined) return '--'
  return `${(rate <= 1 ? rate * 100 : rate).toFixed(1)}%`
}

interface FlipCard {
  label: string
  icon: string
  value: number | null
  state: PanelState
  cls: string
}
const flipCards = computed<FlipCard[]>(() => [
  { label: '数据源总数', icon: 'Coin', value: dsTotal.value, state: dsState.value, cls: 'cyan' },
  { label: 'API 总数', icon: 'Connection', value: apiTotal.value, state: apiState.value, cls: 'violet' },
  { label: '质量规则数', icon: 'CircleCheck', value: ruleTotal.value, state: ruleState.value, cls: 'green' },
  { label: '近 7 日调用', icon: 'TrendCharts', value: totalCalls.value, state: callStatsState.value, cls: 'yellow' },
])

const trendDates = computed(() => callStats.value?.daily_dates ?? [])
const trendValues = computed(() => callStats.value?.daily_calls ?? [])

const topApis = computed(() =>
  (callStats.value?.top_apis ?? []).slice(0, 8).map((t) => ({
    name: t.path || `api#${t.api_id ?? '-'}`,
    value: t.calls ?? 0,
    cost: t.avg_cost_ms,
  })),
)

// ==================== 中列：健康分仪表 / Top5 / 服务矩阵 ====================
const latestProfile = computed<TableProfile | null>(() => {
  if (!profiles.value.length) return null
  return [...profiles.value].sort((a, b) => (b.ran_at ?? 0) - (a.ran_at ?? 0))[0]
})
const latestScore = computed<number | null>(() => latestProfile.value?.health_score ?? null)
const latestRanAt = computed(() => {
  const ts = latestProfile.value?.ran_at
  if (!ts) return ''
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
})
const healthRows = computed(() =>
  [...profiles.value]
    .sort((a, b) => (b.health_score ?? 0) - (a.health_score ?? 0))
    .slice(0, 5)
    .map((p) => ({
      table: p.table_name || '-',
      score: p.health_score ?? 0,
      rows: p.rows_cnt ?? null,
      nullRate: p.null_rate ?? null,
    })),
)

/** 任一数据接口成功 → 网关可达，服务矩阵全绿（探活示意） */
const servicesAlive = computed(
  () =>
    [dsState, apiState, ruleState, callStatsState, profilesState, logsState, assetsState].some(
      (s) => s.value === 'ok',
    ),
)

// ==================== 右列：规则分布 / 资产计数 / 日志 ====================
const RULE_TYPES: { key: string; label: string }[] = [
  { key: 'not_null', label: '非空校验' },
  { key: 'unique', label: '唯一性' },
  { key: 'range', label: '值域' },
  { key: 'row_count', label: '行数校验' },
  { key: 'sql', label: 'SQL 自定义' },
  { key: 'other', label: '其他' },
]
/** 演示占位分布（页面标注「演示数据」徽标） */
const MOCK_RULE_DIST = [12, 8, 6, 5, 3, 2]

const ruleDist = computed<{ items: { name: string; value: number }[]; mock: boolean }>(() => {
  const demo = (): { items: { name: string; value: number }[]; mock: boolean } => ({
    mock: true,
    items: RULE_TYPES.map((t, i) => ({ name: t.label, value: MOCK_RULE_DIST[i] })),
  })
  if (ruleState.value !== 'ok' || !rules.value.length) return demo()
  const known = RULE_TYPES.filter((t) => t.key !== 'other')
  const counts = new Map<string, number>()
  for (const r of rules.value) {
    const k = r.rule_type || 'other'
    counts.set(k, (counts.get(k) ?? 0) + 1)
  }
  const items = known.map((t) => ({ name: t.label, value: counts.get(t.key) ?? 0 }))
  items.push({
    name: '其他',
    value: [...counts.entries()]
      .filter(([k]) => !known.some((t) => t.key === k))
      .reduce((s, [, v]) => s + v, 0),
  })
  if (!items.reduce((s, it) => s + it.value, 0)) return demo()
  return { mock: false, items }
})
</script>

<template>
  <div class="bs-viewport">
    <div class="bs-screen" :style="screenStyle">
      <!-- 屏幕四角装饰 -->
      <i class="bs-frame f-tl" /><i class="bs-frame f-tr" /><i class="bs-frame f-bl" /><i class="bs-frame f-br" />

      <!-- ==================== 顶栏 ==================== -->
      <header class="bs-header">
        <div class="hd-side hd-left">
          <span class="hd-date">{{ dateText }}</span>
          <span class="hd-week">{{ weekText }}</span>
        </div>
        <div class="hd-title">
          <svg class="hd-wing" viewBox="0 0 240 20" aria-hidden="true">
            <path d="M0 10 H196 L208 4 H228 L236 10 H240" fill="none" stroke="rgba(0,212,255,0.65)" stroke-width="1.5" />
            <path d="M24 14 H170" stroke="rgba(0,212,255,0.28)" stroke-width="1" />
            <rect x="224" y="6.5" width="7" height="7" transform="rotate(45 227.5 10)" fill="#00d4ff" />
          </svg>
          <h1 class="hd-text">数据中台运营驾驶舱</h1>
          <svg class="hd-wing mirror" viewBox="0 0 240 20" aria-hidden="true">
            <path d="M0 10 H196 L208 4 H228 L236 10 H240" fill="none" stroke="rgba(0,212,255,0.65)" stroke-width="1.5" />
            <path d="M24 14 H170" stroke="rgba(0,212,255,0.28)" stroke-width="1" />
            <rect x="224" y="6.5" width="7" height="7" transform="rotate(45 227.5 10)" fill="#00d4ff" />
          </svg>
        </div>
        <div class="hd-side hd-right">
          <span class="hd-clock">{{ clockText }}</span>
          <button class="hd-btn" type="button" @click="toggleFullscreen">
            <el-icon :size="14"><Aim v-if="isFull" /><FullScreen v-else /></el-icon>
            {{ isFull ? '退出全屏' : '全屏' }}
          </button>
          <button class="hd-btn hd-btn-back" type="button" @click="backConsole">
            <el-icon :size="14"><Back /></el-icon>
            返回控制台
          </button>
        </div>
        <i class="hd-underline" />
      </header>

      <!-- 演示模式横幅：首轮加载全部失败时出现 -->
      <div v-if="booted && !anyOk" class="bs-banner">
        演示模式 · 后端数据链路未连接（未登录或服务未启动）—— 图表为骨架/占位与「演示数据」标注，控制台登录后自动接入实时数据
      </div>

      <!-- ==================== 三列主体 ==================== -->
      <main class="bs-body" :class="{ 'with-banner': booted && !anyOk }">
        <!-- 左列 -->
        <section class="bs-col">
          <div class="bs-cards">
            <div v-for="c in flipCards" :key="c.label" class="flip-card">
              <el-icon :size="24" class="flip-icon" :class="c.cls"><component :is="c.icon" /></el-icon>
              <div class="flip-meta">
                <FlipCounter :value="c.state === 'ok' ? c.value : null" class="flip-num" :class="c.cls" />
                <span class="flip-label">{{ c.label }}</span>
              </div>
            </div>
          </div>

          <ScreenCard title="近 7 日 API 调用趋势" class="h-trend">
            <template #extra>
              <span class="kv">成功率 <b>{{ callStatsState === 'ok' ? ratePct(callStats?.overall?.success_rate) : '--' }}</b></span>
              <span class="kv">均耗 <b>{{ callStatsState === 'ok' && callStats?.overall?.avg_cost_ms != null ? `${callStats.overall.avg_cost_ms}ms` : '--' }}</b></span>
              <span class="kv">缓存 <b>{{ callStatsState === 'ok' && callStats?.overall?.cache_hits != null ? callStats.overall.cache_hits : '--' }}</b></span>
            </template>
            <TrendArea
              :dates="trendDates"
              :values="trendValues"
              :loading="callStatsState === 'loading'"
              :failed="callStatsState === 'fail'"
            />
          </ScreenCard>

          <ScreenCard title="热点 API TOP8" class="h-top">
            <template #extra>
              <span class="kv">按近 7 日调用量排序</span>
            </template>
            <BarTop
              :items="topApis"
              :loading="callStatsState === 'loading'"
              :failed="callStatsState === 'fail'"
            />
          </ScreenCard>
        </section>

        <!-- 中列 -->
        <section class="bs-col">
          <ScreenCard title="表健康分总览" class="h-gauge">
            <template #extra>
              <span class="kv">画像时间 <b>{{ latestRanAt || '--' }}</b></span>
            </template>
            <GaugeRing
              :score="latestScore"
              :table-name="latestProfile?.table_name"
              :loading="profilesState === 'loading'"
              :failed="profilesState === 'fail'"
            />
          </ScreenCard>

          <ScreenCard title="表健康分 TOP5" class="h-health">
            <template #extra>
              <span class="ht-legend">
                <i class="lg-dot" style="background: #00ff88" />≥90 优
                <i class="lg-dot" style="background: #ffd855" />70-89 良
                <i class="lg-dot" style="background: #ff5e7a" />&lt;70 差
              </span>
            </template>
            <HealthTop
              :rows="healthRows"
              :loading="profilesState === 'loading'"
              :failed="profilesState === 'fail'"
            />
          </ScreenCard>

          <ScreenCard title="平台服务状态矩阵" class="h-service">
            <ServiceGrid :alive="servicesAlive" />
          </ScreenCard>
        </section>

        <!-- 右列 -->
        <section class="bs-col">
          <ScreenCard title="质量规则类型分布" :badge="ruleDist.mock ? '演示数据' : undefined" class="h-donut">
            <template #extra>
              <span class="kv">共 {{ ruleState === 'ok' ? (ruleTotal ?? 0) : '--' }} 条规则</span>
            </template>
            <!-- 失败时回退为标注的演示分布（保持大屏活性），loading 时骨架 -->
            <DonutTypes
              :items="ruleDist.items"
              :loading="ruleState === 'loading'"
            />
          </ScreenCard>

          <ScreenCard title="数据资产规模" class="h-assets">
            <AssetCounts
              :counts="assets"
              :loading="assetsState === 'loading'"
              :failed="assetsState === 'fail'"
            />
          </ScreenCard>

          <ScreenCard title="数据服务调用日志" class="h-logs">
            <template #extra>
              <span class="kv live"><i class="live-dot" />30s 自动刷新 · 轮播高亮</span>
            </template>
            <LogTicker
              :logs="logs"
              :loading="logsState === 'loading'"
              :failed="logsState === 'fail'"
            />
          </ScreenCard>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* ========== 视口与缩放容器 ========== */
.bs-viewport {
  position: fixed;
  inset: 0;
  z-index: 1000;
  overflow: hidden;
  background: #050b14;
}
.bs-screen {
  position: absolute;
  left: 50%;
  top: 50%;
  transform-origin: center center;
  display: flex;
  flex-direction: column;
  color: #d8ecff;
  font-family: 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', sans-serif;
  background:
    radial-gradient(ellipse 1200px 500px at 50% -10%, rgba(0, 212, 255, 0.12), transparent 65%),
    radial-gradient(ellipse 900px 420px at 50% 115%, rgba(0, 130, 255, 0.10), transparent 70%),
    linear-gradient(180deg, #0c1c36 0%, #0a1628 42%, #060d1a 100%);
}
/* 网格底纹 */
.bs-screen::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(rgba(0, 212, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.045) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse 90% 80% at 50% 40%, #000 30%, transparent 100%);
}
/* 屏幕四角大角标 */
.bs-frame {
  position: absolute;
  width: 26px;
  height: 26px;
  border: 0 solid rgba(0, 212, 255, 0.8);
  filter: drop-shadow(0 0 6px rgba(0, 212, 255, 0.7));
  z-index: 5;
}
.f-tl { top: 6px; left: 6px; border-top-width: 2px; border-left-width: 2px; }
.f-tr { top: 6px; right: 6px; border-top-width: 2px; border-right-width: 2px; }
.f-bl { bottom: 6px; left: 6px; border-bottom-width: 2px; border-left-width: 2px; }
.f-br { bottom: 6px; right: 6px; border-bottom-width: 2px; border-right-width: 2px; }

/* ========== 顶栏 ========== */
.bs-header {
  position: relative;
  flex-shrink: 0;
  height: 84px;
  display: flex;
  align-items: center;
  padding: 0 34px;
  background: linear-gradient(180deg, rgba(0, 212, 255, 0.10), rgba(0, 212, 255, 0));
}
.hd-underline {
  position: absolute;
  left: 24px;
  right: 24px;
  bottom: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.75) 18%, rgba(0, 212, 255, 0.75) 82%, transparent);
  filter: drop-shadow(0 0 6px rgba(0, 212, 255, 0.6));
}
.hd-underline::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 180px;
  height: 2px;
  background: linear-gradient(90deg, transparent, #7ff3ff, transparent);
  animation: hd-sweep 4s linear infinite;
}
@keyframes hd-sweep {
  0% { transform: translateX(-200px); }
  100% { transform: translateX(1900px); }
}
.hd-side {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}
.hd-right {
  justify-content: flex-end;
}
.hd-title {
  display: flex;
  align-items: center;
  gap: 18px;
  flex-shrink: 0;
}
.hd-wing {
  width: 240px;
  height: 20px;
  opacity: 0.9;
}
.hd-wing.mirror {
  transform: scaleX(-1);
}
.hd-text {
  margin: 0;
  font-size: 34px;
  font-weight: 700;
  letter-spacing: 8px;
  white-space: nowrap;
  background: linear-gradient(180deg, #eaffff 20%, #7fdcff 55%, #00d4ff 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  text-shadow: 0 0 22px rgba(0, 212, 255, 0.45);
}
.hd-date,
.hd-week {
  font-size: 14px;
  color: #7f9bbf;
  letter-spacing: 1px;
  font-family: 'Consolas', monospace;
}
.hd-clock {
  font-size: 24px;
  font-weight: 700;
  color: #7ff3ff;
  letter-spacing: 2px;
  font-family: 'DIN Alternate', 'Bahnschrift', 'Consolas', monospace;
  text-shadow: 0 0 12px rgba(0, 212, 255, 0.55);
}
.hd-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 12px;
  font-size: 12px;
  color: #bfe8ff;
  cursor: pointer;
  background: rgba(0, 212, 255, 0.08);
  border: 1px solid rgba(0, 212, 255, 0.35);
  border-radius: 3px;
  transition: all 0.2s;
  white-space: nowrap;
}
.hd-btn:hover {
  background: rgba(0, 212, 255, 0.18);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.35);
}
.hd-btn-back {
  color: #ffd855;
  border-color: rgba(255, 216, 85, 0.4);
  background: rgba(255, 216, 85, 0.06);
}
.hd-btn-back:hover {
  background: rgba(255, 216, 85, 0.14);
  box-shadow: 0 0 10px rgba(255, 216, 85, 0.3);
}

/* ========== 演示模式横幅 ========== */
.bs-banner {
  flex-shrink: 0;
  margin: 8px 24px 0;
  padding: 5px 14px;
  font-size: 13px;
  letter-spacing: 1px;
  color: #ffd855;
  border: 1px solid rgba(255, 216, 85, 0.4);
  border-radius: 3px;
  background: rgba(255, 216, 85, 0.07);
}

/* ========== 三列主体 ========== */
.bs-body {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 468px 1fr 468px;
  gap: 14px;
  padding: 14px 24px 20px;
}
.bs-body.with-banner {
  padding-top: 8px;
}
.bs-col {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 0;
  min-width: 0;
}

/* 各面板高度分配（1920 设计稿像素） */
.h-trend { height: 348px; flex-shrink: 0; }
.h-top { flex: 1; min-height: 0; }
.h-gauge { height: 292px; flex-shrink: 0; }
.h-health { height: 330px; flex-shrink: 0; }
.h-service { flex: 1; min-height: 0; }
.h-donut { height: 312px; flex-shrink: 0; }
.h-assets { height: 148px; flex-shrink: 0; }
.h-logs { flex: 1; min-height: 0; }

/* ========== 左列：数字翻牌 2×2 ========== */
.bs-cards {
  flex-shrink: 0;
  height: 232px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 8px;
}
.flip-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 18px;
  border: 1px solid rgba(0, 212, 255, 0.20);
  border-radius: 4px;
  background:
    linear-gradient(160deg, rgba(0, 212, 255, 0.10) 0%, rgba(10, 22, 40, 0.55) 70%);
  box-shadow: inset 0 0 18px rgba(0, 212, 255, 0.07), 0 0 12px rgba(0, 212, 255, 0.08);
  overflow: hidden;
}
/* 翻牌卡左下扫光 */
.flip-card::after {
  content: '';
  position: absolute;
  left: -60%;
  top: 0;
  width: 40%;
  height: 100%;
  background: linear-gradient(105deg, transparent, rgba(0, 212, 255, 0.10), transparent);
  animation: flip-sweep 6s ease-in-out infinite;
}
@keyframes flip-sweep {
  0%, 60% { left: -60%; }
  100% { left: 130%; }
}
.flip-icon {
  flex-shrink: 0;
  padding: 11px;
  box-sizing: border-box;
  border-radius: 50%;
}
.flip-icon.cyan { color: #00d4ff; background: rgba(0, 212, 255, 0.10); box-shadow: 0 0 12px rgba(0, 212, 255, 0.35); }
.flip-icon.violet { color: #7b8cff; background: rgba(123, 140, 255, 0.10); box-shadow: 0 0 12px rgba(123, 140, 255, 0.35); }
.flip-icon.green { color: #00ff88; background: rgba(0, 255, 136, 0.08); box-shadow: 0 0 12px rgba(0, 255, 136, 0.3); }
.flip-icon.yellow { color: #ffd855; background: rgba(255, 216, 85, 0.08); box-shadow: 0 0 12px rgba(255, 216, 85, 0.3); }
.flip-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.flip-meta :deep(.fc-num) {
  font-size: 32px;
  line-height: 1.1;
}
.flip-num.green { color: #00ff88; text-shadow: 0 0 12px rgba(0, 255, 136, 0.5); }
.flip-num.yellow { color: #ffd855; text-shadow: 0 0 12px rgba(255, 216, 85, 0.5); }
.flip-num.violet { color: #9faeff; text-shadow: 0 0 12px rgba(123, 140, 255, 0.5); }
.flip-label {
  font-size: 13px;
  letter-spacing: 2px;
  color: #7f9bbf;
  white-space: nowrap;
}

/* ========== 卡片 extra 槽 ========== */
.kv {
  font-size: 11px;
  color: #5a7191;
  white-space: nowrap;
}
.kv b {
  color: #a9c3e2;
  font-family: 'Consolas', monospace;
  font-weight: 600;
}
.kv.live {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #7f9bbf;
}
.live-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #00ff88;
  box-shadow: 0 0 6px rgba(0, 255, 136, 0.9);
  animation: live-blink 2s ease-in-out infinite;
}
@keyframes live-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.35; }
}
.ht-legend {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: #7f9bbf;
  white-space: nowrap;
}
.lg-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 2px;
  margin-left: 6px;
}
</style>
