<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'DatasvcDatasvc' })
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import * as echarts from 'echarts'
import {
  apiList,
  createApi,
  updateApi,
  deleteApi,
  publishApi,
  offlineApi,
  registerApp,
  grantApi,
  listCallLogs,
  getCallStats,
  getApiDoc,
  type ApiInfo,
  type AppResp,
  type ApiCallLog,
  type CallStatsResp,
  type ApiDocResp,
} from '@/api/datasvc'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共工具 ----------
function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

/** 网关路由未就绪（404 / 网关 5xx）→ 页面容错展示“功能上线中”，不弹错误提示 */
function isOffline(e: unknown): boolean {
  return e instanceof ApiError && (e.code === 404 || e.code === 502 || e.code === 503)
}

/** 字符串 ID 过滤条件 → 合法正整数才下发 */
function numOrNull(s: string): number | undefined {
  const n = Number(s)
  return s !== '' && Number.isInteger(n) && n > 0 ? n : undefined
}

/** 成功率兼容 0~1 与 0~100 两种量纲 */
function ratePct(rate?: number): string {
  if (rate === null || rate === undefined) return '-'
  const pct = rate <= 1 ? rate * 100 : rate
  return `${pct.toFixed(1)}%`
}

// ---------- 文案映射 ----------
const statusText: Record<string, string> = {
  draft: '草稿',
  published: '已发布',
  offline: '已下线',
}
const statusTagType: Record<string, 'info' | 'success' | 'danger' | 'warning'> = {
  draft: 'info',
  published: 'success',
  offline: 'danger',
}
const modeOptions = [
  { value: 'sql', label: 'SQL 模式' },
  { value: 'wizard', label: '向导模式' },
  { value: 'register', label: '注册模式' },
]

/** 调用日志状态三色映射 */
const logStatusText: Record<string, string> = { ok: '成功', error: '失败', denied: '拒绝' }
const logStatusTag: Record<string, 'success' | 'danger' | 'warning'> = {
  ok: 'success',
  error: 'danger',
  denied: 'warning',
}
const logStatusOptions: Record<string, string> = { ok: '成功', error: '失败', denied: '拒绝' }

// ==================== 页签一：API 管理 ====================
// ---------- 列表状态 ----------
const loading = ref(false)
const list = ref<ApiInfo[]>([])
const total = ref(0)
const query = reactive({ status: '', page: 1, size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const resp = await apiList({ status: query.status || undefined, page: query.page, size: query.size })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取 API 列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 数据源下拉（新建 API 时选择） ----------
const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 100 })
    dsOptions.value = resp.list || []
  } catch {
    // 数据源下拉仅辅助录入，失败不打断页面
  }
}

// ---------- 新建 / 编辑对话框 ----------
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formEl = ref<FormInstance>()
const form = reactive({
  name: '',
  path: '',
  mode: 'sql',
  datasource_id: undefined as number | undefined,
  sql_text: '',
  params_json: '',
  cache_ttl: 0,
  owner: '',
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: '请输入 API 名称', trigger: 'blur' }],
  path: [
    { required: true, message: '请输入 API 路径', trigger: 'blur' },
    {
      pattern: /^\S+$/,
      message: '路径不能包含空白字符',
      trigger: 'blur',
    },
  ],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  sql_text: [{ required: true, message: '请输入 SQL 查询语句', trigger: 'blur' }],
})

/** JSON 字段校验：非必填，但填了必须能 JSON.parse */
function jsonValidator(field: string) {
  return (_rule: unknown, value: string, callback: (err?: Error) => void) => {
    const s = (value || '').trim()
    if (!s) return callback()
    try {
      JSON.parse(s)
      callback()
    } catch {
      callback(new Error(`${field} 必须是合法的 JSON`))
    }
  }
}

const extraRules = reactive<FormRules>({
  params_json: [{ validator: jsonValidator('params_json'), trigger: 'blur' }],
})

function resetForm() {
  Object.assign(form, {
    name: '',
    path: '',
    mode: 'sql',
    datasource_id: undefined,
    sql_text: '',
    params_json: '',
    cache_ttl: 0,
    owner: '',
  })
}

