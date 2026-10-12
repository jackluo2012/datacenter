<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用；菜单种子 theme/ThemeList → 路由 name ThemeTheme）
defineOptions({ name: 'ThemeTheme' })
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  saveThemeDomain,
  deleteThemeDomain,
  themeDomainList,
  saveTheme,
  deleteTheme,
  themeList,
  saveThemeModel,
  deleteThemeModel,
  themeModelList,
  deployThemeModel,
  saveThemeMv,
  deleteThemeMv,
  themeMvList,
  refreshThemeMv,
  type ThemeDomainInfo,
  type ThemeInfo,
  type ThemeModelInfo,
  type ThemeMvInfo,
} from '@/api/theme'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共 ----------
const activeTab = ref('domain')

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

/** 标识符白名单（对齐 theme rpc validate.go）：物理表名/视图名/编码会真实建表建视图，严格防注入 */
const IDENT_RE = /^[a-z][a-z0-9_]*$/
/** 字段名白名单：字母开头，仅字母/数字/下划线 */
const FIELD_NAME_RE = /^[A-Za-z][A-Za-z0-9_]*$/
/** MySQL 标识符 64 字符上限（<layer>_<table> 整体计算） */
const IDENT_MAX = 64
/** 字段类型白名单（对齐后端 ValidFieldType） */
const FIELD_TYPES = ['string', 'int', 'bigint', 'decimal', 'bool', 'date', 'datetime']

function onTabChange(name: string | number) {
  const tab = String(name)
  if (tab === 'domain') fetchDomains()
  else if (tab === 'theme') fetchThemes()
  else if (tab === 'model') fetchModels()
  else if (tab === 'mv') fetchMvs()
}

/** 数据源下拉（模型目标数据源 / 物化视图），进页面拉一次 */
const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 200 })
    dsOptions.value = resp.list || []
  } catch {
    /* 下拉数据拉取失败不阻断页面 */
  }
}

function dsNameOf(id?: number) {
  if (!id) return '-'
  return dsOptions.value.find((d) => d.id === id)?.name || String(id)
}

// ---------- 主题域 ----------
// 契约说明：网关 SaveThemeDomainReq 无 id 字段（rpc 恒按新建处理），故只提供新建/删除，无编辑
const dmLoading = ref(false)
const dmList = ref<ThemeDomainInfo[]>([])

async function fetchDomains() {
  dmLoading.value = true
  try {
    const resp = await themeDomainList()
    dmList.value = resp.list || []
  } catch (e) {
    ElMessage.error(errText(e, '获取主题域列表失败'))
  } finally {
    dmLoading.value = false
  }
}

const dmDialogVisible = ref(false)
const dmSaving = ref(false)
const dmFormRef = ref<FormInstance>()
const dmForm = reactive({ code: '', name: '', comment: '' })
const dmRules: FormRules = {
  code: [
    { required: true, message: '请输入域编码', trigger: 'blur' },
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
  name: [{ required: true, message: '请输入域名称', trigger: 'blur' }],
}

function openDmCreate() {
  Object.assign(dmForm, { code: '', name: '', comment: '' })
  dmDialogVisible.value = true
}

async function onSaveDm() {
  const valid = await dmFormRef.value?.validate().catch(() => false)
  if (!valid) return
  dmSaving.value = true
  try {
    await saveThemeDomain({
      code: dmForm.code.trim(),
      name: dmForm.name.trim(),
      comment: dmForm.comment.trim() || undefined,
    })
    ElMessage.success('主题域创建成功')
    dmDialogVisible.value = false
    fetchDomains()
  } catch (e) {
    ElMessage.error(errText(e, '保存主题域失败'))
  } finally {
    dmSaving.value = false
  }
}

async function onDeleteDm(row: ThemeDomainInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除主题域「${row.name || row.code}」吗？域下存在主题时后端将拒绝删除。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteThemeDomain(row.id)
    ElMessage.success('删除成功')
    fetchDomains()
  } catch (e) {
    ElMessage.error(errText(e, '删除主题域失败（域下存在主题时会被拒绝）'))
  }
}

// ---------- 主题 ----------
const thLoading = ref(false)
const thList = ref<ThemeInfo[]>([])
const thTotal = ref(0)
const thQuery = reactive({ domain_id: undefined as number | undefined, page: 1, size: 10 })

