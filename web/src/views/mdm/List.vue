<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  saveMdmModel,
  deleteMdmModel,
  mdmModelList,
  mdmNextCode,
  saveMdmRecord,
  publishMdmRecord,
  deleteMdmRecord,
  mdmRecordList,
  distributeMdmRecord,
  mdmDistList,
  type MdmModelInfo,
  type MdmRecordInfo,
  type MdmDistribution,
} from '@/api/mdm'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共 ----------
const activeTab = ref('model')

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

function prettyJson(raw?: string) {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 200 })
    dsOptions.value = resp.list || []
  } catch {
    /* 下拉数据拉取失败不阻断页面 */
  }
}

function onTabChange(name: string | number) {
  const tab = String(name)
  if (tab === 'model') fetchModels()
  else if (tab === 'record') {
    fetchMdlOptions()
    fetchRecords()
  } else if (tab === 'dist') {
    fetchMdlOptions()
    fetchRecOptions()
    fetchDists()
  }
}

// ---------- 模型 ----------
const mdlLoading = ref(false)
const mdlList = ref<MdmModelInfo[]>([])
const mdlTotal = ref(0)
const mdlQuery = reactive({ keyword: '', page: 1, size: 10 })

const codeStrategyOptions = [
  { value: 'segment', label: '号段模式（segment）' },
  { value: 'auto', label: '自增（auto）' },
]

const fieldTypeOptions = ['string', 'int', 'decimal', 'bool', 'date']

async function fetchModels() {
  mdlLoading.value = true
  try {
    const resp = await mdmModelList({ ...mdlQuery, keyword: mdlQuery.keyword || undefined })
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

/** 模型字段行（fields_json 动态行编辑） */
interface FieldRow {
  name: string
  type: string
  length?: number
  required?: boolean
  comment?: string
}

interface MdlForm {
  model_code: string
  model_name: string
  table_name: string
  fields: FieldRow[]
  code_strategy: string
}

const mdlDialogVisible = ref(false)
const mdlSaving = ref(false)
const mdlEditingId = ref<number | null>(null)
const mdlFormRef = ref<FormInstance>()
const mdlForm = reactive<MdlForm>({
  model_code: '',
  model_name: '',
  table_name: '',
  fields: [],
  code_strategy: 'segment',
})
const mdlRules: FormRules = {
  model_code: [{ required: true, message: '请输入模型编码', trigger: 'blur' }],
  model_name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
}

function addField() {
  mdlForm.fields.push({ name: '', type: 'string', length: undefined, required: false, comment: '' })
}

function removeField(idx: number) {
  mdlForm.fields.splice(idx, 1)
}

/** 从已存 fields_json 尽力还原行编辑数据 */
function parseFields(raw?: string): FieldRow[] {
  if (!raw) return []
  try {
    const v = JSON.parse(raw)
    if (!Array.isArray(v)) return []
    return v.map((it) => ({
      name: String(it?.name ?? ''),
      type: String(it?.type ?? 'string'),
      length: it?.length ?? undefined,
      required: !!it?.required,
      comment: it?.comment ?? '',
    }))
  } catch {
    return []
  }
}

function openMdlCreate() {
  mdlEditingId.value = null
  Object.assign(mdlForm, {
    model_code: '',
    model_name: '',
    table_name: '',
    fields: [],
    code_strategy: 'segment',
  })
  addField()
  mdlDialogVisible.value = true
}

function openMdlEdit(row: MdmModelInfo) {
  mdlEditingId.value = row.id ?? null
  Object.assign(mdlForm, {
    model_code: row.model_code || '',
    model_name: row.model_name || '',
    table_name: row.table_name || '',
    fields: parseFields(row.fields_json),
    code_strategy: row.code_strategy || 'segment',
  })
  if (mdlForm.fields.length === 0) addField()
  mdlDialogVisible.value = true
}

async function onSaveMdl() {
  const valid = await mdlFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const fields = mdlForm.fields
  if (fields.length === 0) {
    ElMessage.warning('请至少定义一个字段')
    return
  }
  const seen = new Set<string>()
  for (let i = 0; i < fields.length; i++) {
    const name = (fields[i].name || '').trim()
    if (!name) {
      ElMessage.warning(`第 ${i + 1} 个字段的 name 不能为空`)
      return
    }
    if (seen.has(name)) {
      ElMessage.warning(`字段 name 重复：${name}`)
      return
    }
    seen.add(name)
  }
  mdlSaving.value = true
  try {
    await saveMdmModel({
      id: mdlEditingId.value ?? undefined,
      model_code: mdlForm.model_code.trim(),
      model_name: mdlForm.model_name.trim(),
      table_name: mdlForm.table_name.trim() || undefined,
      fields_json: JSON.stringify(
        fields.map((f) => ({
          name: f.name.trim(),
          type: f.type,
          length: f.length || undefined,
          required: !!f.required,
          comment: f.comment?.trim() || undefined,
        })),
      ),
      code_strategy: mdlForm.code_strategy || undefined,
    })
    ElMessage.success(mdlEditingId.value === null ? '模型创建成功' : '模型更新成功')
    mdlDialogVisible.value = false
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '保存模型失败'))
  } finally {
    mdlSaving.value = false
  }
}