function openCreate() {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: ApiInfo) {
  editingId.value = row.id ?? null
  Object.assign(form, {
    name: row.name || '',
    path: row.path || '',
    mode: row.mode || 'sql',
    datasource_id: row.datasource_id,
    sql_text: row.sql_text || '',
    params_json: row.params_json || '',
    cache_ttl: row.cache_ttl || 0,
    owner: row.owner || '',
  })
  dialogVisible.value = true
}

async function onSave() {
  const valid = await formEl.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (editingId.value === null) {
      await createApi({
        name: form.name.trim(),
        path: form.path.trim(),
        mode: form.mode,
        datasource_id: Number(form.datasource_id),
        sql_text: form.sql_text,
        params_json: form.params_json.trim() || undefined,
        cache_ttl: form.cache_ttl || 0,
        owner: form.owner.trim() || undefined,
      })
      ElMessage.success('创建成功，状态为草稿，发布后可被调用')
    } else {
      // 更新接口仅支持 name/sql_text/params_json/cache_ttl，其余字段创建后不可改
      await updateApi({
        id: editingId.value,
        name: form.name.trim(),
        sql_text: form.sql_text,
        params_json: form.params_json.trim() || undefined,
        cache_ttl: form.cache_ttl || 0,
      })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

// ---------- 发布 / 下线 / 删除 ----------
async function onPublish(row: ApiInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定发布 API「${row.name}」吗？发布后外部应用可凭授权调用。`,
    '发布确认',
    { confirmButtonText: '发布', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await publishApi(row.id)
    ElMessage.success('发布成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '发布失败')
  }
}

async function onOffline(row: ApiInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定下线 API「${row.name}」吗？下线后调用将被拒绝。`,
    '下线确认',
    { confirmButtonText: '下线', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await offlineApi(row.id)
    ElMessage.success('下线成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '下线失败')
  }
}

async function onDelete(row: ApiInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除 API「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await deleteApi(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ---------- 注册应用对话框 ----------
const appVisible = ref(false)
const appSaving = ref(false)
const appFormEl = ref<FormInstance>()
const appForm = reactive({ name: '', rate_limit_qps: 10, ip_whitelist: '' })
const appResult = ref<AppResp | null>(null)

const appRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
})

function openApp() {
  appResult.value = null
  Object.assign(appForm, { name: '', rate_limit_qps: 10, ip_whitelist: '' })
  appVisible.value = true
}

async function onRegisterApp() {
  const valid = await appFormEl.value?.validate().catch(() => false)
  if (!valid) return
  appSaving.value = true
  try {
    const resp = await registerApp({
      name: appForm.name.trim(),
      rate_limit_qps: appForm.rate_limit_qps || undefined,
      ip_whitelist: appForm.ip_whitelist.trim() || undefined,
    })
    appResult.value = resp
    ElMessage.success('注册成功，app_secret 仅展示一次，请立即保存')
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '注册应用失败')
  } finally {
    appSaving.value = false
  }
}

async function copyText(text: string | undefined, label: string) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label} 已复制`)
  } catch {
    ElMessage.error('复制失败，请手动选择复制')
  }
}

// ---------- 授权对话框 ----------
const grantVisible = ref(false)
const grantSaving = ref(false)
const grantFormEl = ref<FormInstance>()
const grantForm = reactive({
  api_id: undefined as number | undefined,
  app_id: undefined as number | undefined,
  field_mask: '',
  forever: true,
  expire_at: '',
})

const grantRules = reactive<FormRules>({
  api_id: [{ required: true, message: '请选择 API', trigger: 'change' }],
  app_id: [{ required: true, message: '请输入应用 ID', trigger: 'blur' }],
})

/** 授权下拉的 API 候选（独立拉取前 100 条，不依赖当前分页） */
const grantApiOptions = ref<ApiInfo[]>([])
async function fetchGrantOptions() {
  try {
    const resp = await apiList({ page: 1, size: 100 })
    grantApiOptions.value = resp.list || []
  } catch {
    grantApiOptions.value = []
  }
}

function openGrant() {
  Object.assign(grantForm, {
    api_id: undefined,
    app_id: undefined,
    field_mask: '',
    forever: true,
    expire_at: '',
  })
  grantApiOptions.value = []
  fetchGrantOptions()
  grantVisible.value = true
}

async function onGrant() {
  const valid = await grantFormEl.value?.validate().catch(() => false)
  if (!valid) return
  if (!grantForm.forever && !grantForm.expire_at) {
    ElMessage.warning('请选择过期时间，或勾选永久有效')
    return
  }
  grantSaving.value = true
  try {
    const expireTs = grantForm.forever
      ? 0
      : Math.floor(new Date(grantForm.expire_at).getTime() / 1000)
    await grantApi({
      api_id: Number(grantForm.api_id),
      app_id: Number(grantForm.app_id),
      expire_at: expireTs,
      field_mask: grantForm.field_mask.trim() || undefined,
    })
    ElMessage.success(grantForm.forever ? '授权成功（永久有效）' : '授权成功')
    grantVisible.value = false
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '授权失败')
  } finally {
    grantSaving.value = false
  }
}

// ==================== 页签二：调用统计 ====================
const statsDays = ref(7)
const statsLoading = ref(false)
const statsOffline = ref(false)
const statsData = ref<CallStatsResp | null>(null)

const trendEl = ref<HTMLDivElement>()
const topEl = ref<HTMLDivElement>()
let trendChart: echarts.ECharts | null = null
let topChart: echarts.ECharts | null = null

const hasDaily = computed(() => (statsData.value?.daily_dates?.length ?? 0) > 0)
const hasTop = computed(() => (statsData.value?.top_apis?.length ?? 0) > 0)

/** 指标卡：总调用 / 成功率 / 平均耗时 / 缓存命中 */
const metricCards = computed(() => {
  const o = statsData.value?.overall
  return [
    { label: '总调用次数', value: o?.total_calls ?? '-' },
    { label: '成功率', value: ratePct(o?.success_rate) },
    { label: '平均耗时', value: o?.avg_cost_ms !== undefined && o?.avg_cost_ms !== null ? `${o.avg_cost_ms} ms` : '-' },
    { label: '缓存命中次数', value: o?.cache_hits ?? '-' },
  ]
})

async function fetchStats() {
  statsLoading.value = true
  try {
    const resp = await getCallStats(statsDays.value)
    statsData.value = resp
    statsOffline.value = false
    renderTrend()
    renderTop()
  } catch (e) {
    if (isOffline(e)) {
      statsOffline.value = true
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '获取调用统计失败')
    }
  } finally {
    statsLoading.value = false
  }
}

/** 日趋势折线（daily_calls） */
function renderTrend() {
  nextTick(() => {
    const el = trendEl.value
    const d = statsData.value
    if (!el || !d?.daily_dates?.length) return
    trendChart?.dispose()
    trendChart = echarts.init(el)
    trendChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 50, right: 20, bottom: 26, top: 30 },
      xAxis: { type: 'category', boundaryGap: false, data: d.daily_dates },
      yAxis: { type: 'value', minInterval: 1 },
      series: [
        {
          name: '调用量',
          type: 'line',
          smooth: true,
          symbolSize: 6,
          data: d.daily_calls || [],
          areaStyle: { opacity: 0.15 },
          itemStyle: { color: '#409eff' },
        },
      ],
    })
  })
}

/** TopN API 横向条形图 */
function renderTop() {
  nextTick(() => {
    const el = topEl.value
    const tops = statsData.value?.top_apis || []
    if (!el || !tops.length) return
    topChart?.dispose()
    topChart = echarts.init(el)
    // 升序排列使调用量最大的 API 显示在最上方，最多展示 10 条
    const sorted = [...tops].sort((a, b) => (a.calls ?? 0) - (b.calls ?? 0)).slice(-10)
    topChart.setOption({
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params: unknown) => {
          const arr = (Array.isArray(params) ? params : [params]) as {
            dataIndex?: number
            name?: string
            value?: number
          }[]
          const p = arr[0]
          if (!p) return ''
          const item = sorted[p.dataIndex ?? -1]
          const cost = item?.avg_cost_ms
          return [
            `<b>${p.name || ''}</b>`,
            `调用量：${p.value ?? 0}`,
            cost !== undefined && cost !== null ? `平均耗时：${cost} ms` : '',
          ]
            .filter(Boolean)
            .join('<br/>')
        },
      },
      grid: { left: 8, right: 40, bottom: 26, top: 30, containLabel: true },
      xAxis: { type: 'value', minInterval: 1 },
      yAxis: {
        type: 'category',
        data: sorted.map((t) => t.path || `#${t.api_id ?? '-'}`),
        axisLabel: { width: 130, overflow: 'truncate' },
      },
      series: [
        {
          name: '调用量',
          type: 'bar',
          barMaxWidth: 18,
          data: sorted.map((t) => t.calls ?? 0),
          itemStyle: { color: '#67c23a', borderRadius: [0, 4, 4, 0] },
          label: { show: true, position: 'right' },
        },
      ],
    })
  })
}

