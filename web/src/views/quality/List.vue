<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  createQualityRule,
  deleteQualityRule,
  qualityRuleList,
  createQualityJob,
  qualityJobList,
  runQualityJob,
  qualityResultList,
  type QualityRule,
  type QualityJob,
  type QualityResult,
  type RunJobResp,
} from '@/api/quality'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

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

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
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

/** 结果页 job 下拉候选 */
const jobOptions = ref<QualityJob[]>([])
async function fetchJobOptions() {
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

// ---------- 页签切换：懒加载对应数据 ----------
const activeTab = ref('rule')
function onTabChange() {
  if (activeTab.value === 'rule' && ruleTotal.value === 0 && ruleList.value.length === 0) {
    fetchRules()
  } else if (activeTab.value === 'job' && jobList.value.length === 0) {
    fetchJobs()
  } else if (activeTab.value === 'result' && resultList.value.length === 0) {
    fetchJobOptions()
    fetchResults()
  }
}

onMounted(() => {
  fetchRules()
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
.run-rate {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