async function onDeleteMdl(row: MdmModelInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除模型「${row.model_name || row.model_code}」吗？其下记录将无法再管理。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteMdmModel(row.id)
    ElMessage.success('删除成功')
    fetchModels()
  } catch (e) {
    ElMessage.error(errText(e, '删除模型失败'))
  }
}

// ---------- 记录 ----------
const recLoading = ref(false)
const recList = ref<MdmRecordInfo[]>([])
const recTotal = ref(0)
const recQuery = reactive({ model_id: undefined as number | undefined, keyword: '', page: 1, size: 10 })

const statusTag: Record<string, { label: string; type: 'info' | 'success' }> = {
  draft: { label: '草稿', type: 'info' },
  published: { label: '已发布', type: 'success' },
}

async function fetchRecords() {
  recLoading.value = true
  try {
    const resp = await mdmRecordList({
      ...recQuery,
      model_id: recQuery.model_id || undefined,
      keyword: recQuery.keyword || undefined,
    })
    recList.value = resp.list || []
    recTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取记录列表失败'))
  } finally {
    recLoading.value = false
  }
}

function recSearch() {
  recQuery.page = 1
  fetchRecords()
}

interface RecForm {
  model_id: number | undefined
  biz_code: string
  data_json: string
}

const recDialogVisible = ref(false)
const recSaving = ref(false)
const recEditingId = ref<number | null>(null)
const recFormRef = ref<FormInstance>()
const recForm = reactive<RecForm>({ model_id: undefined, biz_code: '', data_json: '' })
const recRules: FormRules = {
  model_id: [{ required: true, message: '请选择所属模型', trigger: 'change' }],
  data_json: [{ required: true, message: '请填写记录数据 JSON', trigger: 'blur' }],
}

/** 模型下拉选项（记录新建/筛选共用），与分页列表解耦 */
const mdlOptions = ref<MdmModelInfo[]>([])
async function fetchMdlOptions() {
  try {
    const resp = await mdmModelList({ page: 1, size: 200 })
    mdlOptions.value = resp.list || []
  } catch {
    /* 忽略 */
  }
}

/** 选中模型后展示字段提示 */
const recModelHint = ref('')
function mdlLabel(m?: MdmModelInfo) {
  return m ? `${m.model_name}（${m.model_code}）` : ''
}

function onRecModelChange(id?: number) {
  const m = mdlOptions.value.find((x) => x.id === id)
  if (m?.fields_json) {
    const names = parseFields(m.fields_json)
      .map((f) => f.name)
      .filter(Boolean)
    recModelHint.value = `模型字段：${names.join('、') || '（未定义）'}`
  } else {
    recModelHint.value = ''
  }
}

function openRecCreate() {
  recEditingId.value = null
  Object.assign(recForm, { model_id: recQuery.model_id || undefined, biz_code: '', data_json: '' })
  recModelHint.value = ''
  onRecModelChange(recForm.model_id)
  recDialogVisible.value = true
}

async function openRecEdit(row: MdmRecordInfo) {
  if (row.id === undefined) return
  recEditingId.value = row.id
  Object.assign(recForm, {
    model_id: row.model_id || undefined,
    biz_code: row.biz_code || '',
    data_json: prettyJson(row.data_json) || row.data_json || '',
  })
  recModelHint.value = ''
  onRecModelChange(recForm.model_id)
  recDialogVisible.value = true
}

async function onFetchBizCode() {
  if (!recForm.model_id) {
    ElMessage.warning('请先选择所属模型')
    return
  }
  try {
    const resp = await mdmNextCode(recForm.model_id)
    recForm.biz_code = resp.biz_code
    ElMessage.success(`已获取业务编码：${resp.biz_code}`)
  } catch (e) {
    ElMessage.error(errText(e, '获取业务编码失败'))
  }
}