async function fetchThemes() {
  thLoading.value = true
  try {
    const resp = await themeList({
      domain_id: thQuery.domain_id || undefined,
      page: thQuery.page,
      size: thQuery.size,
    })
    thList.value = resp.list || []
    thTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取主题列表失败'))
  } finally {
    thLoading.value = false
  }
}

function thSearch() {
  thQuery.page = 1
  fetchThemes()
}

/** 主题下拉（主题域过滤条件之外的模型表单下拉），进页面拉一次，保存/删除主题后同步 */
const themeOptions = ref<ThemeInfo[]>([])
async function fetchThemeOptions() {
  try {
    const resp = await themeList({ page: 1, size: 200 })
    themeOptions.value = resp.list || []
  } catch {
    /* 忽略 */
  }
}

const themeNameMap = computed(() => {
  const m = new Map<number, string>()
  for (const t of themeOptions.value) if (t.id != null) m.set(t.id, t.name || t.code || '')
  return m
})

function themeNameOf(id?: number) {
  if (!id) return '-'
  return themeNameMap.value.get(id) || String(id)
}

const thDialogVisible = ref(false)
const thSaving = ref(false)
const thFormRef = ref<FormInstance>()
const thForm = reactive({
  domain_id: undefined as number | undefined,
  code: '',
  name: '',
  comment: '',
})
const thRules: FormRules = {
  domain_id: [{ required: true, message: '请选择所属主题域', trigger: 'change' }],
  code: [
    { required: true, message: '请输入主题编码', trigger: 'blur' },
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
  name: [{ required: true, message: '请输入主题名称', trigger: 'blur' }],
}

function openThCreate() {
  Object.assign(thForm, { domain_id: undefined, code: '', name: '', comment: '' })
  thDialogVisible.value = true
}

async function onSaveTh() {
  const valid = await thFormRef.value?.validate().catch(() => false)
  if (!valid) return
  thSaving.value = true
  try {
    await saveTheme({
      domain_id: thForm.domain_id as number,
      code: thForm.code.trim(),
      name: thForm.name.trim(),
      comment: thForm.comment.trim() || undefined,
    })
    ElMessage.success('主题创建成功')
    thDialogVisible.value = false
    fetchThemes()
    fetchThemeOptions()
  } catch (e) {
    ElMessage.error(errText(e, '保存主题失败'))
  } finally {
    thSaving.value = false
  }
}

async function onDeleteTh(row: ThemeInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除主题「${row.name || row.code}」吗？主题下存在模型时后端将拒绝删除。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteTheme(row.id)
    ElMessage.success('删除成功')
    fetchThemes()
    fetchThemeOptions()
  } catch (e) {
    ElMessage.error(errText(e, '删除主题失败（主题下存在模型时会被拒绝）'))
  }
}

// ---------- 分层模型 ----------
const mdlLoading = ref(false)
const mdlList = ref<ThemeModelInfo[]>([])
const mdlTotal = ref(0)
const mdlQuery = reactive({ theme_id: undefined as number | undefined, layer: '', page: 1, size: 10 })

const layerOptions = [
  { value: 'ods', label: 'ods 贴源层' },
  { value: 'dwd', label: 'dwd 明细层' },
  { value: 'dws', label: 'dws 汇总层' },
  { value: 'ads', label: 'ads 应用层' },
  { value: 'dim', label: 'dim 维度层' },
]

/** 五色分层 tag：ods紫 / dwd蓝 / dws青 / ads绿 / dim橙（色值见底部 scoped 样式） */
const layerTagClass: Record<string, string> = {
  ods: 'layer-ods',
  dwd: 'layer-dwd',
  dws: 'layer-dws',
  ads: 'layer-ads',
  dim: 'layer-dim',
}

/** 物理表名 = [db.]<layer>_<table_name>（部署时真实建表） */
function physicalTable(row: Pick<ThemeModelInfo, 'layer' | 'table_name' | 'db_name'>) {
  const t = `${row.layer || ''}_${row.table_name || ''}`
  const db = (row.db_name || '').trim()
  return db ? `${db}.${t}` : t
}

async function fetchModels() {
  mdlLoading.value = true
  try {
    const resp = await themeModelList({
      theme_id: mdlQuery.theme_id || undefined,
      layer: mdlQuery.layer || undefined,
      page: mdlQuery.page,
      size: mdlQuery.size,
    })
    mdlList.value = resp.list || []
    mdlTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取模型列表失败'))
  } finally {
    mdlLoading.value = false
  }
}