// ==================== 页签三：调用日志 ====================
const logLoading = ref(false)
const logOffline = ref(false)
const logList = ref<ApiCallLog[]>([])
const logTotal = ref(0)
const logQuery = reactive({ api_id: '', app_id: '', status: '', page: 1, size: 10 })

async function fetchCallLogs() {
  logLoading.value = true
  try {
    const resp = await listCallLogs({
      api_id: numOrNull(logQuery.api_id),
      app_id: numOrNull(logQuery.app_id),
      status: logQuery.status || undefined,
      page: logQuery.page,
      size: logQuery.size,
    })
    logList.value = resp.list || []
    logTotal.value = resp.total || 0
    logOffline.value = false
  } catch (e) {
    if (isOffline(e)) {
      logOffline.value = true
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '获取调用日志失败')
    }
  } finally {
    logLoading.value = false
  }
}

function onLogSearch() {
  logQuery.page = 1
  fetchCallLogs()
}

// ==================== API 文档弹窗 ====================
const docVisible = ref(false)
const docLoading = ref(false)
const docOffline = ref(false)
const doc = ref<ApiDocResp | null>(null)
const docRow = ref<ApiInfo | null>(null)

/** 简易 Markdown → HTML（先整体转义防注入，再识别标题/代码块/表格/列表/加粗/行内代码） */
function renderMarkdown(md?: string): string {
  if (!md) return ''
  const esc = (s: string) =>
    s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
  const inline = (s: string) =>
    esc(s)
      .replace(/`([^`]+)`/g, '<code class="md-code-inline">$1</code>')
      .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')

  const lines = md.replace(/\r\n/g, '\n').split('\n')
  const out: string[] = []
  let inCode = false
  let tableBuf: string[] = []

  const flushTable = () => {
    if (!tableBuf.length) return
    const cells = (line: string) =>
      line
        .replace(/^\|/, '')
        .replace(/\|$/, '')
        .split('|')
        .map((c) => c.trim())
    const rows = tableBuf.map(cells)
    const isSep = (r: string[]) => r.length > 0 && r.every((c) => /^:?-{1,}:?$/.test(c))
    const bodyRows =
      rows.length > 1 && isSep(rows[1]) ? [rows[0], ...rows.slice(2)] : rows
    const head = bodyRows[0]
    const body = bodyRows.slice(1)
    const html = ['<table class="md-table"><thead><tr>']
    if (head) for (const c of head) html.push(`<th>${inline(c)}</th>`)
    html.push('</tr></thead><tbody>')
    for (const r of body) {
      html.push('<tr>')
      for (const c of r) html.push(`<td>${inline(c)}</td>`)
      html.push('</tr>')
    }
    html.push('</tbody></table>')
    out.push(html.join(''))
    tableBuf = []
  }

  for (const line of lines) {
    const t = line.trim()
    if (t.startsWith('```')) {
      flushTable()
      if (inCode) {
        out.push('</code></pre>')
        inCode = false
      } else {
        out.push('<pre class="md-code"><code>')
        inCode = true
      }
      continue
    }
    if (inCode) {
      out.push(esc(line))
      continue
    }
    if (t.startsWith('|')) {
      tableBuf.push(t)
      continue
    }
    flushTable()
    if (!t) continue
    if (t.startsWith('### ')) out.push(`<h4 class="md-h">${inline(t.slice(4))}</h4>`)
    else if (t.startsWith('## ')) out.push(`<h3 class="md-h">${inline(t.slice(3))}</h3>`)
    else if (t.startsWith('# ')) out.push(`<h2 class="md-h md-h1">${inline(t.slice(2))}</h2>`)
    else if (/^[-*] /.test(t)) out.push(`<div class="md-li">• ${inline(t.slice(2))}</div>`)
    else if (/^\d+\. /.test(t)) out.push(`<div class="md-li">${inline(t)}</div>`)
    else out.push(`<p class="md-p">${inline(t)}</p>`)
  }
  flushTable()
  if (inCode) out.push('</code></pre>')
  return out.join('\n')
}