async function onSaveRec() {
  const valid = await recFormRef.value?.validate().catch(() => false)
  if (!valid) return
  // data_json 提交前 parse 校验：须为 JSON 对象
  let dataJson = recForm.data_json.trim()
  try {
    const v = JSON.parse(dataJson)
    if (!v || typeof v !== 'object' || Array.isArray(v)) {
      ElMessage.error('data_json 须为 JSON 对象（键值对），不能是数组或标量')
      return
    }
    dataJson = JSON.stringify(v)
  } catch {
    ElMessage.error('data_json 不是合法 JSON，请检查格式')
    return
  }
  recSaving.value = true
  try {
    await saveMdmRecord({
      id: recEditingId.value ?? undefined,
      model_id: recForm.model_id as number,
      biz_code: recForm.biz_code.trim() || undefined,
      data_json: dataJson,
    })
    ElMessage.success(recEditingId.value === null ? '记录创建成功' : '记录更新成功')
    recDialogVisible.value = false
    fetchRecords()
  } catch (e) {
    ElMessage.error(errText(e, '保存记录失败'))
  } finally {
    recSaving.value = false
  }
}

async function onPublishRec(row: MdmRecordInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    row.status === 'published'
      ? `记录「${row.biz_code || row.id}」已发布，重新发布将版本+1。确定继续吗？`
      : `确定发布记录「${row.biz_code || row.id}」吗？发布后进入正式主数据。`,
    '发布确认',
    { confirmButtonText: '发布', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await publishMdmRecord(row.id)
    ElMessage.success(`发布成功，当前版本 v${resp.version ?? '-'}`)
    fetchRecords()
  } catch (e) {
    ElMessage.error(errText(e, '发布失败'))
  }
}

