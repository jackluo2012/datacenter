<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  saveWordRoot,
  deleteWordRoot,
  wordRootList,
  saveDict,
  dictInfo,
  deleteDict,
  dictList,
  saveElement,
  deleteElement,
  elementList,
  saveStdModel,
  stdModelInfo,
  deleteStdModel,
  stdModelList,
  publishStdModel,
  stdCheck,
  stdCheckList,
  type StdWordRoot,
  type StdDict,
  type StdDictItem,
  type StdElement,
  type StdModel,
  type StdCheckReport,
} from '@/api/standard'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共 ----------
const activeTab = ref('wordroot')

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

/** 数据源下拉（核对页签用），进页面拉一次 */
const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 200 })
    dsOptions.value = resp.list || []
  } catch {
    /* 下拉数据拉取失败不阻断页面 */
  }
}

/** 字典编码下拉（数据元值域引用用），进页面拉一次 */
const dictCodeOptions = ref<string[]>([])
async function fetchDictCodeOptions() {
  try {
    const resp = await dictList({ page: 1, size: 200 })
    dictCodeOptions.value = (resp.list || []).map((d) => d.dict_code || '').filter(Boolean)
  } catch {
    /* 忽略 */
  }
}

function onTabChange(name: string | number) {
  const tab = String(name)
  if (tab === 'wordroot') fetchWordRoots()
  else if (tab === 'dict') fetchDicts()
  else if (tab === 'element') fetchElements()
  else if (tab === 'model') fetchModels()
  else if (tab === 'check') {
    fetchPublishedModels()
    fetchCheckHistory()
  }
}

// ---------- 词根 ----------
const wrLoading = ref(false)
const wrList = ref<StdWordRoot[]>([])
const wrTotal = ref(0)
const wrQuery = reactive({ keyword: '', page: 1, size: 10 })

const typeOptions = [
  { value: 'common', label: '通用词根' },
  { value: 'attribute', label: '修饰词根' },
  { value: 'dimension', label: '维度词根' },
  { value: 'indicator', label: '指标词根' },
]

async function fetchWordRoots() {
  wrLoading.value = true
  try {
    const resp = await wordRootList({ ...wrQuery, keyword: wrQuery.keyword || undefined })
    wrList.value = resp.list || []
    wrTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取词根列表失败'))
  } finally {
    wrLoading.value = false
  }
}

function wrSearch() {
  wrQuery.page = 1
  fetchWordRoots()
}

const wrDialogVisible = ref(false)
const wrSaving = ref(false)
const wrEditingId = ref<number | null>(null)
const wrFormRef = ref<FormInstance>()
const wrForm = reactive({ name: '', cn_name: '', type: 'common', comment: '' })
const wrRules: FormRules = {
  name: [{ required: true, message: '请输入词根名称', trigger: 'blur' }],
}

function openWrCreate() {
  wrEditingId.value = null
  Object.assign(wrForm, { name: '', cn_name: '', type: 'common', comment: '' })
  wrDialogVisible.value = true
}

function openWrEdit(row: StdWordRoot) {
  wrEditingId.value = row.id ?? null
  Object.assign(wrForm, {
    name: row.name || '',
    cn_name: row.cn_name || '',
    type: row.type || 'common',
    comment: row.comment || '',
  })
  wrDialogVisible.value = true
}

async function onSaveWr() {
  const valid = await wrFormRef.value?.validate().catch(() => false)
  if (!valid) return
  wrSaving.value = true
  try {
    await saveWordRoot({
      id: wrEditingId.value ?? undefined,
      name: wrForm.name.trim(),
      cn_name: wrForm.cn_name.trim() || undefined,
      type: wrForm.type || undefined,
      comment: wrForm.comment.trim() || undefined,
    })
    ElMessage.success(wrEditingId.value === null ? '词根创建成功' : '词根更新成功')
    wrDialogVisible.value = false
    fetchWordRoots()
  } catch (e) {
    ElMessage.error(errText(e, '保存词根失败'))
  } finally {
    wrSaving.value = false
  }
}

async function onDeleteWr(row: StdWordRoot) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除词根「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteWordRoot(row.id)
    ElMessage.success('删除成功')
    fetchWordRoots()
  } catch (e) {
    ElMessage.error(errText(e, '删除词根失败'))
  }
}

// ---------- 标准字典 ----------
const dictLoading = ref(false)
const dictListData = ref<StdDict[]>([])
const dictTotal = ref(0)
const dictQuery = reactive({ keyword: '', page: 1, size: 10 })

