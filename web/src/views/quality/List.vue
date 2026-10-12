<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import * as echarts from 'echarts'
import {
  createQualityRule,
  deleteQualityRule,
  qualityRuleList,
  createQualityJob,
  qualityJobList,
  runQualityJob,
  qualityResultList,
  saveAlertChannel,
  deleteAlertChannel,
  listAlertChannels,
  testAlertChannel,
  listErrorData,
  profileTable,
  listProfiles,
  type QualityRule,
  type QualityJob,
  type QualityResult,
  type RunJobResp,
  type AlertChannel,
  type ErrorData,
  type TableProfile,
  type ProfileReply,
  type TestAlertResp,
} from '@/api/quality'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共工具 ----------
function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

function fmtDate(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleDateString('zh-CN')
}

/** 网关路由未就绪（404 / 网关 5xx）→ 页面容错展示“功能上线中”，不弹错误提示 */
function isOffline(e: unknown): boolean {
  return e instanceof ApiError && (e.code === 404 || e.code === 502 || e.code === 503)
}

/** 健康分配色：<60 红 / 60-80 黄 / >=80 绿 */
function scoreColor(score?: number): string {
  const s = score ?? 0
  if (s < 60) return '#f56c6c'
  if (s < 80) return '#e6a23c'
  return '#67c23a'
}

function scoreTagType(score?: number): 'danger' | 'warning' | 'success' {
  const s = score ?? 0
  if (s < 60) return 'danger'
  if (s < 80) return 'warning'
  return 'success'
}

/**
 * 解析 rows_json / profile_json 为动态列表格数据（参考 AiChat.vue 动态列做法）。
 * 兼容：对象数组 / 基础类型数组 / {rows:[...]} / {columns:[...]} / {list:[...]} / 单层键值对象
 */
function parseJsonRows(raw?: string): { columns: string[]; rows: Record<string, unknown>[] } {
  if (!raw) return { columns: [], rows: [] }
  try {
    const v: unknown = JSON.parse(raw)
    if (Array.isArray(v)) {
      const rows = v.filter(
        (x): x is Record<string, unknown> => !!x && typeof x === 'object' && !Array.isArray(x),
      )
      if (rows.length) {
        const cols: string[] = []
        for (const r of rows) {
          for (const k of Object.keys(r)) if (!cols.includes(k)) cols.push(k)
        }
        return { columns: cols, rows }
      }
      return { columns: ['value'], rows: v.map((x) => ({ value: x as unknown })) }
    }
    if (v && typeof v === 'object') {
      const obj = v as Record<string, unknown>
      if (Array.isArray(obj.rows)) return parseJsonRows(JSON.stringify(obj.rows))
      if (Array.isArray(obj.columns)) return parseJsonRows(JSON.stringify(obj.columns))
      if (Array.isArray(obj.list)) return parseJsonRows(JSON.stringify(obj.list))
      return {
        columns: ['field', 'value'],
        rows: Object.entries(obj).map(([field, value]) => ({
          field,
          value: value !== null && typeof value === 'object' ? JSON.stringify(value) : (value as unknown),
        })),
      }
    }
    return { columns: [], rows: [] }
  } catch {
    return { columns: [], rows: [] }
  }
}