function mdlSearch() {
  mdlQuery.page = 1
  fetchModels()
}

interface FieldRow {
  name: string
  type: string
  comment: string
}

const mdlDialogVisible = ref(false)
const mdlSaving = ref(false)
const mdlFormRef = ref<FormInstance>()
const mdlForm = reactive({
  theme_id: undefined as number | undefined,
  layer: 'dwd',
  table_name: '',
  model_name: '',
  datasource_id: undefined as number | undefined,
  db_name: '',
  dim_json: '',
})
/** 字段定义动态行编辑；pkIndex 为勾选主键的行下标（恰一个 pk） */
const mdlFields = ref<FieldRow[]>([])
const pkIndex = ref(0)

/** 表名前缀实时提示：物理表名 = <layer>_<table_name> */
const physicalPreview = computed(() => {
  const t = mdlForm.table_name.trim()
  if (!t) return ''
  return physicalTable({ layer: mdlForm.layer, table_name: t, db_name: mdlForm.db_name })
})

const mdlRules: FormRules = {
  theme_id: [{ required: true, message: '请选择所属主题', trigger: 'change' }],
  layer: [{ required: true, message: '请选择数仓分层', trigger: 'change' }],
  table_name: [
    { required: true, message: '请输入表名（不含分层前缀）', trigger: 'blur' },
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
  model_name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
  datasource_id: [{ required: true, message: '请选择目标数据源', trigger: 'change' }],
  db_name: [
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
}

function addFieldRow() {
  mdlFields.value.push({ name: '', type: 'string', comment: '' })
}

function removeFieldRow(idx: number) {
  mdlFields.value.splice(idx, 1)
  if (idx === pkIndex.value) pkIndex.value = 0
  else if (idx < pkIndex.value) pkIndex.value -= 1
}

function openMdlCreate() {
  Object.assign(mdlForm, {
    theme_id: undefined,
    layer: 'dwd',
    table_name: '',
    model_name: '',
    datasource_id: undefined,
    db_name: '',
    dim_json: '',
  })
  // 预置一个 id 主键行，减少空表单误提交
  mdlFields.value = [{ name: 'id', type: 'bigint', comment: '主键' }]
  pkIndex.value = 0
  mdlDialogVisible.value = true
}

/** 字段定义校验并收敛为规范 JSON（对齐后端 ValidateFieldsJson：name 合法不重复 + type 白名单 + pk 恰有一个） */
function buildFieldsJson(): string | null {
  const rows = mdlFields.value
  if (rows.length === 0) {
    ElMessage.error('请至少添加一个字段')
    return null
  }
  const seen = new Set<string>()
  for (let i = 0; i < rows.length; i++) {
    const name = rows[i].name.trim()
    if (!name) {
      ElMessage.error(`第 ${i + 1} 个字段缺少字段名`)
      return null
    }
    if (!FIELD_NAME_RE.test(name)) {
      ElMessage.error(`字段名 ${name} 不合法（须以字母开头，仅含字母/数字/下划线）`)
      return null
    }
    if (seen.has(name)) {
      ElMessage.error(`字段 ${name} 重复定义`)
      return null
    }
    seen.add(name)
    if (!FIELD_TYPES.includes(rows[i].type)) {
      ElMessage.error(`字段 ${name} 类型 ${rows[i].type} 不在白名单（${FIELD_TYPES.join('/')}）`)
      return null
    }
  }
  if (pkIndex.value < 0 || pkIndex.value >= rows.length) {
    ElMessage.error('请勾选一个主键字段（pk 必须恰有一个）')
    return null
  }
  return JSON.stringify(
    rows.map((f, i) => {
      const def: { name: string; type: string; comment?: string; pk?: boolean } = {
        name: f.name.trim(),
        type: f.type,
      }
      if (f.comment.trim()) def.comment = f.comment.trim()
      if (i === pkIndex.value) def.pk = true
      return def
    }),
  )
}

/** dim_json 可选校验（对齐后端 ValidateDimJson：[{dim_table,join_key}] 且两项均非空） */
function buildDimJson(): string | null {
  const raw = mdlForm.dim_json.trim()
  if (!raw) return ''
  try {
    const v: unknown = JSON.parse(raw)
    if (!Array.isArray(v)) {
      ElMessage.error('dim_json 须为 [{dim_table,join_key}] JSON 数组')
      return null
    }
    const dims: { dim_table: string; join_key: string }[] = []
    for (let i = 0; i < v.length; i++) {
      const it = v[i] as { dim_table?: unknown; join_key?: unknown }
      const dt = String(it?.dim_table ?? '').trim()
      const jk = String(it?.join_key ?? '').trim()
      if (!dt || !jk) {
        ElMessage.error(`dim_json 第 ${i + 1} 项 dim_table/join_key 不能为空`)
        return null
      }
      dims.push({ dim_table: dt, join_key: jk })
    }
    return JSON.stringify(dims)
  } catch {
    ElMessage.error('dim_json 不是合法 JSON，请检查格式')
    return null
  }
}

async function onSaveMdl() {
  const valid = await mdlFormRef.value?.validate().catch(() => false)
  if (!valid) return
  // <layer>_<table> 整体长度把守（部署时按此真实建表）
  if (physicalPreview.value.length > IDENT_MAX) {
    ElMessage.error(`物理表名 ${physicalPreview.value} 超 ${IDENT_MAX} 字符，请缩短表名`)
    return
  }
  const fieldsJson = buildFieldsJson()
  if (fieldsJson === null) return
  const dimJson = buildDimJson()
  if (dimJson === null) return
  mdlSaving.value = true
  try {
    await saveThemeModel({
      theme_id: mdlForm.theme_id as number,
      layer: mdlForm.layer,
      table_name: mdlForm.table_name.trim(),
      model_name: mdlForm.model_name.trim(),
      fields_json: fieldsJson,
      dim_json: dimJson || undefined,
      datasource_id: mdlForm.datasource_id as number,
      db_name: mdlForm.db_name.trim() || undefined,
    })
    ElMessage.success('模型创建成功，可点击「部署」在目标库真实建表')
    mdlDialogVisible.value = false
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '保存模型失败'))
  } finally {
    mdlSaving.value = false
  }
}