async function fetchDicts() {
  dictLoading.value = true
  try {
    const resp = await dictList({ ...dictQuery, keyword: dictQuery.keyword || undefined })
    dictListData.value = resp.list || []
    dictTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取字典列表失败'))
  } finally {
    dictLoading.value = false
  }
}

function dictSearch() {
  dictQuery.page = 1
  fetchDicts()
}

interface DictForm {
  dict_code: string
  dict_name: string
  comment: string
  items: StdDictItem[]
}

const dictDialogVisible = ref(false)
const dictSaving = ref(false)
const dictEditingId = ref<number | null>(null)
const dictFormRef = ref<FormInstance>()
const dictForm = reactive<DictForm>({ dict_code: '', dict_name: '', comment: '', items: [] })
const dictRules: FormRules = {
  dict_code: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
  dict_name: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
}

function emptyDictForm() {
  Object.assign(dictForm, { dict_code: '', dict_name: '', comment: '', items: [] })
}

function openDictCreate() {
  dictEditingId.value = null
  emptyDictForm()
  addDictItem()
  dictDialogVisible.value = true
}

/** 编辑：基本信息取列表行，字典项走详情接口补全 */
async function openDictEdit(row: StdDict) {
  if (row.id === undefined) return
  dictEditingId.value = Number(row.id)
  emptyDictForm()
  Object.assign(dictForm, {
    dict_code: row.dict_code || '',
    dict_name: row.dict_name || '',
    comment: row.comment || '',
  })
  dictDialogVisible.value = true
  try {
    const detail = await dictInfo(Number(row.id))
    dictForm.items = (detail.items || []).map((it) => ({ ...it }))
    if (dictForm.items.length === 0) addDictItem()
  } catch (e) {
    ElMessage.warning(errText(e, '加载字典项失败，请保存前自行补全'))
  }
}

function addDictItem() {
  dictForm.items.push({ item_code: '', item_value: '', sort: dictForm.items.length + 1 })
}

function removeDictItem(idx: number) {
  dictForm.items.splice(idx, 1)
}

async function onSaveDict() {
  const valid = await dictFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const items = dictForm.items
  if (items.length === 0) {
    ElMessage.warning('请至少添加一个字典项')
    return
  }
  for (let i = 0; i < items.length; i++) {
    if (!items[i].item_code?.trim() || !items[i].item_value?.trim()) {
      ElMessage.warning(`第 ${i + 1} 行字典项的编码与值不能为空`)
      return
    }
  }
  dictSaving.value = true
  try {
    await saveDict({
      id: dictEditingId.value ?? undefined,
      dict_code: dictForm.dict_code.trim(),
      dict_name: dictForm.dict_name.trim(),
      comment: dictForm.comment.trim() || undefined,
      items: items.map((it, i) => ({
        item_code: (it.item_code || '').trim(),
        item_value: (it.item_value || '').trim(),
        sort: it.sort ?? i + 1,
      })),
    })
    ElMessage.success(dictEditingId.value === null ? '字典创建成功' : '字典更新成功')
    dictDialogVisible.value = false
    fetchDicts()
    fetchDictCodeOptions()
  } catch (e) {
    ElMessage.error(errText(e, '保存字典失败'))
  } finally {
    dictSaving.value = false
  }
}

async function onDeleteDict(row: StdDict) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除字典「${row.dict_name || row.dict_code}」吗？其字典项将一并删除。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteDict(Number(row.id))
    ElMessage.success('删除成功')
    fetchDicts()
    fetchDictCodeOptions()
  } catch (e) {
    ElMessage.error(errText(e, '删除字典失败'))
  }
}

// 字典详情（含字典项）
const dictDetailVisible = ref(false)
const dictDetailLoading = ref(false)
const dictDetail = ref<StdDict | null>(null)

async function openDictDetail(row: StdDict) {
  if (row.id === undefined) return
  dictDetailVisible.value = true
  dictDetailLoading.value = true
  try {
    dictDetail.value = await dictInfo(Number(row.id))
  } catch (e) {
    ElMessage.error(errText(e, '获取字典详情失败'))
    dictDetailVisible.value = false
  } finally {
    dictDetailLoading.value = false
  }
}

// ---------- 数据元 ----------
const elLoading = ref(false)
const elList = ref<StdElement[]>([])
const elTotal = ref(0)
const elQuery = reactive({ keyword: '', page: 1, size: 10 })

const dataTypeOptions = ['varchar', 'int', 'bigint', 'decimal', 'date', 'datetime', 'text']