function formatCell(v: unknown): string {
  if (v === null || v === undefined || v === '') return '-'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

/** JSON 字段校验：填了必须能 JSON.parse */
function jsonValidator(field: string, required = false) {
  return (_rule: unknown, value: string, callback: (err?: Error) => void) => {
    const s = (value || '').trim()
    if (!s) return required ? callback(new Error(`请输入 ${field}`)) : callback()
    try {
      JSON.parse(s)
      callback()
    } catch {
      callback(new Error(`${field} 必须是合法的 JSON`))
    }
  }
}

// ---------- 文案映射 ----------
const ruleTypeOptions = [
  { value: 'row_count', label: '行数检查' },
  { value: 'not_null', label: '非空检查' },
  { value: 'unique', label: '唯一性检查' },
  { value: 'value_range', label: '值域检查' },
  { value: 'consistency', label: '一致性检查' },
  { value: 'fluctuation', label: '波动率检查' },
]
const ruleTypeText: Record<string, string> = Object.fromEntries(
  ruleTypeOptions.map((o) => [o.value, o.label]),
)
/** 行内安全取文案（字段可能缺失） */
function ruleTypeLabel(t?: string) {
  return (t && ruleTypeText[t]) || t || '-'
}
/** 这些类型需要目标列（后端 BuildCheckSQL 要拼接列名） */
const columnRequiredTypes = ['not_null', 'unique', 'value_range', 'consistency']

/** 按规则类型提示 params_json 格式 */
const paramsPlaceholder = computed(() => {
  switch (ruleForm.rule_type) {
    case 'value_range':
      return '必填：{"min":0,"max":100}'
    case 'consistency':
      return '必填：{"ref_table":"dim_user","ref_column":"id"}'
    case 'fluctuation':
      return '必填：{"baseline":10000,"ratio_pct":5}'
    default:
      return '该类型无需参数，可留空'
  }
})

// ==================== 页签一：质量规则 ====================
const ruleLoading = ref(false)
const ruleList = ref<QualityRule[]>([])
const ruleTotal = ref(0)
const ruleQuery = reactive({ rule_type: '', page: 1, size: 10 })

async function fetchRules() {
  ruleLoading.value = true
  try {
    const resp = await qualityRuleList({
      rule_type: ruleQuery.rule_type || undefined,
      page: ruleQuery.page,
      size: ruleQuery.size,
    })
    ruleList.value = resp.list || []
    ruleTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取规则列表失败')
  } finally {
    ruleLoading.value = false
  }
}

function onRuleSearch() {
  ruleQuery.page = 1
  fetchRules()
}

// ---------- 新建规则对话框 ----------
const ruleDialogVisible = ref(false)
const ruleSaving = ref(false)
const ruleFormEl = ref<FormInstance>()
const ruleForm = reactive({
  name: '',
  rule_type: 'not_null',
  target_table: '',
  target_column: '',
  params_json: '',
  threshold_json: '',
  owner: '',
})

const ruleRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  rule_type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  target_table: [{ required: true, message: '请输入目标表名', trigger: 'blur' }],
  target_column: [
    {
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (columnRequiredTypes.includes(ruleForm.rule_type) && !value.trim()) {
          return callback(new Error('该规则类型必须填写目标列'))
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  params_json: [{ validator: jsonValidator('params_json'), trigger: 'blur' }],
  threshold_json: [{ validator: jsonValidator('threshold_json'), trigger: 'blur' }],
})

function openRuleCreate() {
  Object.assign(ruleForm, {
    name: '',
    rule_type: 'not_null',
    target_table: '',
    target_column: '',
    params_json: '',
    threshold_json: '',
    owner: '',
  })
  ruleDialogVisible.value = true
}

async function onSaveRule() {
  const valid = await ruleFormEl.value?.validate().catch(() => false)
  if (!valid) return
  ruleSaving.value = true
  try {
    await createQualityRule({
      name: ruleForm.name.trim(),
      rule_type: ruleForm.rule_type,
      target_table: ruleForm.target_table.trim(),
      target_column: ruleForm.target_column.trim() || undefined,
      params_json: ruleForm.params_json.trim() || undefined,
      threshold_json: ruleForm.threshold_json.trim() || undefined,
      owner: ruleForm.owner.trim() || undefined,
    })
    ElMessage.success('规则创建成功')
    ruleDialogVisible.value = false
    fetchRules()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '创建规则失败')
  } finally {
    ruleSaving.value = false
  }
}

async function onDeleteRule(row: QualityRule) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除规则「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await deleteQualityRule(row.id)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ==================== 页签二：质量任务 ====================
const jobLoading = ref(false)
const jobList = ref<QualityJob[]>([])
const jobTotal = ref(0)
const jobQuery = reactive({ page: 1, size: 10 })

const dsOptions = ref<DataSourceInfo[]>([])
/** 规则多选候选（拉前 100 条） */
const ruleOptions = ref<QualityRule[]>([])

async function fetchJobs() {
  jobLoading.value = true
  try {
    const resp = await qualityJobList({ page: jobQuery.page, size: jobQuery.size })
    jobList.value = resp.list || []
    jobTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取任务列表失败')
  } finally {
    jobLoading.value = false
  }
}

function onJobSearch() {
  jobQuery.page = 1
  fetchJobs()
}

/** 数据源下拉懒加载（健康分页签也复用） */
async function ensureDsOptions() {
  if (dsOptions.value.length) return
  try {
    const resp = await dataSourceList({ page: 1, size: 100 })
    dsOptions.value = resp.list || []
  } catch {
    // 下拉仅辅助录入，失败不打断页面
  }
}

async function fetchJobFormOptions() {
  try {
    const [dsResp, ruleResp] = await Promise.all([
      dataSourceList({ page: 1, size: 100 }),
      qualityRuleList({ page: 1, size: 100 }),
    ])
    dsOptions.value = dsResp.list || []
    ruleOptions.value = ruleResp.list || []
  } catch {
    ElMessage.warning('数据源/规则候选加载失败，可手动刷新重试')
  }
}

// ---------- 新建任务对话框 ----------
const jobDialogVisible = ref(false)
const jobSaving = ref(false)
const jobFormEl = ref<FormInstance>()
const jobForm = reactive({
  name: '',
  datasource_id: undefined as number | undefined,
  db_name: '',
  rule_ids: [] as number[],
  cron: '',
})

const jobRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  rule_ids: [
    {
      type: 'array',
      required: true,
      message: '请至少选择一条质量规则',
      trigger: 'change',
    },
  ],
})

function openJobCreate() {
  Object.assign(jobForm, {
    name: '',
    datasource_id: undefined,
    db_name: '',
    rule_ids: [],
    cron: '',
  })
  fetchJobFormOptions()
  jobDialogVisible.value = true
}

async function onSaveJob() {
  const valid = await jobFormEl.value?.validate().catch(() => false)
  if (!valid) return
  jobSaving.value = true
  try {
    await createQualityJob({
      name: jobForm.name.trim(),
      datasource_id: Number(jobForm.datasource_id),
      db_name: jobForm.db_name.trim() || undefined,
      rule_ids: jobForm.rule_ids.join(','),
      cron: jobForm.cron.trim() || undefined,
    })
    ElMessage.success('任务创建成功')
    jobDialogVisible.value = false
    fetchJobs()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '创建任务失败')
  } finally {
    jobSaving.value = false
  }
}

// ---------- 执行任务 ----------
const runningId = ref<number | null>(null)
const runResultVisible = ref(false)
const runResult = ref<RunJobResp | null>(null)