async function onDeployMdl(row: ThemeModelInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定部署模型「${row.model_name || row.table_name}」吗？将在目标数据源执行幂等建表 DDL（create table if not exists ${physicalTable(row)}）。`,
    '部署确认',
    { confirmButtonText: '部署', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await deployThemeModel(row.id)
    if (resp.ok) {
      // 成功 msg 携带生成 SQL 摘要（模型 <表> 部署成功，DDL: ...）
      ElMessageBox.alert(
        `部署成功，耗时 ${resp.cost_ms ?? '-'} ms${resp.msg ? `\n\n${resp.msg}` : ''}`,
        '部署结果',
        { confirmButtonText: '知道了', type: 'success' },
      ).catch(() => {})
    } else {
      // 失败细节（数据源/DDL 报错）随正常响应回告，直接展示原因
      ElMessageBox.alert(resp.msg || '部署失败，未返回原因', '部署结果', {
        confirmButtonText: '知道了',
        type: 'error',
      }).catch(() => {})
    }
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '部署请求失败'))
  }
}

// 查看建表 DDL（create_sql 由后端按 fields_json 实时生成，随列表返回）
const sqlVisible = ref(false)
const sqlModelName = ref('')
const sqlText = ref('')

function openCreateSql(row: ThemeModelInfo) {
  sqlModelName.value = row.model_name || row.table_name || `#${row.id ?? ''}`
  sqlText.value = row.create_sql || ''
  sqlVisible.value = true
}

async function onDeleteMdl(row: ThemeModelInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除模型「${row.model_name || row.table_name}」吗？仅删除模型定义，目标库中已建的物理表不受影响。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteThemeModel(row.id)
    ElMessage.success('删除成功')
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '删除模型失败'))
  }
}

// ---------- 物化视图 ----------
const mvLoading = ref(false)
const mvList = ref<ThemeMvInfo[]>([])
const mvTotal = ref(0)
const mvQuery = reactive({ page: 1, size: 10 })

/** 刷新结果三色：success 绿 / failed 红 / 空（尚未刷新）灰 */
const mvStatusTag: Record<string, { label: string; type: 'success' | 'danger' | 'info' }> = {
  success: { label: '成功', type: 'success' },
  failed: { label: '失败', type: 'danger' },
}