async function fetchElements() {
  elLoading.value = true
  try {
    const resp = await elementList({ ...elQuery, keyword: elQuery.keyword || undefined })
    elList.value = resp.list || []
    elTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取数据元列表失败'))
  } finally {
    elLoading.value = false
  }
}

function elSearch() {
  elQuery.page = 1
  fetchElements()
}

const elDialogVisible = ref(false)
const elSaving = ref(false)
const elEditingId = ref<number | null>(null)
const elFormRef = ref<FormInstance>()
const elForm = reactive({
  element_code: '',
  element_name: '',
  data_type: 'varchar',
  data_length: undefined as number | undefined,
  data_precision: undefined as number | undefined,
  dict_code: '',
  comment: '',
})
const elRules: FormRules = {
  element_code: [{ required: true, message: '请输入数据元编码', trigger: 'blur' }],
  element_name: [{ required: true, message: '请输入数据元名称', trigger: 'blur' }],
  data_type: [{ required: true, message: '请选择数据类型', trigger: 'change' }],
}

function openElCreate() {
  elEditingId.value = null
  Object.assign(elForm, {
    element_code: '',
    element_name: '',
    data_type: 'varchar',
    data_length: undefined,
    data_precision: undefined,
    dict_code: '',
    comment: '',
  })
  elDialogVisible.value = true
}

function openElEdit(row: StdElement) {
  elEditingId.value = row.id ?? null
  Object.assign(elForm, {
    element_code: row.element_code || '',
    element_name: row.element_name || '',
    data_type: row.data_type || 'varchar',
    data_length: row.data_length || undefined,
    data_precision: row.data_precision || undefined,
    dict_code: row.dict_code || '',
    comment: row.comment || '',
  })
  elDialogVisible.value = true
}

async function onSaveEl() {
  const valid = await elFormRef.value?.validate().catch(() => false)
  if (!valid) return
  // varchar/decimal 后端要求必填长度，前端先行校验
  if ((elForm.data_type === 'varchar' || elForm.data_type === 'decimal') && !elForm.data_length) {
    ElMessage.warning(`${elForm.data_type} 类型必须填写长度`)
    return
  }
  elSaving.value = true
  try {
    await saveElement({
      id: elEditingId.value ?? undefined,
      element_code: elForm.element_code.trim(),
      element_name: elForm.element_name.trim(),
      data_type: elForm.data_type,
      data_length: elForm.data_length || undefined,
      data_precision: elForm.data_precision || undefined,
      dict_code: elForm.dict_code || undefined,
      comment: elForm.comment.trim() || undefined,
    })
    ElMessage.success(elEditingId.value === null ? '数据元创建成功' : '数据元更新成功')
    elDialogVisible.value = false
    fetchElements()
  } catch (e) {
    ElMessage.error(errText(e, '保存数据元失败'))
  } finally {
    elSaving.value = false
  }
}