async function onRunJob(row: QualityJob) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定执行任务「${row.name}」吗？将串行执行其绑定的 ${row.rule_ids || '全部'} 规则。`,
    '执行确认',
    { confirmButtonText: '执行', cancelButtonText: '取消', type: 'info' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  runningId.value = row.id
  try {
    const resp = await runQualityJob(row.id)
    runResult.value = resp
    runResultVisible.value = true
    fetchJobs()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '执行失败')
  } finally {
    runningId.value = null
  }
}

// ---------- 任务执行日志（结果明细） ----------
const logVisible = ref(false)
const logJob = ref<QualityJob | null>(null)
const logLoading = ref(false)
const logList = ref<QualityResult[]>([])
const logTotal = ref(0)
const logQuery = reactive({ page: 1, size: 10 })

function openLog(row: QualityJob) {
  logJob.value = row
  logQuery.page = 1
  logVisible.value = true
  fetchJobResults()
}

async function fetchJobResults() {
  if (!logJob.value?.id) return
  logLoading.value = true
  try {
    const resp = await qualityResultList({
      job_id: logJob.value.id,
      page: logQuery.page,
      size: logQuery.size,
    })
    logList.value = resp.list || []
    logTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取执行结果失败')
  } finally {
    logLoading.value = false
  }
}

// ==================== 页签三：校验结果 ====================
const resultLoading = ref(false)
const resultList = ref<QualityResult[]>([])
const resultTotal = ref(0)
const resultQuery = reactive({
  job_id: undefined as number | undefined,
  pass_filter: undefined as number | undefined,
  page: 1,
  size: 10,
})

const passOptions = [
  { value: 1, label: '通过' },
  { value: 0, label: '不通过' },
]

/** 结果 / 错误数据页共用的 job 下拉候选 */
const jobOptions = ref<QualityJob[]>([])
async function fetchJobOptions() {
  if (jobOptions.value.length) return
  try {
    const resp = await qualityJobList({ page: 1, size: 100 })
    jobOptions.value = resp.list || []
  } catch {
    jobOptions.value = []
  }
}

async function fetchResults() {
  resultLoading.value = true
  try {
    const resp = await qualityResultList({
      job_id: resultQuery.job_id,
      pass_filter: resultQuery.pass_filter,
      page: resultQuery.page,
      size: resultQuery.size,
    })
    resultList.value = resp.list || []
    resultTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取校验结果失败')
  } finally {
    resultLoading.value = false
  }
}

function onResultSearch() {
  resultQuery.page = 1
  fetchResults()
}

// ==================== 页签四：告警配置 ====================
const channelTypeOptions = [
  { value: 'dingtalk', label: '钉钉机器人' },
  { value: 'feishu', label: '飞书机器人' },
]
const minLevelOptions = [
  { value: 'failed', label: 'failed（仅失败）' },
  { value: 'partial', label: 'partial（含部分通过）' },
]
const channelTypeText: Record<string, string> = Object.fromEntries(
  channelTypeOptions.map((o) => [o.value, o.label]),
)

const channelLoading = ref(false)
const channelOffline = ref(false)
const channelList = ref<AlertChannel[]>([])

async function fetchChannels() {
  channelLoading.value = true
  try {
    const resp = await listAlertChannels({ page: 1, size: 200 })
    channelList.value = resp.list || []
    channelOffline.value = false
  } catch (e) {
    if (isOffline(e)) {
      channelOffline.value = true
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '获取告警通道失败')
    }
  } finally {
    channelLoading.value = false
  }
}

/** 启停开关：复用 save 接口整体保存（status 字段后端契约确认中，未支持时后端会忽略） */
const statusTogglingId = ref<number | null>(null)
async function onChannelStatusChange(row: AlertChannel, val: string | number | boolean) {
  if (row.id === undefined) return
  const next = Number(val) ? 1 : 0
  const prev = row.status ?? 1
  if (next === prev) return
  statusTogglingId.value = row.id
  try {
    await saveAlertChannel({
      id: row.id,
      name: row.name || '',
      type: row.type || 'dingtalk',
      webhook_url: row.webhook_url || '',
      secret: row.secret || undefined,
      min_level: row.min_level || 'failed',
      status: next,
    })
    row.status = next
    ElMessage.success(next === 1 ? '通道已启用' : '通道已停用')
  } catch (e) {
    row.status = prev
    ElMessage.error(e instanceof ApiError ? e.message : '状态更新失败')
  } finally {
    statusTogglingId.value = null
  }
}

// ---------- 新建 / 编辑通道对话框 ----------
const channelDialogVisible = ref(false)
const channelSaving = ref(false)
const channelEditingId = ref<number | null>(null)
const channelFormEl = ref<FormInstance>()
const channelForm = reactive({
  name: '',
  type: 'dingtalk',
  webhook_url: '',
  secret: '',
  min_level: 'failed',
})

const channelRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入通道名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择通道类型', trigger: 'change' }],
  webhook_url: [
    { required: true, message: '请输入 Webhook 地址', trigger: 'blur' },
    { pattern: /^https?:\/\//, message: '需以 http(s):// 开头', trigger: 'blur' },
  ],
  min_level: [{ required: true, message: '请选择最低告警级别', trigger: 'change' }],
})

function openChannelCreate() {
  channelEditingId.value = null
  Object.assign(channelForm, { name: '', type: 'dingtalk', webhook_url: '', secret: '', min_level: 'failed' })
  channelDialogVisible.value = true
}

function openChannelEdit(row: AlertChannel) {
  channelEditingId.value = row.id ?? null
  Object.assign(channelForm, {
    name: row.name || '',
    type: row.type || 'dingtalk',
    webhook_url: row.webhook_url || '',
    secret: row.secret || '',
    min_level: row.min_level || 'failed',
  })
  channelDialogVisible.value = true
}

async function onSaveChannel() {
  const valid = await channelFormEl.value?.validate().catch(() => false)
  if (!valid) return
  channelSaving.value = true
  try {
    await saveAlertChannel({
      id: channelEditingId.value ?? 0,
      name: channelForm.name.trim(),
      type: channelForm.type,
      webhook_url: channelForm.webhook_url.trim(),
      secret: channelForm.secret.trim() || undefined,
      min_level: channelForm.min_level,
    })
    ElMessage.success(channelEditingId.value === null ? '通道创建成功' : '通道更新成功')
    channelDialogVisible.value = false
    fetchChannels()
  } catch (e) {
    if (isOffline(e)) {
      ElMessage.warning('告警配置接口上线中，后端路由合入后即可保存')
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '保存通道失败')
    }
  } finally {
    channelSaving.value = false
  }
}

async function onDeleteChannel(row: AlertChannel) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除告警通道「${row.name}」吗？删除后该通道不再接收告警。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await deleteAlertChannel(row.id)
    ElMessage.success('删除成功')
    fetchChannels()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ---------- 发送测试 ----------
const testingId = ref<number | null>(null)
const testResultVisible = ref(false)
const testResult = ref<TestAlertResp | null>(null)

async function onTestChannel(row: AlertChannel) {
  if (row.id === undefined) return
  testingId.value = row.id
  try {
    const resp = await testAlertChannel(row.id)
    testResult.value = resp
    testResultVisible.value = true
  } catch (e) {
    if (isOffline(e)) {
      ElMessage.warning('测试接口上线中，后端路由合入后可用')
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '发送测试失败')
    }
  } finally {
    testingId.value = null
  }
}

// ==================== 页签五：错误数据 ====================
const edLoading = ref(false)
const edOffline = ref(false)
const edList = ref<ErrorData[]>([])
const edTotal = ref(0)
const edQuery = reactive({
  job_id: undefined as number | undefined,
  rule_id: '',
  page: 1,
  size: 10,
})

/** 规则 ID 手工输入 → 合法正整数才下发 */
const edRuleId = computed(() => {
  const n = Number(edQuery.rule_id)
  return edQuery.rule_id && Number.isInteger(n) && n > 0 ? n : undefined
})

async function fetchErrorData() {
  edLoading.value = true
  try {
    const resp = await listErrorData({
      job_id: edQuery.job_id,
      rule_id: edRuleId.value,
      page: edQuery.page,
      size: edQuery.size,
    })
    edList.value = resp.list || []
    edTotal.value = resp.total || 0
    edOffline.value = false
  } catch (e) {
    if (isOffline(e)) {
      edOffline.value = true
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '获取错误数据失败')
    }
  } finally {
    edLoading.value = false
  }
}

function onEdSearch() {
  edQuery.page = 1
  fetchErrorData()
}

// ---------- 错误样本弹窗（动态列渲染 rows_json） ----------
const sampleVisible = ref(false)
const sampleRecord = ref<ErrorData | null>(null)
const sampleColumns = ref<string[]>([])
const sampleRows = ref<Record<string, unknown>[]>([])
const sampleRaw = ref('')

function openSample(row: ErrorData) {
  sampleRecord.value = row
  const parsed = parseJsonRows(row.rows_json)
  sampleColumns.value = parsed.columns
  sampleRows.value = parsed.rows
  sampleRaw.value = row.rows_json || ''
  sampleVisible.value = true
}

// ==================== 页签六：健康分 ====================
const healthFormEl = ref<FormInstance>()
const healthForm = reactive({
  datasource_id: undefined as number | undefined,
  db_name: '',
  table_name: '',
  pk_column: '',
})
const healthRules = reactive<FormRules>({
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  table_name: [{ required: true, message: '请输入表名', trigger: 'blur' }],
})

const healthLoading = ref(false)
const healthOffline = ref(false)
const profile = ref<TableProfile | null>(null)
const profileColumns = ref<string[]>([])
const profileRows = ref<Record<string, unknown>[]>([])

/** /quality/profile/table 可能回 {profile} 包装或直接 TableProfile，做兼容 */
function normalizeProfile(resp: ProfileReply | TableProfile | null | undefined): TableProfile | null {
  if (!resp || typeof resp !== 'object') return null
  const p = (resp as ProfileReply).profile ?? (resp as TableProfile)
  return p && typeof p === 'object' ? p : null
}

/** 空值率：兼容 0~1 与 0~100 两种量纲 */
const nullRateText = computed(() => {
  const r = profile.value?.null_rate
  if (r === null || r === undefined) return '-'
  const pct = r <= 1 ? r * 100 : r
  return `${pct.toFixed(1)}%`
})

// ---------- 健康分仪表盘 ----------
const gaugeEl = ref<HTMLDivElement>()
let gaugeChart: echarts.ECharts | null = null

function renderGauge(score: number) {
  nextTick(() => {
    const el = gaugeEl.value
    if (!el) return
    gaugeChart?.dispose()
    gaugeChart = echarts.init(el)
    const color = scoreColor(score)
    gaugeChart.setOption({
      series: [
        {
          type: 'gauge',
          min: 0,
          max: 100,
          startAngle: 210,
          endAngle: -30,
          radius: '95%',
          axisLine: {
            lineStyle: { width: 18, color: [[0.6, '#f56c6c'], [0.8, '#e6a23c'], [1, '#67c23a']] },
          },
          pointer: { length: '60%', width: 5, itemStyle: { color: 'auto' } },
          axisTick: { distance: -18, length: 6, lineStyle: { color: '#fff', width: 1 } },
          splitLine: { distance: -18, length: 18, lineStyle: { color: '#fff', width: 2 } },
          axisLabel: { distance: 26, color: '#909399', fontSize: 11 },
          title: { offsetCenter: [0, '64%'], fontSize: 13, color: '#909399' },
          detail: {
            valueAnimation: true,
            offsetCenter: [0, '28%'],
            fontSize: 32,
            fontWeight: 600,
            color,
            formatter: '{value}',
          },
          data: [{ value: score, name: '健康分' }],
        },
      ],
    })
  })
}

// ---------- 历史趋势 ----------
const histEl = ref<HTMLDivElement>()
let histChart: echarts.ECharts | null = null
const histLoading = ref(false)
const histOffline = ref(false)
const historyList = ref<TableProfile[]>([])

function renderHistory(list: TableProfile[]) {
  nextTick(() => {
    const el = histEl.value
    if (!el) return
    histChart?.dispose()
    histChart = echarts.init(el)
    histChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 16, bottom: 26, top: 30 },
      xAxis: { type: 'category', boundaryGap: false, data: list.map((p) => fmtDate(p.ran_at)) },
      yAxis: { type: 'value', min: 0, max: 100 },
      series: [
        {
          name: '健康分',
          type: 'line',
          smooth: true,
          symbolSize: 6,
          data: list.map((p) => p.health_score ?? 0),
          areaStyle: { opacity: 0.12 },
          itemStyle: { color: '#409eff' },
          markLine: {
            silent: true,
            symbol: 'none',
            lineStyle: { color: '#f56c6c', type: 'dashed' },
            data: [{ yAxis: 60, label: { formatter: '及格线 60', position: 'insideEndTop' } }],
          },
        },
      ],
    })
  })
}

async function fetchHistory() {
  if (!healthForm.datasource_id) return
  histLoading.value = true
  try {
    const resp = await listProfiles({
      datasource_id: Number(healthForm.datasource_id),
      table_name: healthForm.table_name.trim() || undefined,
      page: 1,
      size: 50,
    })
    const sorted = [...(resp.list || [])].sort((a, b) => (a.ran_at || 0) - (b.ran_at || 0))
    historyList.value = sorted.slice(-10)
    histOffline.value = false
    if (historyList.value.length) renderHistory(historyList.value)
  } catch (e) {
    if (isOffline(e)) histOffline.value = true
    // 历史趋势加载失败不打断主分析流程
  } finally {
    histLoading.value = false
  }
}

// ---------- 开始分析 ----------
async function onAnalyze() {
  const valid = await healthFormEl.value?.validate().catch(() => false)
  if (!valid) return
  healthLoading.value = true
  try {
    const resp = await profileTable({
      datasource_id: Number(healthForm.datasource_id),
      db_name: healthForm.db_name.trim() || undefined,
      table_name: healthForm.table_name.trim(),
      pk_column: healthForm.pk_column.trim() || undefined,
    })
    const p = normalizeProfile(resp)
    profile.value = p
    healthOffline.value = false
    if (p) {
      renderGauge(p.health_score ?? 0)
      const parsed = parseJsonRows(p.profile_json)
      profileColumns.value = parsed.columns
      profileRows.value = parsed.rows
    }
    fetchHistory()
  } catch (e) {
    if (isOffline(e)) {
      healthOffline.value = true
      ElMessage.warning('健康分析接口上线中，后端路由合入后即可使用')
    } else {
      ElMessage.error(e instanceof ApiError ? e.message : '表健康分析失败')
    }
  } finally {
    healthLoading.value = false
  }
}

// ---------- 页签切换：懒加载对应数据 ----------
const activeTab = ref('rule')
function onTabChange() {
  switch (activeTab.value) {
    case 'rule':
      if (!ruleList.value.length && !ruleLoading.value) fetchRules()
      break
    case 'job':
      if (!jobList.value.length && !jobLoading.value) fetchJobs()
      break
    case 'result':
      if (!resultList.value.length && !resultLoading.value) {
        fetchJobOptions()
        fetchResults()
      }
      break
    case 'channel':
      if (!channelList.value.length && !channelLoading.value && !channelOffline.value) fetchChannels()
      break
    case 'errordata':
      if (!edList.value.length && !edLoading.value && !edOffline.value) {
        fetchJobOptions()
        fetchErrorData()
      }
      break
    case 'health':
      ensureDsOptions()
      break
  }
}

function onResize() {
  gaugeChart?.resize()
  histChart?.resize()
}

onMounted(() => {
  fetchRules()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  gaugeChart?.dispose()
  gaugeChart = null
  histChart?.dispose()
  histChart = null
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ==================== 规则 ==================== -->
        <el-tab-pane label="质量规则" name="rule">
          <div class="toolbar">
            <el-select
              v-model="ruleQuery.rule_type"
              placeholder="全部类型"
              clearable
              style="width: 150px"
              @change="onRuleSearch"
            >
              <el-option v-for="t in ruleTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
            <el-button type="primary" @click="openRuleCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建规则
            </el-button>
            <el-button @click="fetchRules">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="ruleLoading" :data="ruleList" stripe border>
            <el-table-column prop="id" label="ID" width="64" />
            <el-table-column prop="name" label="规则名称" min-width="140" show-overflow-tooltip />
            <el-table-column label="类型" width="120">
              <template #default="{ row }">
                <el-tag size="small">{{ ruleTypeLabel(row.rule_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="target_table" label="目标表" min-width="140" show-overflow-tooltip />
            <el-table-column label="目标列" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.target_column || '-' }}</template>
            </el-table-column>
            <el-table-column prop="params_json" label="参数" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">{{ row.params_json || '-' }}</template>
            </el-table-column>
            <el-table-column prop="threshold_json" label="阈值" min-width="130" show-overflow-tooltip>
              <template #default="{ row }">{{ row.threshold_json || '-' }}</template>
            </el-table-column>
            <el-table-column prop="owner" label="负责人" width="100" show-overflow-tooltip />
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="ruleQuery.page"
            v-model:page-size="ruleQuery.size"
            :total="ruleTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchRules"
            @size-change="onRuleSearch"
          />
        </el-tab-pane>

        <!-- ==================== 任务 ==================== -->
        <el-tab-pane label="质量任务" name="job">
          <div class="toolbar">
            <el-button type="primary" @click="openJobCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建任务
            </el-button>
            <el-button @click="fetchJobs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="jobLoading" :data="jobList" stripe border>
            <el-table-column prop="id" label="ID" width="64" />
            <el-table-column prop="name" label="任务名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="datasource_id" label="数据源" width="80" align="center" />
            <el-table-column label="目标库" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.db_name || '默认库' }}</template>
            </el-table-column>
            <el-table-column prop="rule_ids" label="规则 ID" min-width="130" show-overflow-tooltip />
            <el-table-column label="调度周期" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.cron || '-' }}</template>
            </el-table-column>
            <el-table-column label="最近状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag
                  v-if="row.last_status"
                  size="small"
                  :type="row.last_status === 'passed' ? 'success' : row.last_status === 'failed' ? 'danger' : 'info'"
                >
                  {{ row.last_status }}
                </el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="最近执行" width="170">
              <template #default="{ row }">{{ fmtTime(row.last_run_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button
                  link
                  type="success"
                  :loading="runningId === row.id"
                  @click="onRunJob(row)"
                >
                  执行
                </el-button>
                <el-button link type="primary" @click="openLog(row)">日志</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="jobQuery.page"
            v-model:page-size="jobQuery.size"
            :total="jobTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchJobs"
            @size-change="onJobSearch"
          />
        </el-tab-pane>

        <!-- ==================== 结果 ==================== -->
        <el-tab-pane label="校验结果" name="result">
          <div class="toolbar">
            <el-select
              v-model="resultQuery.job_id"
              placeholder="全部任务"
              clearable
              filterable
              style="width: 200px"
              @change="onResultSearch"
            >
              <el-option
                v-for="j in jobOptions"
                :key="j.id"
                :label="`#${j.id} ${j.name}`"
                :value="j.id!"
              />
            </el-select>
            <el-select
              v-model="resultQuery.pass_filter"
              placeholder="全部结果"
              clearable
              style="width: 130px"
              @change="onResultSearch"
            >
              <el-option v-for="p in passOptions" :key="p.value" :label="p.label" :value="p.value" />
            </el-select>
            <el-button type="primary" @click="onResultSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button @click="fetchResults">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="resultLoading" :data="resultList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="job_id" label="任务 ID" width="80" align="center" />
            <el-table-column prop="rule_id" label="规则 ID" width="80" align="center" />
            <el-table-column label="规则类型" width="120">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ ruleTypeLabel(row.rule_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="check_value" label="检查值" width="110" align="right" />
            <el-table-column label="结果" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="row.pass_flag ? 'success' : 'danger'" size="small">
                  {{ row.pass_flag ? '通过' : '不通过' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="detail_json" label="明细" min-width="220" show-overflow-tooltip />
            <el-table-column label="执行时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.run_at) }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="resultQuery.page"
            v-model:page-size="resultQuery.size"
            :total="resultTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchResults"
            @size-change="onResultSearch"
          />
        </el-tab-pane>

        <!-- ==================== 告警配置 ==================== -->
        <el-tab-pane label="告警配置" name="channel">
          <el-alert
            v-if="channelOffline"
            title="告警通道接口暂未就绪（功能上线中），后端路由合入后即可配置"
            type="info"
            show-icon
            :closable="false"
            class="offline-alert"
          />
          <div v-else class="toolbar">
            <el-button type="primary" @click="openChannelCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建通道
            </el-button>
            <el-button @click="fetchChannels">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-empty v-if="channelOffline" description="告警配置 · 功能上线中" :image-size="90" />
          <template v-else>
            <el-table v-loading="channelLoading" :data="channelList" stripe border>
              <el-table-column prop="id" label="ID" width="64" />
              <el-table-column prop="name" label="通道名称" min-width="130" show-overflow-tooltip />
              <el-table-column label="类型" width="120">
                <template #default="{ row }">
                  <el-tag size="small" :type="row.type === 'feishu' ? 'warning' : 'primary'">
                    {{ channelTypeText[row.type] || row.type || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="webhook_url" label="Webhook" min-width="220" show-overflow-tooltip />
              <el-table-column label="最低告警级别" width="140">
                <template #default="{ row }">
                  <el-tag size="small" :type="row.min_level === 'partial' ? 'warning' : 'danger'" effect="plain">
                    {{ row.min_level || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="启用状态" width="100" align="center">
                <template #default="{ row }">
                  <el-switch
                    :model-value="row.status ?? 1"
                    :active-value="1"
                    :inactive-value="0"
                    :loading="statusTogglingId === row.id"
                    @change="onChannelStatusChange(row, $event)"
                  />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="170" fixed="right">
                <template #default="{ row }">
                  <el-button
                    link
                    type="success"
                    :loading="testingId === row.id"
                    @click="onTestChannel(row)"
                  >
                    发送测试
                  </el-button>
                  <el-button link type="primary" @click="openChannelEdit(row)">编辑</el-button>
                  <el-button link type="danger" @click="onDeleteChannel(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty
              v-if="!channelLoading && !channelList.length"
              description="暂无告警通道，点击「新建通道」接入钉钉 / 飞书机器人"
              :image-size="90"
            />
          </template>
        </el-tab-pane>

        <!-- ==================== 错误数据 ==================== -->
        <el-tab-pane label="错误数据" name="errordata">
          <el-alert
            v-if="edOffline"
            title="错误数据接口暂未就绪（功能上线中），后端路由合入后即可查询"
            type="info"
            show-icon
            :closable="false"
            class="offline-alert"
          />
          <div v-else class="toolbar">
            <el-select
              v-model="edQuery.job_id"
              placeholder="全部任务"
              clearable
              filterable
              style="width: 200px"
              @change="onEdSearch"
            >
              <el-option
                v-for="j in jobOptions"
                :key="j.id"
                :label="`#${j.id} ${j.name}`"
                :value="j.id!"
              />
            </el-select>
            <el-input
              v-model="edQuery.rule_id"
              placeholder="规则 ID"
              clearable
              style="width: 140px"
              @keyup.enter="onEdSearch"
              @clear="onEdSearch"
            />
            <el-button type="primary" @click="onEdSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button @click="fetchErrorData">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-empty v-if="edOffline" description="错误数据 · 功能上线中" :image-size="90" />
          <template v-else>
            <el-table v-loading="edLoading" :data="edList" stripe border>
              <el-table-column prop="job_id" label="任务 ID" width="90" align="center" />
              <el-table-column prop="rule_id" label="规则 ID" width="90" align="center" />
              <el-table-column prop="sample_count" label="样本条数" width="100" align="center" />
              <el-table-column prop="rows_json" label="样本预览" min-width="260" show-overflow-tooltip>
                <template #default="{ row }">{{ row.rows_json || '-' }}</template>
              </el-table-column>
              <el-table-column label="采集时间" width="170">
                <template #default="{ row }">{{ fmtTime(row.run_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="100" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="openSample(row)">查看样本</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty
              v-if="!edLoading && !edList.length"
              description="暂无错误数据样本：仅校验失败且采样开启的规则会产生样本"
              :image-size="90"
            />

            <el-pagination
              v-model:current-page="edQuery.page"
              v-model:page-size="edQuery.size"
              :total="edTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next, jumper"
              class="pager"
              @current-change="fetchErrorData"
              @size-change="onEdSearch"
            />
          </template>
        </el-tab-pane>

        <!-- ==================== 健康分 ==================== -->
        <el-tab-pane label="健康分" name="health">
          <el-alert
            v-if="healthOffline"
            title="健康分析接口暂未就绪（功能上线中），后端路由合入后即可使用"
            type="info"
            show-icon
            :closable="false"
            class="offline-alert"
          />
          <el-form
            ref="healthFormEl"
            :model="healthForm"
            :rules="healthRules"
            inline
            :disabled="healthOffline"
            class="health-form"
          >
            <el-form-item label="数据源" prop="datasource_id">
              <el-select
                v-model="healthForm.datasource_id"
                filterable
                placeholder="选择数据源"
                style="width: 210px"
              >
                <el-option
                  v-for="d in dsOptions"
                  :key="d.id"
                  :label="`${d.name}（ID:${d.id} · ${d.type}）`"
                  :value="d.id!"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="目标库">
              <el-input v-model="healthForm.db_name" placeholder="可选" style="width: 130px" />
            </el-form-item>
            <el-form-item label="表名" prop="table_name">
              <el-input v-model="healthForm.table_name" placeholder="如 orders" style="width: 160px" />
            </el-form-item>
            <el-form-item label="主键列">
              <el-input v-model="healthForm.pk_column" placeholder="如 id（用于重复检测）" style="width: 180px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="healthLoading" @click="onAnalyze">
                <el-icon><DataAnalysis /></el-icon>&nbsp;开始分析
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="profile" v-loading="healthLoading" class="health-result">
            <el-row :gutter="16">
              <el-col :xs="24" :md="8">
                <el-card shadow="never">
                  <div ref="gaugeEl" class="gauge-chart" />
                </el-card>
              </el-col>
              <el-col :xs="24" :md="16">
                <el-card shadow="never">
                  <el-row :gutter="12">
                    <el-col :span="6">
                      <div class="mini-stat">
                        <div class="mini-value">{{ profile.rows_cnt ?? '-' }}</div>
                        <div class="mini-label">总行数</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="mini-stat">
                        <div class="mini-value">{{ profile.cols_cnt ?? '-' }}</div>
                        <div class="mini-label">列数</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="mini-stat">
                        <div class="mini-value" :class="{ 'mini-warn': profile.null_rate && profile.null_rate > 0.1 }">
                          {{ nullRateText }}
                        </div>
                        <div class="mini-label">空值率</div>
                      </div>
                    </el-col>
                    <el-col :span="6">
                      <div class="mini-stat">
                        <div class="mini-value" :class="{ 'mini-bad': (profile.dup_pk_cnt ?? 0) > 0 }">
                          {{ profile.dup_pk_cnt ?? '-' }}
                        </div>
                        <div class="mini-label">主键重复</div>
                      </div>
                    </el-col>
                  </el-row>

                  <div class="hist-title">
                    <span>
                      近 10 次健康分趋势
                      <el-tag :type="scoreTagType(profile.health_score)" size="small" class="score-tag">
                        当前 {{ profile.health_score ?? '-' }}
                      </el-tag>
                    </span>
                    <el-button link type="primary" :loading="histLoading" @click="fetchHistory">
                      刷新
                    </el-button>
                  </div>
                  <div v-show="historyList.length" ref="histEl" class="hist-chart" />
                  <el-empty
                    v-if="!historyList.length && !histLoading"
                    :description="histOffline ? '历史趋势接口上线中' : '暂无历史分析记录'"
                    :image-size="60"
                  />
                </el-card>
              </el-col>
            </el-row>

            <el-card v-if="profileColumns.length" shadow="never" class="mt16">
              <template #header><span class="card-title">逐列画像（profile_json）</span></template>
              <el-table :data="profileRows" size="small" border max-height="360">
                <el-table-column
                  v-for="col in profileColumns"
                  :key="col"
                  :prop="col"
                  :label="col"
                  min-width="110"
                  show-overflow-tooltip
                >
                  <template #default="{ row }">{{ formatCell(row[col]) }}</template>
                </el-table-column>
              </el-table>
            </el-card>
          </div>
          <el-empty
            v-else-if="!healthOffline"
            description="选择表后点击「开始分析」，查看健康分仪表盘 / 空值率 / 主键重复与逐列画像"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新建规则对话框 -->
    <el-dialog v-model="ruleDialogVisible" title="新建质量规则" width="620px" destroy-on-close>
      <el-form ref="ruleFormEl" :model="ruleForm" :rules="ruleRules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="如：订单表主键唯一性" maxlength="64" />
        </el-form-item>
        <el-form-item label="规则类型" prop="rule_type">
          <el-select v-model="ruleForm.rule_type" style="width: 100%">
            <el-option v-for="t in ruleTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标表" prop="target_table">
          <el-input v-model="ruleForm.target_table" placeholder="如 orders（仅字母/数字/下划线）" />
        </el-form-item>
        <el-form-item label="目标列" prop="target_column">
          <el-input
            v-model="ruleForm.target_column"
            :placeholder="columnRequiredTypes.includes(ruleForm.rule_type) ? '必填，如 order_id' : '该类型可留空'"
          />
        </el-form-item>
        <el-form-item label="参数 JSON" prop="params_json">
          <el-input
            v-model="ruleForm.params_json"
            type="textarea"
            :rows="3"
            :placeholder="paramsPlaceholder"
          />
        </el-form-item>
        <el-form-item label="阈值 JSON" prop="threshold_json">
          <el-input
            v-model="ruleForm.threshold_json"
            type="textarea"
            :rows="2"
            placeholder='可选，行数阈值，如：{"op":"gte","value":100} 或 {"op":"between","value":[100,2000]}'
          />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="ruleForm.owner" placeholder="如：data-team" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSaving" @click="onSaveRule">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新建任务对话框 -->
    <el-dialog v-model="jobDialogVisible" title="新建质量任务" width="600px" destroy-on-close>
      <el-form ref="jobFormEl" :model="jobForm" :rules="jobRules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="jobForm.name" placeholder="如：订单库日质量巡检" maxlength="64" />
        </el-form-item>
        <el-form-item label="数据源" prop="datasource_id">
          <el-select v-model="jobForm.datasource_id" filterable placeholder="选择数据源" style="width: 100%">
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（ID:${d.id} · ${d.type}）`"
              :value="d.id!"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标库">
          <el-input v-model="jobForm.db_name" placeholder="可选，留空使用数据源默认库" />
        </el-form-item>
        <el-form-item label="质量规则" prop="rule_ids">
          <el-select
            v-model="jobForm.rule_ids"
            multiple
            filterable
            placeholder="选择要执行的规则（可多选）"
            style="width: 100%"
          >
            <el-option
              v-for="r in ruleOptions"
              :key="r.id"
              :label="`#${r.id} ${r.name}（${ruleTypeLabel(r.rule_type)}）`"
              :value="r.id!"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="调度 Cron">
          <el-input v-model="jobForm.cron" placeholder="可选，如 0 3 * * *" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="jobDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="jobSaving" @click="onSaveJob">保存</el-button>
      </template>
    </el-dialog>

    <!-- 执行结果对话框 -->
    <el-dialog v-model="runResultVisible" title="任务执行完成" width="420px">
      <template v-if="runResult">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Job ID">{{ runResult.job_id }}</el-descriptions-item>
          <el-descriptions-item label="最终状态">
            <el-tag
              size="small"
              :type="runResult.last_status === 'passed' ? 'success' : runResult.last_status === 'failed' ? 'danger' : 'info'"
            >
              {{ runResult.last_status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="规则总数">{{ runResult.total_rules }}</el-descriptions-item>
          <el-descriptions-item label="通过数">
            {{ runResult.passed }}
            <span class="run-rate">（{{ runResult.total_rules ? Math.round((runResult.passed / runResult.total_rules) * 100) : 0 }}%）</span>
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="runResultVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 任务执行日志对话框 -->
    <el-dialog
      v-model="logVisible"
      :title="`执行日志 · ${logJob?.name || ''}（Job #${logJob?.id ?? '-'}）`"
      width="820px"
    >
      <el-table v-loading="logLoading" :data="logList" stripe border max-height="420">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="rule_id" label="规则 ID" width="80" align="center" />
        <el-table-column label="规则类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ ruleTypeLabel(row.rule_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="check_value" label="检查值" width="100" align="right" />
        <el-table-column label="结果" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.pass_flag ? 'success' : 'danger'" size="small">
              {{ row.pass_flag ? '通过' : '不通过' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="detail_json" label="明细" min-width="220" show-overflow-tooltip />
        <el-table-column label="执行时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.run_at) }}</template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="logQuery.page"
        v-model:page-size="logQuery.size"
        :total="logTotal"
        :page-sizes="[10, 20, 50]"
        layout="total, prev, pager, next"
        class="pager"
        @current-change="fetchJobResults"
      />
      <template #footer>
        <el-button @click="logVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 新建/编辑告警通道对话框 -->
    <el-dialog
      v-model="channelDialogVisible"
      :title="channelEditingId === null ? '新建告警通道' : '编辑告警通道'"
      width="580px"
      destroy-on-close
    >
      <el-form ref="channelFormEl" :model="channelForm" :rules="channelRules" label-width="110px">
        <el-form-item label="通道名称" prop="name">
          <el-input v-model="channelForm.name" placeholder="如：数据质量告警群" maxlength="64" />
        </el-form-item>
        <el-form-item label="通道类型" prop="type">
          <el-select v-model="channelForm.type" style="width: 100%">
            <el-option v-for="t in channelTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="Webhook" prop="webhook_url">
          <el-input v-model="channelForm.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
        </el-form-item>
        <el-form-item label="加签密钥">
          <el-input
            v-model="channelForm.secret"
            placeholder="可选，钉钉机器人「加签」安全设置下的 SEC 开头密钥"
            show-password
          />
        </el-form-item>
        <el-form-item label="最低告警级别" prop="min_level">
          <el-select v-model="channelForm.min_level" style="width: 100%">
            <el-option v-for="m in minLevelOptions" :key="m.value" :label="m.label" :value="m.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-alert
        title="min_level=failed 仅在任务最终失败时告警；partial 表示存在部分通过的规则也会告警"
        type="info"
        :closable="false"
        show-icon
        class="form-alert"
      />
      <template #footer>
        <el-button @click="channelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="channelSaving" @click="onSaveChannel">保存</el-button>
      </template>
    </el-dialog>

    <!-- 通道测试结果对话框 -->
    <el-dialog v-model="testResultVisible" title="通道测试结果" width="420px">
      <template v-if="testResult">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="是否成功">
            <el-tag :type="testResult.ok ? 'success' : 'danger'" size="small">
              {{ testResult.ok ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="返回信息">{{ testResult.msg || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button type="primary" @click="testResultVisible = false">知道了</el-button>
      </template>
    </el-dialog>

    <!-- 错误样本对话框（动态列） -->
    <el-dialog
      v-model="sampleVisible"
      :title="`错误样本 · Job #${sampleRecord?.job_id ?? '-'} / Rule #${sampleRecord?.rule_id ?? '-'}`"
      width="860px"
    >
      <el-alert
        v-if="!sampleColumns.length && sampleRaw"
        title="样本 JSON 解析失败，以下为原始内容"
        type="warning"
        :closable="false"
        show-icon
        class="form-alert"
      />
      <el-table
        v-if="sampleColumns.length"
        :data="sampleRows"
        size="small"
        border
        max-height="420"
      >
        <el-table-column
          v-for="col in sampleColumns"
          :key="col"
          :prop="col"
          :label="col"
          min-width="110"
          show-overflow-tooltip
        >
          <template #default="{ row }">{{ formatCell(row[col]) }}</template>
        </el-table-column>
      </el-table>
      <pre v-else-if="sampleRaw" class="raw-pre">{{ sampleRaw }}</pre>
      <el-empty v-else description="该记录没有样本数据" :image-size="70" />
      <div v-if="sampleColumns.length" class="sample-meta">
        共 {{ sampleRows.length }} 条样本
        <template v-if="sampleRecord?.sample_count !== undefined">（采集 {{ sampleRecord.sample_count }} 条）</template>
      </div>
      <template #footer>
        <el-button @click="sampleVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
/* 健康分配色 token（与全局设计系统/echarts gauge 梯度一致）：
   <60 红 #f56c6c · 60-80 黄 #e6a23c · >=80 绿 #67c23a，
   与 Dashboard 健康分 el-tag（danger/warning/success）完全同源 */
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
.run-rate {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.offline-alert {
  margin-bottom: 14px;
}
.form-alert {
  margin-top: 4px;
}
.raw-pre {
  max-height: 420px;
  overflow: auto;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--dc-radius-sm, 6px);
  padding: 12px;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
.sample-meta {
  margin-top: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.health-form {
  margin-bottom: 14px;
}
.health-result {
  min-height: 120px;
}
.gauge-chart {
  width: 100%;
  height: 260px;
  border-radius: var(--dc-radius, 10px);
  background: radial-gradient(circle at 50% 58%, var(--el-color-primary-light-9) 0%, transparent 68%);
}
.hist-chart {
  width: 100%;
  height: 200px;
}
.hist-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 14px 0 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
/* 区块标题左侧品牌色条（与全局 .card-title 规则同款） */
.hist-title > span {
  display: inline-flex;
  align-items: center;
}
.hist-title > span::before {
  content: '';
  display: inline-block;
  width: 4px;
  height: 14px;
  border-radius: 2px;
  background: var(--dc-grad-bar, linear-gradient(180deg, #2563eb, #06b6d4));
  margin-right: 8px;
}
.score-tag {
  margin-left: 8px;
  font-weight: 600;
}
.mini-stat {
  text-align: center;
  padding: 12px 4px;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--dc-radius, 10px);
  transition: var(--dc-transition, all 0.22s);
}
.mini-stat:hover {
  transform: translateY(-2px);
  box-shadow: var(--dc-shadow-sm, 0 1px 3px rgba(15, 23, 42, 0.08));
  border-color: var(--el-color-primary-light-8);
}
.mini-value {
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.3;
  font-variant-numeric: tabular-nums;
}
.mini-warn {
  color: var(--dc-warning, #e6a23c);
}
.mini-bad {
  color: var(--dc-danger, #f56c6c);
}
.mini-label {
  margin-top: 2px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.mt16 {
  margin-top: 16px;
}
.card-title {
  font-weight: 600;
  font-size: 14px;
}
</style>