async function onDeleteRec(row: MdmRecordInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除记录「${row.biz_code || row.id}」吗？`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteMdmRecord(row.id)
    ElMessage.success('删除成功')
    fetchRecords()
  } catch (e) {
    ElMessage.error(errText(e, '删除记录失败'))
  }
}

// ---------- 分发 ----------
const distFormRef = ref<FormInstance>()
const distForm = reactive({
  record_id: undefined as number | undefined,
  target_ds_id: undefined as number | undefined,
  target_db: '',
  target_table: '',
})
const distRules: FormRules = {
  record_id: [{ required: true, message: '请选择主数据记录', trigger: 'change' }],
  target_ds_id: [{ required: true, message: '请选择目标数据源', trigger: 'change' }],
  target_table: [{ required: true, message: '请输入目标表名（仅字母/数字/下划线）', trigger: 'blur' }],
}

/** 分发表单的记录下拉（跨模型取一页供选择） */
const recOptions = ref<MdmRecordInfo[]>([])
const recOptionsLoading = ref(false)
async function fetchRecOptions() {
  recOptionsLoading.value = true
  try {
    const resp = await mdmRecordList({ page: 1, size: 200 })
    recOptions.value = resp.list || []
  } catch {
    /* 忽略 */
  } finally {
    recOptionsLoading.value = false
  }
}

const distributing = ref(false)
async function onDistribute() {
  const valid = await distFormRef.value?.validate().catch(() => false)
  if (!valid) return
  distributing.value = true
  try {
    const resp = await distributeMdmRecord({
      record_id: distForm.record_id as number,
      target_ds_id: distForm.target_ds_id as number,
      target_db: distForm.target_db.trim() || undefined,
      target_table: distForm.target_table.trim(),
    })
    const d = resp.distribution
    if (d?.status === 'success') {
      ElMessage.success(`分发成功：已写入 ${d.target_table || distForm.target_table}`)
    } else {
      ElMessage.warning(`分发完成，状态 ${d?.status || '未知'}：${d?.msg || ''}`)
    }
    fetchDists()
  } catch (e) {
    ElMessage.error(errText(e, '分发失败'))
  } finally {
    distributing.value = false
  }
}

const distLoading = ref(false)
const distList = ref<MdmDistribution[]>([])
const distTotal = ref(0)
const distQuery = reactive({ record_id: undefined as number | undefined, page: 1, size: 10 })

async function fetchDists() {
  distLoading.value = true
  try {
    const resp = await mdmDistList({
      ...distQuery,
      record_id: distQuery.record_id || undefined,
    })
    distList.value = resp.list || []
    distTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取分发记录失败'))
  } finally {
    distLoading.value = false
  }
}

function distSearch() {
  distQuery.page = 1
  fetchDists()
}

function distStatusTag(status?: string): 'success' | 'danger' | 'warning' | 'info' {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'pending') return 'warning'
  return 'info'
}

onMounted(() => {
  fetchModels()
  fetchMdlOptions()
  fetchDsOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ===== 模型 ===== -->
        <el-tab-pane label="模型" name="model">
          <div class="toolbar">
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
            <el-table-column prop="model_code" label="模型编码" min-width="130" show-overflow-tooltip />
            <el-table-column prop="model_name" label="模型名称" min-width="150" show-overflow-tooltip />
            <el-table-column prop="table_name" label="物理表" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">{{ row.table_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="字段数" width="80">
              <template #default="{ row }">{{ parseFields(row.fields_json).length }}</template>
            </el-table-column>
            <el-table-column label="编码策略" width="130">
              <template #default="{ row }">
                <el-tag size="small" type="warning">{{ row.code_strategy || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
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

        <!-- ===== 记录 ===== -->
        <el-tab-pane label="记录" name="record">
          <div class="toolbar">
            <el-select
              v-model="recQuery.model_id"
              placeholder="全部模型"
              clearable
              filterable
              style="width: 220px"
              @change="recSearch"
            >
              <el-option
                v-for="m in mdlOptions"
                :key="m.id"
                :label="mdlLabel(m)"
                :value="m.id as number"
              />
            </el-select>
            <el-input
              v-model="recQuery.keyword"
              placeholder="按业务编码搜索"
              clearable
              style="width: 180px"
              @keyup.enter="recSearch"
              @clear="recSearch"
            />
            <el-button type="primary" plain @click="recSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openRecCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建记录
            </el-button>
            <el-button @click="fetchRecords">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="recLoading" :data="recList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="model_id" label="模型ID" width="90" />
            <el-table-column prop="biz_code" label="业务编码" min-width="140" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="(statusTag[row.status] || statusTag.draft).type" size="small">
                  {{ (statusTag[row.status] || { label: row.status || '-' }).label }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="版本" width="80">
              <template #default="{ row }">{{ row.version ? `v${row.version}` : '-' }}</template>
            </el-table-column>
            <el-table-column label="记录数据" min-width="240" show-overflow-tooltip>
              <template #default="{ row }">
                <code class="data-preview">{{ row.data_json }}</code>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="success" @click="onPublishRec(row)">发布</el-button>
                <el-button link type="primary" @click="openRecEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteRec(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="recQuery.page"
            v-model:page-size="recQuery.size"
            :total="recTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchRecords"
            @size-change="recSearch"
          />
        </el-tab-pane>

        <!-- ===== 分发 ===== -->
        <el-tab-pane label="分发" name="dist">
          <el-form ref="distFormRef" :model="distForm" :rules="distRules" inline class="dist-form">
            <el-form-item label="主数据记录" prop="record_id">
              <el-select
                v-model="distForm.record_id"
                placeholder="选择要分发的记录"
                filterable
                v-loading="recOptionsLoading"
                style="width: 260px"
              >
                <el-option
                  v-for="r in recOptions"
                  :key="r.id"
                  :label="`#${r.id} ${r.biz_code || '（无编码）'}`"
                  :value="r.id as number"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="目标数据源" prop="target_ds_id">
              <el-select
                v-model="distForm.target_ds_id"
                placeholder="选择数据源"
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
            <el-form-item label="目标库">
              <el-input
                v-model="distForm.target_db"
                placeholder="留空取数据源默认库"
                style="width: 170px"
              />
            </el-form-item>
            <el-form-item label="目标表" prop="target_table">
              <el-input v-model="distForm.target_table" placeholder="如 mdm_user" style="width: 170px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="distributing" @click="onDistribute">
                <el-icon><Position /></el-icon>&nbsp;执行分发
              </el-button>
            </el-form-item>
          </el-form>

          <el-alert
            title="目标表不存在时将按模型 fields_json 自动建表；重复分发按业务编码幂等覆盖"
            type="info"
            :closable="false"
            show-icon
            class="dist-tip"
          />

          <div class="toolbar">
            <el-select
              v-model="distQuery.record_id"
              placeholder="全部记录"
              clearable
              filterable
              style="width: 260px"
              @change="distSearch"
            >
              <el-option
                v-for="r in recOptions"
                :key="r.id"
                :label="`#${r.id} ${r.biz_code || '（无编码）'}`"
                :value="r.id as number"
              />
            </el-select>
            <el-button @click="fetchDists">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="distLoading" :data="distList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="record_id" label="记录ID" width="90" />
            <el-table-column prop="target_ds_id" label="目标数据源" width="110" />
            <el-table-column label="目标库表" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                {{ [row.target_db, row.target_table].filter(Boolean).join('.') || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="distStatusTag(row.status)" size="small">{{ row.status || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="msg" label="信息" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ row.msg || '-' }}</template>
            </el-table-column>
            <el-table-column label="分发时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.distributed_at) }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="distQuery.page"
            v-model:page-size="distQuery.size"
            :total="distTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchDists"
            @size-change="distSearch"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 模型新建/编辑（字段动态行编辑） -->
    <el-dialog
      v-model="mdlDialogVisible"
      :title="mdlEditingId === null ? '新建主数据模型' : '编辑主数据模型'"
      width="720px"
      destroy-on-close
    >
      <el-form ref="mdlFormRef" :model="mdlForm" :rules="mdlRules" label-width="90px">
        <el-form-item label="模型编码" prop="model_code">
          <el-input v-model="mdlForm.model_code" placeholder="如：mdm_user" maxlength="64" />
        </el-form-item>
        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="mdlForm.model_name" placeholder="如：用户主数据" maxlength="64" />
        </el-form-item>
        <el-form-item label="物理表">
          <el-input v-model="mdlForm.table_name" placeholder="留空按模型编码落表（仅字母/数字/下划线）" maxlength="64" />
        </el-form-item>
        <el-form-item label="编码策略">
          <el-select v-model="mdlForm.code_strategy" style="width: 100%">
            <el-option
              v-for="s in codeStrategyOptions"
              :key="s.value"
              :label="s.label"
              :value="s.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="字段定义">
          <div class="fields-editor">
            <div v-for="(f, idx) in mdlForm.fields" :key="idx" class="field-row">
              <el-input v-model="f.name" placeholder="name 字段名" style="width: 150px" />
              <el-select v-model="f.type" style="width: 110px">
                <el-option v-for="t in fieldTypeOptions" :key="t" :label="t" :value="t" />
              </el-select>
              <el-input-number
                v-model="f.length"
                :min="0"
                :disabled="f.type !== 'string' && f.type !== 'decimal'"
                placeholder="长度"
                controls-position="right"
                style="width: 110px"
              />
              <el-checkbox v-model="f.required">必填</el-checkbox>
              <el-input v-model="f.comment" placeholder="备注" style="width: 130px" />
              <el-button link type="danger" @click="removeField(idx)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="primary" plain size="small" @click="addField">
              <el-icon><Plus /></el-icon>&nbsp;添加字段
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mdlDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mdlSaving" @click="onSaveMdl">保存</el-button>
      </template>
    </el-dialog>

    <!-- 记录新建/编辑 -->
    <el-dialog
      v-model="recDialogVisible"
      :title="recEditingId === null ? '新建主数据记录' : '编辑主数据记录'"
      width="640px"
      destroy-on-close
    >
      <el-form ref="recFormRef" :model="recForm" :rules="recRules" label-width="90px">
        <el-form-item label="所属模型" prop="model_id">
          <el-select
            v-model="recForm.model_id"
            placeholder="选择模型"
            filterable
            :disabled="recEditingId !== null"
            style="width: 100%"
            @change="onRecModelChange"
          >
            <el-option
              v-for="m in mdlOptions"
              :key="m.id"
              :label="mdlLabel(m)"
              :value="m.id as number"
            />
          </el-select>
          <div v-if="recModelHint" class="form-tip">{{ recModelHint }}</div>
        </el-form-item>
        <el-form-item label="业务编码">
          <div class="bizcode-row">
            <el-input
              v-model="recForm.biz_code"
              placeholder="留空则由编码策略生成"
              maxlength="64"
            />
            <el-button @click="onFetchBizCode">自动取码</el-button>
          </div>
        </el-form-item>
        <el-form-item label="记录数据" prop="data_json">
          <el-input
            v-model="recForm.data_json"
            type="textarea"
            :rows="9"
            placeholder='按模型字段填写 JSON 对象，如：{"name":"张三","age":30}'
          />
          <div class="form-tip">JSON 对象，字段须与模型 fields_json 定义一致（多余字段会被拒绝）</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="recSaving" @click="onSaveRec">保存</el-button>
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
.fields-editor {
  width: 100%;
}
.field-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.data-preview {
  font-size: 12px;
  color: #606266;
  word-break: break-all;
}
.bizcode-row {
  display: flex;
  gap: 8px;
  width: 100%;
}
.form-tip {
  width: 100%;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
.dist-form {
  margin-bottom: 8px;
}
.dist-tip {
  margin-bottom: 14px;
}
</style>