async function onDeleteEl(row: StdElement) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除数据元「${row.element_name || row.element_code}」吗？`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteElement(row.id)
    ElMessage.success('删除成功')
    fetchElements()
  } catch (e) {
    ElMessage.error(errText(e, '删除数据元失败'))
  }
}

// ---------- 标准模型 ----------
const mdlLoading = ref(false)
const mdlList = ref<StdModel[]>([])
const mdlTotal = ref(0)
const mdlQuery = reactive({ status: '', keyword: '', page: 1, size: 10 })

const statusTag: Record<string, { label: string; type: 'info' | 'success' }> = {
  draft: { label: '草稿', type: 'info' },
  published: { label: '已发布', type: 'success' },
}

async function fetchModels() {
  mdlLoading.value = true
  try {
    const resp = await stdModelList({
      ...mdlQuery,
      status: mdlQuery.status || undefined,
      keyword: mdlQuery.keyword || undefined,
    })
    mdlList.value = resp.list || []
    mdlTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取标准模型列表失败'))
  } finally {
    mdlLoading.value = false
  }
}

function mdlSearch() {
  mdlQuery.page = 1
  fetchModels()
}

const mdlDialogVisible = ref(false)
const mdlSaving = ref(false)
const mdlEditingId = ref<number | null>(null)
const mdlFormRef = ref<FormInstance>()
const mdlForm = reactive({ model_code: '', model_name: '', elements_json: '' })
const mdlRules: FormRules = {
  model_code: [{ required: true, message: '请输入模型编码', trigger: 'blur' }],
  model_name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
}

function openMdlCreate() {
  mdlEditingId.value = null
  Object.assign(mdlForm, { model_code: '', model_name: '', elements_json: '' })
  mdlDialogVisible.value = true
}

function openMdlEdit(row: StdModel) {
  mdlEditingId.value = row.id ?? null
  Object.assign(mdlForm, {
    model_code: row.model_code || '',
    model_name: row.model_name || '',
    elements_json: row.elements_json || '',
  })
  mdlDialogVisible.value = true
}

/** elements_json 提交前 parse 校验：数组且 element_code 非空不重复（对齐后端 ParseElementsJson） */
function validateElementsJson(): { element_code: string; required: boolean; pk: boolean }[] | null {
  const raw = mdlForm.elements_json.trim()
  if (!raw) return []
  try {
    const v = JSON.parse(raw)
    if (!Array.isArray(v)) {
      ElMessage.error('elements_json 须为 JSON 数组')
      return null
    }
    const seen = new Set<string>()
    const out: { element_code: string; required: boolean; pk: boolean }[] = []
    for (const it of v) {
      const code = String(it?.element_code ?? '').trim()
      if (!code) {
        ElMessage.error('elements_json 每项必须包含非空 element_code')
        return null
      }
      if (seen.has(code)) {
        ElMessage.error(`elements_json 中 element_code 重复：${code}`)
        return null
      }
      seen.add(code)
      out.push({ element_code: code, required: !!it.required, pk: !!it.pk })
    }
    return out
  } catch {
    ElMessage.error('elements_json 不是合法 JSON，请检查格式')
    return null
  }
}

async function onSaveMdl() {
  const valid = await mdlFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const elements = validateElementsJson()
  if (elements === null) return
  mdlSaving.value = true
  try {
    await saveStdModel({
      id: mdlEditingId.value ?? undefined,
      model_code: mdlForm.model_code.trim(),
      model_name: mdlForm.model_name.trim(),
      elements_json: JSON.stringify(elements),
    })
    ElMessage.success(mdlEditingId.value === null ? '模型创建成功' : '模型更新成功')
    mdlDialogVisible.value = false
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '保存标准模型失败'))
  } finally {
    mdlSaving.value = false
  }
}

async function onPublishMdl(row: StdModel) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    row.status === 'published'
      ? `模型「${row.model_name || row.model_code}」已发布，重新发布将版本+1 并刷新快照。确定继续吗？`
      : `确定发布标准模型「${row.model_name || row.model_code}」吗？发布后将生成版本快照。`,
    '发布确认',
    { confirmButtonText: '发布', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await publishStdModel({ id: row.id, operator: 'console' })
    ElMessage.success(`发布成功，当前版本 v${resp.version}`)
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '发布失败'))
  }
}

async function onDeleteMdl(row: StdModel) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除标准模型「${row.model_name || row.model_code}」吗？`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteStdModel(row.id)
    ElMessage.success('删除成功')
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '删除标准模型失败'))
  }
}

// 查看发布快照
const snapVisible = ref(false)
const snapLoading = ref(false)
const snapJson = ref('')
const snapModel = ref('')