function mvStatusOf(s?: string) {
  return mvStatusTag[s || ''] || { label: '未刷新', type: 'info' as const }
}

const refreshModeLabel: Record<string, string> = {
  manual: '手动刷新',
  schedule: '定时刷新',
}

function qualifiedView(row: Pick<ThemeMvInfo, 'db_name' | 'view_name'>) {
  const db = (row.db_name || '').trim()
  return db ? `${db}.${row.view_name || ''}` : row.view_name || '-'
}

async function fetchMvs() {
  mvLoading.value = true
  try {
    const resp = await themeMvList({ page: mvQuery.page, size: mvQuery.size })
    mvList.value = resp.list || []
    mvTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取物化视图列表失败'))
  } finally {
    mvLoading.value = false
  }
}

const mvDialogVisible = ref(false)
const mvSaving = ref(false)
const mvFormRef = ref<FormInstance>()
const mvForm = reactive({
  name: '',
  datasource_id: undefined as number | undefined,
  db_name: '',
  view_name: '',
  select_sql: '',
  refresh_mode: 'manual',
})
const mvRules: FormRules = {
  name: [{ required: true, message: '请输入视图名称', trigger: 'blur' }],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  view_name: [
    { required: true, message: '请输入视图名', trigger: 'blur' },
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
  select_sql: [{ required: true, message: '请输入查询语句', trigger: 'blur' }],
  db_name: [
    { pattern: IDENT_RE, message: '须以小写字母开头，仅含小写字母/数字/下划线', trigger: 'blur' },
  ],
}

function openMvCreate() {
  Object.assign(mvForm, {
    name: '',
    datasource_id: undefined,
    db_name: '',
    view_name: '',
    select_sql: '',
    refresh_mode: 'manual',
  })
  mvDialogVisible.value = true
}

async function onSaveMv() {
  const valid = await mvFormRef.value?.validate().catch(() => false)
  if (!valid) return
  // select_sql 校验（对齐后端 ValidateSelectSql：SELECT 开头 + 无分号，执行时与 create view 拼接）
  const sql = mvForm.select_sql.trim()
  if (!/^select/i.test(sql)) {
    ElMessage.error('select_sql 须以 SELECT 开头')
    return
  }
  if (sql.includes(';')) {
    ElMessage.error('select_sql 不允许包含分号（防多语句注入）')
    return
  }
  mvSaving.value = true
  try {
    await saveThemeMv({
      name: mvForm.name.trim(),
      datasource_id: mvForm.datasource_id as number,
      db_name: mvForm.db_name.trim() || undefined,
      view_name: mvForm.view_name.trim(),
      select_sql: sql,
      refresh_mode: mvForm.refresh_mode || undefined,
    })
    ElMessage.success('物化视图创建成功，可点击「刷新」真实建视图')
    mvDialogVisible.value = false
    fetchMvs()
  } catch (e) {
    ElMessage.error(errText(e, '保存物化视图失败'))
  } finally {
    mvSaving.value = false
  }
}

async function onRefreshMv(row: ThemeMvInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定刷新物化视图「${row.name || row.view_name}」吗？将在目标库执行 create or replace view 原子重建。`,
    '刷新确认',
    { confirmButtonText: '刷新', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await refreshThemeMv(row.id)
    if (resp.ok) {
      ElMessageBox.alert(
        `刷新成功，耗时 ${resp.cost_ms ?? '-'} ms${resp.msg ? `\n\n${resp.msg}` : ''}`,
        '刷新结果',
        { confirmButtonText: '知道了', type: 'success' },
      ).catch(() => {})
    } else {
      // 失败原因（select_sql 非法/执行报错）随正常响应回告，且 last_status 已落 failed
      ElMessageBox.alert(resp.msg || '刷新失败，未返回原因', '刷新结果', {
        confirmButtonText: '知道了',
        type: 'error',
      }).catch(() => {})
    }
    fetchMvs()
  } catch (e) {
    ElMessage.error(errText(e, '刷新请求失败'))
  }
}

async function onDeleteMv(row: ThemeMvInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除物化视图「${row.name || row.view_name}」吗？仅删除定义存档，目标库中已建的视图不受影响。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteThemeMv(row.id)
    ElMessage.success('删除成功')
    fetchMvs()
  } catch (e) {
    ElMessage.error(errText(e, '删除物化视图失败'))
  }
}

onMounted(() => {
  fetchDomains()
  fetchThemeOptions()
  fetchDsOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ===== 主题域 ===== -->
        <el-tab-pane label="主题域" name="domain">
          <div class="toolbar">
            <el-button type="primary" @click="openDmCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建主题域
            </el-button>
            <el-button @click="fetchDomains">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
            <span class="form-tip">共 {{ dmList.length }} 个域；域下存在主题时不可删除</span>
          </div>

          <el-table v-loading="dmLoading" :data="dmList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="code" label="域编码" min-width="140" show-overflow-tooltip />
            <el-table-column prop="name" label="域名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="comment" label="备注" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.comment || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteDm(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- ===== 主题 ===== -->
        <el-tab-pane label="主题" name="theme">
          <div class="toolbar">
            <el-select
              v-model="thQuery.domain_id"
              placeholder="全部主题域"
              clearable
              filterable
              style="width: 200px"
              @change="thSearch"
            >
              <el-option
                v-for="d in dmList"
                :key="d.id"
                :label="`${d.name}（${d.code}）`"
                :value="d.id as number"
              />
            </el-select>
            <el-button type="primary" @click="openThCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建主题
            </el-button>
            <el-button @click="fetchThemes">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="thLoading" :data="thList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="code" label="主题编码" min-width="140" show-overflow-tooltip />
            <el-table-column prop="name" label="主题名称" min-width="160" show-overflow-tooltip />
            <el-table-column label="所属域" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">
                {{ (row.domain_id && dmList.find((d) => d.id === row.domain_id)?.name) || row.domain_id || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="comment" label="备注" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.comment || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteTh(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="thQuery.page"
            v-model:page-size="thQuery.size"
            :total="thTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchThemes"
            @size-change="thSearch"
          />
        </el-tab-pane>

        <!-- ===== 分层模型 ===== -->
        <el-tab-pane label="分层模型" name="model">
          <div class="toolbar">
            <el-select
              v-model="mdlQuery.theme_id"
              placeholder="全部主题"
              clearable
              filterable
              style="width: 200px"
              @change="mdlSearch"
            >
              <el-option
                v-for="t in themeOptions"
                :key="t.id"
                :label="t.name || t.code"
                :value="t.id as number"
              />
            </el-select>
            <el-select
              v-model="mdlQuery.layer"
              placeholder="全部分层"
              clearable
              style="width: 140px"
              @change="mdlSearch"
            >
              <el-option v-for="l in layerOptions" :key="l.value" :label="l.label" :value="l.value" />
            </el-select>
            <el-button type="primary" @click="openMdlCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建模型
            </el-button>
            <el-button @click="fetchModels">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="mdlLoading" :data="mdlList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="model_name" label="模型名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="分层" width="90">
              <template #default="{ row }">
                <el-tag size="small" :class="layerTagClass[row.layer] || ''">{{ row.layer || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="物理表" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <span>{{ physicalTable(row) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="所属主题" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ themeNameOf(row.theme_id) }}</template>
            </el-table-column>
            <el-table-column label="数据源" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ dsNameOf(row.datasource_id) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="170" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openCreateSql(row)">SQL</el-button>
                <el-button link type="success" @click="onDeployMdl(row)">部署</el-button>
                <el-button link type="danger" @click="onDeleteMdl(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="mdlQuery.page"
            v-model:page-size="mdlQuery.size"
            :total="mdlTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchModels"
            @size-change="mdlSearch"
          />
        </el-tab-pane>

        <!-- ===== 物化视图 ===== -->
        <el-tab-pane label="物化视图" name="mv">
          <div class="toolbar">
            <el-button type="primary" @click="openMvCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建物化视图
            </el-button>
            <el-button @click="fetchMvs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
            <span class="form-tip">刷新 = 在目标库执行 create or replace view 真实重建</span>
          </div>

          <el-table v-loading="mvLoading" :data="mvList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="视图" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ qualifiedView(row) }}</template>
            </el-table-column>
            <el-table-column label="数据源" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ dsNameOf(row.datasource_id) }}</template>
            </el-table-column>
            <el-table-column label="刷新模式" width="100">
              <template #default="{ row }">
                <el-tag size="small" :type="row.refresh_mode === 'schedule' ? 'primary' : 'info'">
                  {{ refreshModeLabel[row.refresh_mode] || row.refresh_mode || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="最近状态" width="100">
              <template #default="{ row }">
                <el-tag :type="mvStatusOf(row.last_status).type" size="small">
                  {{ mvStatusOf(row.last_status).label }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="最近刷新" width="170">
              <template #default="{ row }">{{ fmtTime(row.last_refresh_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="130" fixed="right">
              <template #default="{ row }">
                <el-button link type="success" @click="onRefreshMv(row)">刷新</el-button>
                <el-button link type="danger" @click="onDeleteMv(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="mvQuery.page"
            v-model:page-size="mvQuery.size"
            :total="mvTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchMvs"
            @size-change="() => { mvQuery.page = 1; fetchMvs() }"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 主题域新建 -->
    <el-dialog v-model="dmDialogVisible" title="新建主题域" width="520px" destroy-on-close>
      <el-form ref="dmFormRef" :model="dmForm" :rules="dmRules" label-width="90px">
        <el-form-item label="域编码" prop="code">
          <el-input v-model="dmForm.code" placeholder="如：transaction（小写标识符）" maxlength="64" />
        </el-form-item>
        <el-form-item label="域名称" prop="name">
          <el-input v-model="dmForm.name" placeholder="如：交易域" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="dmForm.comment" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dmDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dmSaving" @click="onSaveDm">保存</el-button>
      </template>
    </el-dialog>

    <!-- 主题新建 -->
    <el-dialog v-model="thDialogVisible" title="新建主题" width="520px" destroy-on-close>
      <el-form ref="thFormRef" :model="thForm" :rules="thRules" label-width="90px">
        <el-form-item label="所属域" prop="domain_id">
          <el-select v-model="thForm.domain_id" placeholder="选择主题域" filterable style="width: 100%">
            <el-option
              v-for="d in dmList"
              :key="d.id"
              :label="`${d.name}（${d.code}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主题编码" prop="code">
          <el-input v-model="thForm.code" placeholder="如：order（小写标识符）" maxlength="64" />
        </el-form-item>
        <el-form-item label="主题名称" prop="name">
          <el-input v-model="thForm.name" placeholder="如：订单主题" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="thForm.comment" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="thDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="thSaving" @click="onSaveTh">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分层模型新建（字段动态行编辑 + dim_json） -->
    <el-dialog v-model="mdlDialogVisible" title="新建分层模型" width="760px" destroy-on-close>
      <el-form ref="mdlFormRef" :model="mdlForm" :rules="mdlRules" label-width="100px">
        <el-form-item label="所属主题" prop="theme_id">
          <el-select v-model="mdlForm.theme_id" placeholder="选择主题" filterable style="width: 100%">
            <el-option
              v-for="t in themeOptions"
              :key="t.id"
              :label="`${t.name}（${t.code}）`"
              :value="t.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="数仓分层" prop="layer">
          <el-select v-model="mdlForm.layer" style="width: 100%">
            <el-option v-for="l in layerOptions" :key="l.value" :label="l.label" :value="l.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="表名" prop="table_name">
          <el-input v-model="mdlForm.table_name" placeholder="如：order_detail（不含分层前缀）" maxlength="64">
            <template #append>{{ mdlForm.layer }}_</template>
          </el-input>
          <div v-if="physicalPreview" class="form-tip">
            物理表名：<b>{{ physicalPreview }}</b>（部署时按此真实建表，整体不超 {{ IDENT_MAX }} 字符）
          </div>
        </el-form-item>
        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="mdlForm.model_name" placeholder="如：订单明细模型" maxlength="64" />
        </el-form-item>
        <el-form-item label="目标数据源" prop="datasource_id">
          <el-select v-model="mdlForm.datasource_id" placeholder="部署建表的目标库" filterable style="width: 100%">
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（${d.type}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标库名">
          <el-input v-model="mdlForm.db_name" placeholder="留空取数据源默认库" maxlength="64" />
        </el-form-item>
        <el-form-item label="字段定义">
          <div class="fields-editor">
            <div v-for="(f, idx) in mdlFields" :key="idx" class="field-row">
              <el-input v-model="f.name" placeholder="字段名，如 order_id" style="width: 190px" />
              <el-select v-model="f.type" style="width: 110px">
                <el-option v-for="t in FIELD_TYPES" :key="t" :label="t" :value="t" />
              </el-select>
              <el-input v-model="f.comment" placeholder="说明（可选）" style="width: 150px" />
              <el-radio v-model="pkIndex" :value="idx">主键</el-radio>
              <el-button link type="danger" @click="removeFieldRow(idx)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <div class="field-actions">
              <el-button type="primary" plain size="small" @click="addFieldRow">
                <el-icon><Plus /></el-icon>&nbsp;添加字段
              </el-button>
            </div>
            <div class="form-tip">
              主键（pk）必须恰有一个；类型白名单 {{ FIELD_TYPES.join('/') }}
            </div>
          </div>
        </el-form-item>
        <el-form-item label="维度声明">
          <el-input
            v-model="mdlForm.dim_json"
            type="textarea"
            :rows="3"
            placeholder='可选，JSON 数组，如：[{"dim_table":"dim_user","join_key":"user_id"}]'
          />
          <div class="form-tip">dim_table/join_key 均不能为空</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mdlDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mdlSaving" @click="onSaveMdl">保存</el-button>
      </template>
    </el-dialog>

    <!-- 物化视图新建 -->
    <el-dialog v-model="mvDialogVisible" title="新建物化视图" width="640px" destroy-on-close>
      <el-form ref="mvFormRef" :model="mvForm" :rules="mvRules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="mvForm.name" placeholder="如：订单汇总视图" maxlength="128" />
        </el-form-item>
        <el-form-item label="数据源" prop="datasource_id">
          <el-select v-model="mvForm.datasource_id" placeholder="建视图的目标库" filterable style="width: 100%">
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（${d.type}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="库名">
          <el-input v-model="mvForm.db_name" placeholder="留空取数据源默认库" maxlength="64" />
        </el-form-item>
        <el-form-item label="视图名" prop="view_name">
          <el-input v-model="mvForm.view_name" placeholder="如：v_order_summary（小写标识符）" maxlength="64" />
        </el-form-item>
        <el-form-item label="查询语句" prop="select_sql">
          <el-input
            v-model="mvForm.select_sql"
            type="textarea"
            :rows="6"
            placeholder="SELECT 开头，不含分号，如：SELECT user_id, COUNT(*) AS cnt FROM dwd_order GROUP BY user_id"
          />
          <div class="form-tip">刷新时拼入 create or replace view，须以 SELECT 开头且不含分号</div>
        </el-form-item>
        <el-form-item label="刷新模式">
          <el-select v-model="mvForm.refresh_mode" style="width: 100%">
            <el-option label="手动刷新（manual）" value="manual" />
            <el-option label="定时刷新（schedule，存档标记）" value="schedule" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mvDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mvSaving" @click="onSaveMv">保存</el-button>
      </template>
    </el-dialog>

    <!-- 建表 DDL 查看 -->
    <el-dialog v-model="sqlVisible" :title="`建表 DDL：${sqlModelName}`" width="680px">
      <pre v-if="sqlText" class="sql-pre">{{ sqlText }}</pre>
      <el-empty v-else description="create_sql 为空（fields_json 缺损或生成失败）" :image-size="70" />
      <template #footer>
        <el-button @click="sqlVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
.form-tip {
  font-size: 12px;
  color: #909399;
}
.fields-editor {
  width: 100%;
}
.field-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.field-actions {
  margin: 2px 0 6px;
}
.sql-pre {
  margin: 0;
  max-height: 420px;
  overflow: auto;
  padding: 10px 12px;
  background: #282c34;
  color: #98c379;
  border-radius: 6px;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 分层五色 tag：ods紫 / dwd蓝 / dws青 / ads绿 / dim橙 */
.el-tag.layer-ods {
  background-color: #f3e8ff;
  border-color: #d8b4fe;
  color: #7c3aed;
}
.el-tag.layer-dwd {
  background-color: #dbeafe;
  border-color: #93c5fd;
  color: #1d4ed8;
}
.el-tag.layer-dws {
  background-color: #cffafe;
  border-color: #67e8f9;
  color: #0e7490;
}
.el-tag.layer-ads {
  background-color: #dcfce7;
  border-color: #86efac;
  color: #15803d;
}
.el-tag.layer-dim {
  background-color: #ffedd5;
  border-color: #fdba74;
  color: #c2620a;
}
</style>