const docHtml = computed(() => renderMarkdown(doc.value?.markdown))

async function openDoc(row: ApiInfo) {
  if (!row.path) {
    ElMessage.warning('该 API 缺少路径，无法生成文档')
    return
  }
  docRow.value = row
  doc.value = null
  docOffline.value = false
  docVisible.value = true
  docLoading.value = true
  try {
    doc.value = await getApiDoc(row.path)
  } catch (e) {
    if (isOffline(e)) {
      docOffline.value = true
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '获取 API 文档失败')
    }
  } finally {
    docLoading.value = false
  }
}

// ---------- 页签切换：懒加载 ----------
const activeTab = ref('api')
function onTabChange() {
  switch (activeTab.value) {
    case 'stats':
      if (!statsData.value && !statsLoading.value && !statsOffline.value) fetchStats()
      break
    case 'logs':
      if (!logList.value.length && !logLoading.value && !logOffline.value) fetchCallLogs()
      break
  }
}

function onResize() {
  trendChart?.resize()
  topChart?.resize()
}

onMounted(() => {
  fetchList()
  fetchDsOptions()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  trendChart?.dispose()
  trendChart = null
  topChart?.dispose()
  topChart = null
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ==================== API 管理 ==================== -->
        <el-tab-pane label="API 管理" name="api">
          <!-- 工具栏 -->
          <div class="toolbar">
            <el-select
              v-model="query.status"
              placeholder="全部状态"
              clearable
              style="width: 140px"
              @change="onSearch"
            >
              <el-option v-for="(label, key) in statusText" :key="key" :label="label" :value="key" />
            </el-select>
            <el-button type="primary" @click="openCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建 API
            </el-button>
            <el-button type="success" @click="openApp">
              <el-icon><Key /></el-icon>&nbsp;注册应用
            </el-button>
            <el-button type="warning" @click="openGrant">
              <el-icon><Lock /></el-icon>&nbsp;授权
            </el-button>
            <el-button @click="fetchList">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <!-- 列表 -->
          <el-table v-loading="loading" :data="list" stripe border>
            <el-table-column prop="id" label="ID" width="64" />
            <el-table-column prop="name" label="API 名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="path" label="路径" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="mono">/datasvc/execute?path={{ row.path }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="mode" label="模式" width="90">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ row.mode }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="datasource_id" label="数据源" width="80" align="center" />
            <el-table-column prop="sql_text" label="SQL" min-width="180" show-overflow-tooltip />
            <el-table-column prop="cache_ttl" label="缓存 TTL(s)" width="100" align="center">
              <template #default="{ row }">{{ row.cache_ttl || '-' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType[row.status] || 'info'" size="small">
                  {{ statusText[row.status] || row.status || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="owner" label="负责人" width="100" show-overflow-tooltip />
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
                <el-button link type="primary" @click="openDoc(row)">文档</el-button>
                <el-button
                  v-if="row.status !== 'published'"
                  link
                  type="success"
                  @click="onPublish(row)"
                >
                  发布
                </el-button>
                <el-button v-if="row.status === 'published'" link type="warning" @click="onOffline(row)">
                  下线
                </el-button>
                <el-button link type="danger" @click="onDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.size"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchList"
            @size-change="onSearch"
          />
        </el-tab-pane>

        <!-- ==================== 调用统计 ==================== -->
        <el-tab-pane label="调用统计" name="stats">
          <el-alert
            v-if="statsOffline"
            title="调用统计接口暂未就绪（功能上线中），后端路由合入后即可查看"
            type="info"
            show-icon
            :closable="false"
            class="offline-alert"
          />
          <div v-else class="toolbar">
            <el-radio-group v-model="statsDays" @change="fetchStats">
              <el-radio-button :value="7">近 7 天</el-radio-button>
              <el-radio-button :value="14">近 14 天</el-radio-button>
              <el-radio-button :value="30">近 30 天</el-radio-button>
            </el-radio-group>
            <el-button @click="fetchStats">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-empty v-if="statsOffline" description="调用统计 · 功能上线中" :image-size="90" />
          <div v-else v-loading="statsLoading">
            <!-- 指标卡 -->
            <el-row :gutter="16" class="metric-row">
              <el-col v-for="m in metricCards" :key="m.label" :xs="12" :sm="6">
                <div class="metric-card">
                  <div class="metric-value">{{ m.value }}</div>
                  <div class="metric-label">{{ m.label }}</div>
                </div>
              </el-col>
            </el-row>

            <el-row :gutter="16">
              <el-col :xs="24" :lg="14">
                <el-card shadow="never">
                  <template #header><span class="card-title">每日调用趋势</span></template>
                  <div v-show="hasDaily" ref="trendEl" class="chart-lg" />
                  <el-empty
                    v-if="!hasDaily && !statsLoading"
                    description="暂无调用数据：API 被调用后将在此展示日趋势"
                    :image-size="80"
                  />
                </el-card>
              </el-col>
              <el-col :xs="24" :lg="10">
                <el-card shadow="never">
                  <template #header><span class="card-title">Top API 调用量</span></template>
                  <div v-show="hasTop" ref="topEl" class="chart-lg" />
                  <el-empty
                    v-if="!hasTop && !statsLoading"
                    description="暂无 API 调用记录"
                    :image-size="80"
                  />
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>

        <!-- ==================== 调用日志 ==================== -->
        <el-tab-pane label="调用日志" name="logs">
          <el-alert
            v-if="logOffline"
            title="调用日志接口暂未就绪（功能上线中），后端路由合入后即可查询"
            type="info"
            show-icon
            :closable="false"
            class="offline-alert"
          />
          <div v-else class="toolbar">
            <el-input
              v-model="logQuery.api_id"
              placeholder="API ID"
              clearable
              style="width: 120px"
              @keyup.enter="onLogSearch"
              @clear="onLogSearch"
            />
            <el-input
              v-model="logQuery.app_id"
              placeholder="应用 App ID"
              clearable
              style="width: 140px"
              @keyup.enter="onLogSearch"
              @clear="onLogSearch"
            />
            <el-select
              v-model="logQuery.status"
              placeholder="全部状态"
              clearable
              style="width: 130px"
              @change="onLogSearch"
            >
              <el-option v-for="(label, key) in logStatusOptions" :key="key" :label="label" :value="key" />
            </el-select>
            <el-button type="primary" @click="onLogSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button @click="fetchCallLogs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-empty v-if="logOffline" description="调用日志 · 功能上线中" :image-size="90" />
          <template v-else>
            <el-table v-loading="logLoading" :data="logList" stripe border>
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="api_id" label="API ID" width="80" align="center" />
              <el-table-column prop="app_id" label="应用 ID" width="80" align="center" />
              <el-table-column prop="path" label="路径" min-width="160" show-overflow-tooltip />
              <el-table-column label="状态" width="90" align="center">
                <template #default="{ row }">
                  <el-tag :type="logStatusTag[row.status] || 'info'" size="small">
                    {{ logStatusText[row.status] || row.status || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="错误码" width="110" show-overflow-tooltip>
                <template #default="{ row }">{{ row.error_code || '-' }}</template>
              </el-table-column>
              <el-table-column prop="rows_cnt" label="返回行数" width="90" align="right" />
              <el-table-column label="耗时" width="90" align="right">
                <template #default="{ row }">
                  {{ row.cost_ms !== undefined && row.cost_ms !== null ? `${row.cost_ms} ms` : '-' }}
                </template>
              </el-table-column>
              <el-table-column label="缓存" width="80" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.cache_hit" type="success" size="small" effect="plain">命中</el-tag>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column label="调用 IP" width="130" show-overflow-tooltip>
                <template #default="{ row }">{{ row.ip || '-' }}</template>
              </el-table-column>
              <el-table-column label="调用时间" width="170">
                <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
              </el-table-column>
            </el-table>
            <el-empty
              v-if="!logLoading && !logList.length"
              description="暂无调用日志"
              :image-size="90"
            />

            <el-pagination
              v-model:current-page="logQuery.page"
              v-model:page-size="logQuery.size"
              :total="logTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              class="pager"
              @current-change="fetchCallLogs"
              @size-change="onLogSearch"
            />
          </template>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingId === null ? '新建 API' : '编辑 API'"
      width="640px"
      destroy-on-close
    >
      <el-form ref="formEl" :model="form" :rules="{ ...rules, ...extraRules }" label-width="100px">
        <el-form-item label="API 名称" prop="name">
          <el-input v-model="form.name" placeholder="如：订单日汇总查询" maxlength="64" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" :disabled="editingId !== null" placeholder="如 /api/order/daily" />
        </el-form-item>
        <el-form-item label="模式" prop="mode">
          <el-select v-model="form.mode" :disabled="editingId !== null" style="width: 100%">
            <el-option v-for="m in modeOptions" :key="m.value" :label="m.label" :value="m.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据源" prop="datasource_id">
          <el-select
            v-model="form.datasource_id"
            :disabled="editingId !== null"
            placeholder="选择数据源"
            style="width: 100%"
          >
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（ID:${d.id} · ${d.type}）`"
              :value="d.id!"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="SQL 语句" prop="sql_text">
          <el-input
            v-model="form.sql_text"
            type="textarea"
            :rows="5"
            placeholder="查询 SQL，如：select order_date, sum(amount) from orders group by order_date"
          />
        </el-form-item>
        <el-form-item label="参数定义" prop="params_json">
          <el-input
            v-model="form.params_json"
            type="textarea"
            :rows="3"
            placeholder='可选，入参定义 JSON，如：{"date":"string","user_id":"int64"}'
          />
        </el-form-item>
        <el-form-item label="缓存 TTL(s)">
          <el-input-number v-model="form.cache_ttl" :min="0" :max="86400" />
          <span class="form-tip">0 表示不缓存</span>
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="form.owner" :disabled="editingId !== null" placeholder="如：data-team" />
        </el-form-item>
      </el-form>
      <el-alert
        v-if="editingId !== null"
        title="路径 / 模式 / 数据源 / 负责人 创建后不可修改，仅可更新名称、SQL、参数与缓存"
        type="info"
        :closable="false"
        show-icon
        class="form-alert"
      />
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 注册应用对话框 -->
    <el-dialog v-model="appVisible" title="注册调用应用" width="560px" destroy-on-close>
      <el-form ref="appFormEl" :model="appForm" :rules="appRules" label-width="100px">
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="appForm.name" placeholder="如：报表中心" maxlength="64" />
        </el-form-item>
        <el-form-item label="QPS 限流">
          <el-input-number v-model="appForm.rate_limit_qps" :min="1" :max="100000" />
        </el-form-item>
        <el-form-item label="IP 白名单">
          <el-input v-model="appForm.ip_whitelist" placeholder="可选，逗号分隔，如 10.0.0.1,10.0.0.2" />
        </el-form-item>
      </el-form>
      <template v-if="appResult">
        <el-alert
          title="app_secret 仅在注册时返回一次，请立即复制保存"
          type="warning"
          :closable="false"
          show-icon
          class="form-alert"
        />
        <el-descriptions :column="1" border class="app-result">
          <el-descriptions-item label="App ID">{{ appResult.app_id }}</el-descriptions-item>
          <el-descriptions-item label="App Key">
            <div class="secret-row">
              <span class="mono">{{ appResult.app_key }}</span>
              <el-button link type="primary" @click="copyText(appResult.app_key, 'App Key')">
                复制
              </el-button>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="App Secret">
            <div class="secret-row">
              <span class="mono">{{ appResult.app_secret }}</span>
              <el-button link type="primary" @click="copyText(appResult.app_secret, 'App Secret')">
                复制
              </el-button>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="appVisible = false">关闭</el-button>
        <el-button v-if="!appResult" type="primary" :loading="appSaving" @click="onRegisterApp">
          注册
        </el-button>
      </template>
    </el-dialog>

    <!-- 授权对话框 -->
    <el-dialog v-model="grantVisible" title="授权 API 给应用" width="560px" destroy-on-close>
      <el-form ref="grantFormEl" :model="grantForm" :rules="grantRules" label-width="100px">
        <el-form-item label="API" prop="api_id">
          <el-select
            v-model="grantForm.api_id"
            filterable
            :loading="grantApiOptions.length === 0"
            placeholder="选择要授权的 API"
            style="width: 100%"
          >
            <el-option
              v-for="a in grantApiOptions"
              :key="a.id"
              :label="`#${a.id} ${a.name}（${a.path}）`"
              :value="a.id!"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="应用 ID" prop="app_id">
          <el-input-number
            v-model="grantForm.app_id"
            :min="1"
            placeholder="注册应用时返回的 App ID"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="字段脱敏">
          <el-input
            v-model="grantForm.field_mask"
            placeholder='可选，返回列脱敏 JSON，如：{"phone":"mask"}'
          />
        </el-form-item>
        <el-form-item label="有效期">
          <el-checkbox v-model="grantForm.forever">永久有效</el-checkbox>
        </el-form-item>
        <el-form-item v-if="!grantForm.forever" label="过期时间">
          <el-date-picker
            v-model="grantForm.expire_at"
            type="datetime"
            placeholder="选择过期时间"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" :loading="grantSaving" @click="onGrant">确认授权</el-button>
      </template>
    </el-dialog>

    <!-- API 文档对话框 -->
    <el-dialog
      v-model="docVisible"
      :title="`API 文档 · ${doc?.api_name || docRow?.name || ''}`"
      width="780px"
      top="6vh"
    >
      <div v-loading="docLoading" class="doc-body">
        <el-empty
          v-if="docOffline"
          description="API 文档接口上线中，后端路由合入后可查看"
          :image-size="90"
        />
        <template v-else-if="doc">
          <el-descriptions :column="2" border size="small" class="doc-head">
            <el-descriptions-item label="API 名称">{{ doc.api_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="路径">
              <span class="mono">{{ doc.path || '-' }}</span>
            </el-descriptions-item>
          </el-descriptions>
          <!-- 简易渲染的 Markdown（标题 / 表格 / 代码块 / 列表） -->
          <div class="doc-md" v-html="docHtml" />
          <el-collapse v-if="doc.openapi_json" class="doc-openapi">
            <el-collapse-item name="openapi">
              <template #title>
                <span class="openapi-title">
                  OpenAPI 3.0 定义
                  <el-button
                    link
                    type="primary"
                    @click.stop="copyText(doc.openapi_json || '', 'OpenAPI JSON')"
                  >
                    复制 JSON
                  </el-button>
                </span>
              </template>
              <pre class="raw-pre">{{ doc.openapi_json }}</pre>
            </el-collapse-item>
          </el-collapse>
        </template>
        <el-empty v-else-if="!docLoading" description="暂无文档" :image-size="90" />
      </div>
      <template #footer>
        <el-button @click="docVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
.form-tip {
  margin-left: 10px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.form-alert {
  margin-top: 4px;
}
.app-result {
  margin-top: 12px;
}
.secret-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.mono {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
}
.offline-alert {
  margin-bottom: 14px;
}
.metric-row {
  margin-bottom: 16px;
}
.metric-card {
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  text-align: center;
  padding: 16px 8px;
}
.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  line-height: 1.3;
}
.metric-label {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
.chart-lg {
  width: 100%;
  height: 300px;
}
.card-title {
  font-weight: 600;
  font-size: 14px;
}
.doc-body {
  min-height: 200px;
  max-height: 66vh;
  overflow: auto;
}
.doc-head {
  margin-bottom: 12px;
}
.doc-openapi {
  margin-top: 12px;
}
.openapi-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 600;
}
.raw-pre {
  max-height: 320px;
  overflow: auto;
  background: #f5f7fa;
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  padding: 12px;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}
/* v-html 注入的 Markdown 节点无 scoped 属性，需 :deep 穿透 */
.doc-md :deep(.md-h) {
  margin: 14px 0 6px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}
.doc-md :deep(.md-h1) {
  font-size: 17px;
}
.doc-md :deep(.md-p) {
  margin: 6px 0;
  font-size: 13px;
  line-height: 1.7;
  color: #606266;
}
.doc-md :deep(.md-li) {
  margin: 2px 0 2px 8px;
  font-size: 13px;
  line-height: 1.7;
  color: #606266;
}
.doc-md :deep(.md-table) {
  border-collapse: collapse;
  width: 100%;
  margin: 10px 0;
  font-size: 13px;
}
.doc-md :deep(.md-table th),
.doc-md :deep(.md-table td) {
  border: 1px solid var(--el-border-color-lighter);
  padding: 6px 10px;
  text-align: left;
}
.doc-md :deep(.md-table th) {
  background: var(--el-fill-color-lighter);
  font-weight: 600;
}
.doc-md :deep(.md-code) {
  background: #f5f7fa;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 12px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 8px 0;
}
.doc-md :deep(.md-code-inline) {
  background: var(--el-fill-color-light);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 12px;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  color: #c7254e;
}
</style>