function prettyJson(raw?: string) {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function openSnapshot(row: StdModel) {
  if (row.id === undefined) return
  snapModel.value = row.model_name || row.model_code || ''
  snapVisible.value = true
  snapLoading.value = true
  try {
    const detail = await stdModelInfo(row.id)
    snapJson.value = detail.snapshot_json || ''
    if (!snapJson.value) {
      ElMessage.info('该模型尚未发布，暂无快照')
    }
  } catch (e) {
    ElMessage.error(errText(e, '获取模型快照失败'))
    snapVisible.value = false
  } finally {
    snapLoading.value = false
  }
}

// ---------- 落标核对 ----------
const checkFormRef = ref<FormInstance>()
const checkForm = reactive({
  model_id: undefined as number | undefined,
  datasource_id: undefined as number | undefined,
  db_name: '',
  table_name: '',
})
const checkRules: FormRules = {
  model_id: [{ required: true, message: '请选择标准模型', trigger: 'change' }],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  table_name: [{ required: true, message: '请输入表名（仅字母/数字/下划线）', trigger: 'blur' }],
}

const publishedModels = ref<StdModel[]>([])
async function fetchPublishedModels() {
  try {
    const resp = await stdModelList({ status: 'published', page: 1, size: 200 })
    publishedModels.value = resp.list || []
  } catch {
    /* 忽略，表单下拉留空 */
  }
}

const checking = ref(false)
const checkReport = ref<StdCheckReport | null>(null)

interface DiffRow {
  column?: string
  expect?: string
  actual?: string
  result?: string
}

const diffRows = ref<DiffRow[]>([])

const resultTag: Record<string, { label: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
  match: { label: '匹配', type: 'success' },
  missing: { label: '缺失', type: 'danger' },
  type_mismatch: { label: '类型不一致', type: 'warning' },
}

function resultOf(r?: string) {
  return resultTag[r || ''] || { label: r || '未知', type: 'info' }
}

function parseDiff(raw?: string): DiffRow[] {
  if (!raw) return []
  try {
    const v = JSON.parse(raw)
    return Array.isArray(v) ? (v as DiffRow[]) : []
  } catch {
    return []
  }
}

async function onRunCheck() {
  const valid = await checkFormRef.value?.validate().catch(() => false)
  if (!valid) return
  checking.value = true
  try {
    const resp = await stdCheck({
      model_id: checkForm.model_id as number,
      datasource_id: checkForm.datasource_id as number,
      db_name: checkForm.db_name.trim() || undefined,
      table_name: checkForm.table_name.trim(),
    })
    checkReport.value = resp.report
    diffRows.value = parseDiff(resp.report.diff_json)
    const bad = diffRows.value.filter((d) => d.result !== 'match').length
    if (resp.report.pass_flag) {
      ElMessage.success(`核对通过：${diffRows.value.length} 项全部匹配`)
    } else {
      ElMessage.warning(`核对完成，存在 ${bad} 项偏差`)
    }
  } catch (e) {
    ElMessage.error(errText(e, '落标核对失败'))
  } finally {
    checking.value = false
  }
}

// 核对历史
const hisLoading = ref(false)
const hisList = ref<StdCheckReport[]>([])
const hisTotal = ref(0)
const hisQuery = reactive({ model_id: undefined as number | undefined, page: 1, size: 10 })

async function fetchCheckHistory() {
  hisLoading.value = true
  try {
    const resp = await stdCheckList({
      ...hisQuery,
      model_id: hisQuery.model_id || undefined,
    })
    hisList.value = resp.list || []
    hisTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取核对历史失败'))
  } finally {
    hisLoading.value = false
  }
}

function hisSearch() {
  hisQuery.page = 1
  fetchCheckHistory()
}

/** 点历史记录回看当时差异 */
function viewHistoryDiff(row: StdCheckReport) {
  checkReport.value = row
  diffRows.value = parseDiff(row.diff_json)
  ElMessage.info(`已载入 #${row.id ?? '-'} 的核对报告`)
}

onMounted(() => {
  fetchWordRoots()
  fetchDsOptions()
  fetchDictCodeOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ===== 词根 ===== -->
        <el-tab-pane label="词根" name="wordroot">
          <div class="toolbar">
            <el-input
              v-model="wrQuery.keyword"
              placeholder="按词根名称搜索"
              clearable
              style="width: 200px"
              @keyup.enter="wrSearch"
              @clear="wrSearch"
            />
            <el-button type="primary" plain @click="wrSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openWrCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建词根
            </el-button>
            <el-button @click="fetchWordRoots">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="wrLoading" :data="wrList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="词根" min-width="120" show-overflow-tooltip />
            <el-table-column prop="cn_name" label="中文名" min-width="140" show-overflow-tooltip />
            <el-table-column label="类型" width="120">
              <template #default="{ row }">
                <el-tag size="small">
                  {{ typeOptions.find((t) => t.value === row.type)?.label || row.type || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="comment" label="备注" min-width="180" show-overflow-tooltip />
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openWrEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteWr(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="wrQuery.page"
            v-model:page-size="wrQuery.size"
            :total="wrTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchWordRoots"
            @size-change="wrSearch"
          />
        </el-tab-pane>

        <!-- ===== 标准字典 ===== -->
        <el-tab-pane label="标准字典" name="dict">
          <div class="toolbar">
            <el-input
              v-model="dictQuery.keyword"
              placeholder="按字典编码/名称搜索"
              clearable
              style="width: 200px"
              @keyup.enter="dictSearch"
              @clear="dictSearch"
            />
            <el-button type="primary" plain @click="dictSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openDictCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建字典
            </el-button>
            <el-button @click="fetchDicts">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="dictLoading" :data="dictListData" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="dict_code" label="字典编码" min-width="140" show-overflow-tooltip />
            <el-table-column prop="dict_name" label="字典名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="comment" label="备注" min-width="160" show-overflow-tooltip />
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDictDetail(row)">详情</el-button>
                <el-button link type="primary" @click="openDictEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteDict(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="dictQuery.page"
            v-model:page-size="dictQuery.size"
            :total="dictTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchDicts"
            @size-change="dictSearch"
          />
        </el-tab-pane>

        <!-- ===== 数据元 ===== -->
        <el-tab-pane label="数据元" name="element">
          <div class="toolbar">
            <el-input
              v-model="elQuery.keyword"
              placeholder="按数据元编码/名称搜索"
              clearable
              style="width: 200px"
              @keyup.enter="elSearch"
              @clear="elSearch"
            />
            <el-button type="primary" plain @click="elSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openElCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建数据元
            </el-button>
            <el-button @click="fetchElements">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="elLoading" :data="elList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="element_code" label="数据元编码" min-width="140" show-overflow-tooltip />
            <el-table-column prop="element_name" label="数据元名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="数据类型" width="110">
              <template #default="{ row }">
                <el-tag size="small" type="warning">{{ row.data_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="data_length" label="长度" width="80">
              <template #default="{ row }">{{ row.data_length || '-' }}</template>
            </el-table-column>
            <el-table-column prop="data_precision" label="小数位" width="80">
              <template #default="{ row }">{{ row.data_precision ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="dict_code" label="值域字典" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ row.dict_code || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openElEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteEl(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="elQuery.page"
            v-model:page-size="elQuery.size"
            :total="elTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchElements"
            @size-change="elSearch"
          />
        </el-tab-pane>

        <!-- ===== 标准模型 ===== -->
        <el-tab-pane label="标准模型" name="model">
          <div class="toolbar">
            <el-select
              v-model="mdlQuery.status"
              placeholder="全部状态"
              clearable
              style="width: 130px"
              @change="mdlSearch"
            >
              <el-option label="草稿" value="draft" />
              <el-option label="已发布" value="published" />
            </el-select>
            <el-input
              v-model="mdlQuery.keyword"
              placeholder="按模型编码/名称搜索"
              clearable
              style="width: 200px"
              @keyup.enter="mdlSearch"
              @clear="mdlSearch"
            />
            <el-button type="primary" plain @click="mdlSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openMdlCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建模型
            </el-button>
            <el-button @click="fetchModels">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="mdlLoading" :data="mdlList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="model_code" label="模型编码" min-width="140" show-overflow-tooltip />
            <el-table-column prop="model_name" label="模型名称" min-width="160" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="(statusTag[row.status] || statusTag.draft).type" size="small">
                  {{ (statusTag[row.status] || { label: row.status || '-' }).label }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="版本" width="90">
              <template #default="{ row }">{{ row.version ? `v${row.version}` : '-' }}</template>
            </el-table-column>
            <el-table-column label="发布时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.published_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="230" fixed="right">
              <template #default="{ row }">
                <el-button link type="success" @click="onPublishMdl(row)">发布</el-button>
                <el-button link type="primary" @click="openSnapshot(row)">快照</el-button>
                <el-button link type="primary" @click="openMdlEdit(row)">编辑</el-button>
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

        <!-- ===== 落标核对 ===== -->
        <el-tab-pane label="落标核对" name="check">
          <el-form
            ref="checkFormRef"
            :model="checkForm"
            :rules="checkRules"
            inline
            class="check-form"
          >
            <el-form-item label="标准模型" prop="model_id">
              <el-select
                v-model="checkForm.model_id"
                placeholder="仅已发布模型"
                filterable
                style="width: 220px"
              >
                <el-option
                  v-for="m in publishedModels"
                  :key="m.id"
                  :label="`${m.model_name}（${m.model_code}）`"
                  :value="m.id as number"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="数据源" prop="datasource_id">
              <el-select
                v-model="checkForm.datasource_id"
                placeholder="MySQL 数据源"
                filterable
                style="width: 200px"
              >
                <el-option
                  v-for="d in dsOptions"
                  :key="d.id"
                  :label="`${d.name}（${d.type}）`"
                  :value="d.id as number"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="库名">
              <el-input
                v-model="checkForm.db_name"
                placeholder="留空取数据源默认库"
                style="width: 180px"
              />
            </el-form-item>
            <el-form-item label="表名" prop="table_name">
              <el-input
                v-model="checkForm.table_name"
                placeholder="如 dwd_order"
                style="width: 180px"
                @keyup.enter="onRunCheck"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="checking" @click="onRunCheck">
                <el-icon><CircleCheck /></el-icon>&nbsp;执行核对
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 差异结果 -->
          <template v-if="checkReport">
            <el-alert
              :title="
                checkReport.pass_flag
                  ? `核对通过：表 ${checkReport.table_name || '-'} 共 ${diffRows.length} 项全部匹配`
                  : `核对存在偏差：表 ${checkReport.table_name || '-'} 共 ${diffRows.length} 项，其中 ${diffRows.filter((d) => d.result !== 'match').length} 项不一致`
              "
              :type="checkReport.pass_flag ? 'success' : 'warning'"
              :closable="false"
              show-icon
              class="diff-alert"
            />
            <el-table :data="diffRows" stripe border size="small" class="diff-table">
              <el-table-column prop="column" label="列（数据元）" min-width="150" show-overflow-tooltip />
              <el-table-column prop="expect" label="期望类型" min-width="140" show-overflow-tooltip />
              <el-table-column label="实际类型" min-width="140" show-overflow-tooltip>
                <template #default="{ row }">{{ row.actual || '（缺列）' }}</template>
              </el-table-column>
              <el-table-column label="结果" width="110">
                <template #default="{ row }">
                  <el-tag :type="resultOf(row.result).type" size="small">
                    {{ resultOf(row.result).label }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </template>
          <el-empty v-else description="选择模型与目标表后执行核对，差异将在此展示" :image-size="70" />

          <!-- 核对历史 -->
          <div class="section-title">
            <span>核对历史</span>
            <el-select
              v-model="hisQuery.model_id"
              placeholder="全部模型"
              clearable
              filterable
              style="width: 220px"
              @change="hisSearch"
            >
              <el-option
                v-for="m in publishedModels"
                :key="m.id"
                :label="m.model_name"
                :value="m.id as number"
              />
            </el-select>
            <el-button @click="fetchCheckHistory">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="hisLoading" :data="hisList" stripe border size="small">
            <el-table-column prop="id" label="报告ID" width="80" />
            <el-table-column prop="model_id" label="模型ID" width="90" />
            <el-table-column prop="table_name" label="核对表" min-width="160" show-overflow-tooltip />
            <el-table-column label="结论" width="100">
              <template #default="{ row }">
                <el-tag :type="row.pass_flag ? 'success' : 'warning'" size="small">
                  {{ row.pass_flag ? '通过' : '有偏差' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="核对时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.run_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewHistoryDiff(row)">查看差异</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="hisQuery.page"
            v-model:page-size="hisQuery.size"
            :total="hisTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchCheckHistory"
            @size-change="hisSearch"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 词根新建/编辑 -->
    <el-dialog
      v-model="wrDialogVisible"
      :title="wrEditingId === null ? '新建词根' : '编辑词根'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="wrFormRef" :model="wrForm" :rules="wrRules" label-width="90px">
        <el-form-item label="词根名称" prop="name">
          <el-input v-model="wrForm.name" placeholder="如：user" maxlength="64" />
        </el-form-item>
        <el-form-item label="中文名">
          <el-input v-model="wrForm.cn_name" placeholder="如：用户" maxlength="64" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="wrForm.type" style="width: 100%">
            <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="wrForm.comment" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="wrDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="wrSaving" @click="onSaveWr">保存</el-button>
      </template>
    </el-dialog>

    <!-- 字典新建/编辑（含动态字典项行编辑） -->
    <el-dialog
      v-model="dictDialogVisible"
      :title="dictEditingId === null ? '新建字典' : '编辑字典'"
      width="680px"
      destroy-on-close
    >
      <el-form ref="dictFormRef" :model="dictForm" :rules="dictRules" label-width="90px">
        <el-form-item label="字典编码" prop="dict_code">
          <el-input v-model="dictForm.dict_code" placeholder="如：gender" maxlength="64" />
        </el-form-item>
        <el-form-item label="字典名称" prop="dict_name">
          <el-input v-model="dictForm.dict_name" placeholder="如：性别字典" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="dictForm.comment" maxlength="255" />
        </el-form-item>
        <el-form-item label="字典项">
          <div class="items-editor">
            <div v-for="(it, idx) in dictForm.items" :key="idx" class="item-row">
              <el-input v-model="it.item_code" placeholder="item_code（编码）" style="width: 180px" />
              <el-input v-model="it.item_value" placeholder="item_value（值）" style="width: 180px" />
              <el-input-number v-model="it.sort" :min="0" controls-position="right" style="width: 110px" />
              <el-button link type="danger" @click="removeDictItem(idx)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="primary" plain size="small" @click="addDictItem">
              <el-icon><Plus /></el-icon>&nbsp;添加字典项
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dictDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="dictSaving" @click="onSaveDict">保存</el-button>
      </template>
    </el-dialog>

    <!-- 字典详情 -->
    <el-dialog v-model="dictDetailVisible" title="字典详情" width="560px">
      <div v-loading="dictDetailLoading">
        <template v-if="dictDetail">
          <el-descriptions :column="2" border size="small" class="detail-desc">
            <el-descriptions-item label="ID">{{ dictDetail.id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ dictDetail.status === 1 ? '启用' : '停用' }}
            </el-descriptions-item>
            <el-descriptions-item label="字典编码">{{ dictDetail.dict_code || '-' }}</el-descriptions-item>
            <el-descriptions-item label="字典名称">{{ dictDetail.dict_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">
              {{ dictDetail.comment || '-' }}
            </el-descriptions-item>
          </el-descriptions>
          <el-table :data="dictDetail.items || []" stripe border size="small" class="detail-table">
            <el-table-column prop="item_code" label="编码" min-width="120" />
            <el-table-column prop="item_value" label="值" min-width="120" />
            <el-table-column prop="sort" label="排序" width="80" />
          </el-table>
        </template>
      </div>
      <template #footer>
        <el-button @click="dictDetailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 数据元新建/编辑 -->
    <el-dialog
      v-model="elDialogVisible"
      :title="elEditingId === null ? '新建数据元' : '编辑数据元'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="elFormRef" :model="elForm" :rules="elRules" label-width="90px">
        <el-form-item label="编码" prop="element_code">
          <el-input v-model="elForm.element_code" placeholder="如：user_name" maxlength="64" />
        </el-form-item>
        <el-form-item label="名称" prop="element_name">
          <el-input v-model="elForm.element_name" placeholder="如：用户名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="数据类型" prop="data_type">
          <el-select v-model="elForm.data_type" style="width: 100%">
            <el-option v-for="t in dataTypeOptions" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="长度">
          <el-input-number v-model="elForm.data_length" :min="0" controls-position="right" />
          <span class="form-tip">varchar/decimal 必填</span>
        </el-form-item>
        <el-form-item label="小数位">
          <el-input-number v-model="elForm.data_precision" :min="0" controls-position="right" />
          <span class="form-tip">decimal 使用</span>
        </el-form-item>
        <el-form-item label="值域字典">
          <el-select
            v-model="elForm.dict_code"
            placeholder="选择字典（可留空）"
            clearable
            filterable
            allow-create
            style="width: 100%"
          >
            <el-option v-for="c in dictCodeOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="elForm.comment" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="elDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="elSaving" @click="onSaveEl">保存</el-button>
      </template>
    </el-dialog>

    <!-- 标准模型新建/编辑 -->
    <el-dialog
      v-model="mdlDialogVisible"
      :title="mdlEditingId === null ? '新建标准模型' : '编辑标准模型'"
      width="620px"
      destroy-on-close
    >
      <el-form ref="mdlFormRef" :model="mdlForm" :rules="mdlRules" label-width="100px">
        <el-form-item label="模型编码" prop="model_code">
          <el-input v-model="mdlForm.model_code" placeholder="如：dwd_order_std" maxlength="64" />
        </el-form-item>
        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="mdlForm.model_name" placeholder="如：订单标准模型" maxlength="64" />
        </el-form-item>
        <el-form-item label="组成数据元">
          <el-input
            v-model="mdlForm.elements_json"
            type="textarea"
            :rows="8"
            placeholder='[{"element_code":"order_id","required":true,"pk":true},{"element_code":"user_name","required":true}]'
          />
          <div class="form-tip">
            JSON 数组，每项 element_code 必填（引用已有数据元），required/pk 可选
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mdlDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mdlSaving" @click="onSaveMdl">保存</el-button>
      </template>
    </el-dialog>

    <!-- 发布快照 -->
    <el-dialog v-model="snapVisible" :title="`发布快照：${snapModel}`" width="640px">
      <div v-loading="snapLoading">
        <pre v-if="snapJson" class="json-pre">{{ prettyJson(snapJson) }}</pre>
        <el-empty v-else description="尚未发布，暂无快照" :image-size="70" />
      </div>
      <template #footer>
        <el-button @click="snapVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
.items-editor {
  width: 100%;
}
.item-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.form-tip {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
.detail-desc {
  margin-bottom: 12px;
}
.json-pre {
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
.check-form {
  margin-bottom: 8px;
}
.diff-alert {
  margin-bottom: 10px;
}
.diff-table {
  margin-bottom: 18px;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 6px 0 12px;
  font-weight: 600;
}
</style>
