<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'OpsOps' })
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  createArchiveTask,
  archiveTaskList,
  deleteArchiveTask,
  runArchive,
  archiveLogList,
  loginLogList,
  auditLogList,
  type OpsArchiveTask,
  type OpsArchiveLog,
  type OpsLoginLog,
  type OpsAuditLog,
} from '@/api/ops'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 公共 ----------
const activeTab = ref('archive')

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

function dsName(id?: number) {
  return dsOptions.value.find((d) => d.id === id)?.name || (id ? `#${id}` : '-')
}

function onTabChange(name: string | number) {
  const tab = String(name)
  if (tab === 'archive') fetchTasks()
  else if (tab === 'login') fetchLoginLogs()
  else if (tab === 'audit') fetchAuditLogs()
}

// ---------- 归档任务 ----------
const taskLoading = ref(false)
const taskList = ref<OpsArchiveTask[]>([])
const taskTotal = ref(0)
const taskQuery = reactive({ page: 1, size: 10 })

const lastStatusTag: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
  success: 'success',
  failed: 'danger',
  running: 'warning',
}

async function fetchTasks() {
  taskLoading.value = true
  try {
    const resp = await archiveTaskList({ ...taskQuery })
    taskList.value = resp.list || []
    taskTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取归档任务列表失败'))
  } finally {
    taskLoading.value = false
  }
}

// 新建对话框
const dialogVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  datasource_id: undefined as number | undefined,
  db_name: '',
  table_name: '',
  condition_sql: '',
  archive_table: '',
  delete_source: false,
  batch_size: 1000,
  cron: '',
})
const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  table_name: [{ required: true, message: '请输入源表名（仅字母/数字/下划线）', trigger: 'blur' }],
  archive_table: [{ required: true, message: '请输入归档表名', trigger: 'blur' }],
}

function openCreate() {
  Object.assign(form, {
    name: '',
    datasource_id: undefined,
    db_name: '',
    table_name: '',
    condition_sql: '',
    archive_table: '',
    delete_source: false,
    batch_size: 1000,
    cron: '',
  })
  dialogVisible.value = true
}

async function onSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await createArchiveTask({
      name: form.name.trim(),
      datasource_id: form.datasource_id as number,
      db_name: form.db_name.trim() || undefined,
      table_name: form.table_name.trim(),
      condition_sql: form.condition_sql.trim() || undefined,
      archive_table: form.archive_table.trim(),
      delete_source: form.delete_source,
      batch_size: form.batch_size || undefined,
      cron: form.cron.trim() || undefined,
    })
    ElMessage.success('归档任务创建成功')
    dialogVisible.value = false
    fetchTasks()
  } catch (e) {
    ElMessage.error(errText(e, '创建归档任务失败'))
  } finally {
    saving.value = false
  }
}

async function onRun(row: OpsArchiveTask) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定立即执行归档任务「${row.name}」吗？${
      row.delete_source ? '该任务将删除源表已归档数据。' : ''
    }`,
    '执行确认',
    { confirmButtonText: '执行', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await runArchive(row.id)
    ElMessageBox.alert(
      `归档行数：${resp.archived_rows}\n删除行数：${resp.deleted_rows}\n执行状态：${resp.status}`,
      `执行结果：${row.name}`,
      { confirmButtonText: '知道了', type: resp.status === 'success' ? 'success' : 'warning' },
    ).catch(() => {})
    fetchTasks()
  } catch (e) {
    ElMessage.error(errText(e, '执行归档任务失败'))
  }
}

async function onDeleteTask(row: OpsArchiveTask) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除归档任务「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteArchiveTask(row.id)
    ElMessage.success('删除成功')
    fetchTasks()
  } catch (e) {
    ElMessage.error(errText(e, '删除归档任务失败'))
  }
}

// 归档日志弹窗
const logDialogVisible = ref(false)
const logLoading = ref(false)
const logList = ref<OpsArchiveLog[]>([])
const logTotal = ref(0)
const logQuery = reactive({ task_id: 0, task_name: '', page: 1, size: 10 })

async function fetchLogs() {
  logLoading.value = true
  try {
    const resp = await archiveLogList({
      ...logQuery,
      task_id: logQuery.task_id || undefined,
    })
    logList.value = resp.list || []
    logTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取归档日志失败'))
  } finally {
    logLoading.value = false
  }
}

function openLogs(row: OpsArchiveTask) {
  logQuery.task_id = row.id ?? 0
  logQuery.task_name = row.name || `#${row.id}`
  logQuery.page = 1
  logDialogVisible.value = true
  fetchLogs()
}

// ---------- 登录日志 ----------
const loginLoading = ref(false)
const loginList = ref<OpsLoginLog[]>([])
const loginTotal = ref(0)
const loginQuery = reactive({
  username: '',
  success_filter: undefined as number | undefined,
  page: 1,
  size: 10,
})

const loginFilterOptions = [
  { value: 1, label: '成功' },
  { value: 2, label: '失败' },
]

async function fetchLoginLogs() {
  loginLoading.value = true
  try {
    const resp = await loginLogList({
      ...loginQuery,
      username: loginQuery.username || undefined,
      success_filter: loginQuery.success_filter || undefined,
    })
    loginList.value = resp.list || []
    loginTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取登录日志失败'))
  } finally {
    loginLoading.value = false
  }
}

function loginSearch() {
  loginQuery.page = 1
  fetchLoginLogs()
}

// ---------- 审计日志 ----------
const auditLoading = ref(false)
const auditList = ref<OpsAuditLog[]>([])
const auditTotal = ref(0)
const auditQuery = reactive({ resource: '', uid: undefined as number | undefined, page: 1, size: 10 })

async function fetchAuditLogs() {
  auditLoading.value = true
  try {
    const resp = await auditLogList({
      ...auditQuery,
      resource: auditQuery.resource || undefined,
      uid: auditQuery.uid || undefined,
    })
    auditList.value = resp.list || []
    auditTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取审计日志失败'))
  } finally {
    auditLoading.value = false
  }
}

function auditSearch() {
  auditQuery.page = 1
  fetchAuditLogs()
}

onMounted(() => {
  fetchTasks()
  fetchDsOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ===== 归档任务 ===== -->
        <el-tab-pane label="归档任务" name="archive">
          <div class="toolbar">
            <el-button type="primary" @click="openCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建归档任务
            </el-button>
            <el-button @click="fetchTasks">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="taskLoading" :data="taskList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="任务名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="数据源" min-width="130" show-overflow-tooltip>
              <template #default="{ row }">{{ dsName(row.datasource_id) }}</template>
            </el-table-column>
            <el-table-column label="源表" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">
                {{ [row.db_name, row.table_name].filter(Boolean).join('.') || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="archive_table" label="归档表" min-width="140" show-overflow-tooltip />
            <el-table-column label="删源数据" width="90">
              <template #default="{ row }">
                <el-tag :type="row.delete_source ? 'danger' : 'info'" size="small">
                  {{ row.delete_source ? '删除' : '保留' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="batch_size" label="批大小" width="80">
              <template #default="{ row }">{{ row.batch_size || '-' }}</template>
            </el-table-column>
            <el-table-column prop="cron" label="调度周期" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ row.cron || '-' }}</template>
            </el-table-column>
            <el-table-column label="最近状态" width="100">
              <template #default="{ row }">
                <el-tag
                  :type="lastStatusTag[row.last_status || ''] || 'info'"
                  size="small"
                >
                  {{ row.last_status || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="最近执行" width="170">
              <template #default="{ row }">{{ fmtTime(row.last_run_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="success" @click="onRun(row)">执行</el-button>
                <el-button link type="primary" @click="openLogs(row)">日志</el-button>
                <el-button link type="danger" @click="onDeleteTask(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="taskQuery.page"
            v-model:page-size="taskQuery.size"
            :total="taskTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchTasks"
            @size-change="fetchTasks"
          />
        </el-tab-pane>

        <!-- ===== 登录日志 ===== -->
        <el-tab-pane label="登录日志" name="login">
          <div class="toolbar">
            <el-input
              v-model="loginQuery.username"
              placeholder="按用户名搜索"
              clearable
              style="width: 200px"
              @keyup.enter="loginSearch"
              @clear="loginSearch"
            />
            <el-select
              v-model="loginQuery.success_filter"
              placeholder="全部结果"
              clearable
              style="width: 130px"
              @change="loginSearch"
            >
              <el-option
                v-for="o in loginFilterOptions"
                :key="o.value"
                :label="o.label"
                :value="o.value"
              />
            </el-select>
            <el-button type="primary" plain @click="loginSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button @click="fetchLoginLogs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="loginLoading" :data="loginList" stripe border>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="username" label="用户名" min-width="120" show-overflow-tooltip />
            <el-table-column prop="uid" label="用户ID" width="90" />
            <el-table-column prop="ip" label="IP" min-width="130" show-overflow-tooltip />
            <el-table-column prop="user_agent" label="User-Agent" min-width="200" show-overflow-tooltip />
            <el-table-column label="结果" width="90">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                  {{ row.success ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="msg" label="信息" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">{{ row.msg || '-' }}</template>
            </el-table-column>
            <el-table-column label="时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="loginQuery.page"
            v-model:page-size="loginQuery.size"
            :total="loginTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchLoginLogs"
            @size-change="loginSearch"
          />
        </el-tab-pane>

        <!-- ===== 审计日志 ===== -->
        <el-tab-pane label="审计日志" name="audit">
          <div class="toolbar">
            <el-input
              v-model="auditQuery.resource"
              placeholder="按资源过滤，如 datasource"
              clearable
              style="width: 200px"
              @keyup.enter="auditSearch"
              @clear="auditSearch"
            />
            <el-input-number
              v-model="auditQuery.uid"
              :min="1"
              controls-position="right"
              placeholder="用户ID"
              style="width: 150px"
            />
            <el-button type="primary" plain @click="auditSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button @click="fetchAuditLogs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="auditLoading" :data="auditList" stripe border>
            <el-table-column type="expand">
              <template #default="{ row }">
                <pre class="json-pre">{{ prettyJson(row.detail_json) || '（无详情）' }}</pre>
              </template>
            </el-table-column>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="uid" label="用户ID" width="90" />
            <el-table-column prop="username" label="用户名" min-width="110" show-overflow-tooltip />
            <el-table-column prop="action" label="操作" min-width="110" show-overflow-tooltip />
            <el-table-column prop="resource" label="资源" min-width="120" show-overflow-tooltip />
            <el-table-column prop="resource_id" label="资源ID" min-width="100" show-overflow-tooltip>
              <template #default="{ row }">{{ row.resource_id || '-' }}</template>
            </el-table-column>
            <el-table-column prop="ip" label="IP" min-width="120" show-overflow-tooltip />
            <el-table-column label="时间" width="170">
              <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="auditQuery.page"
            v-model:page-size="auditQuery.size"
            :total="auditTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchAuditLogs"
            @size-change="auditSearch"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新建归档任务 -->
    <el-dialog v-model="dialogVisible" title="新建归档任务" width="620px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="如：订单流水按月归档" maxlength="64" />
        </el-form-item>
        <el-form-item label="数据源" prop="datasource_id">
          <el-select v-model="form.datasource_id" filterable style="width: 100%">
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（${d.type}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="库名">
          <el-input v-model="form.db_name" placeholder="留空取数据源默认库（仅字母/数字/下划线）" />
        </el-form-item>
        <el-form-item label="源表名" prop="table_name">
          <el-input v-model="form.table_name" placeholder="如 dc_order_log" />
        </el-form-item>
        <el-form-item label="归档条件">
          <el-input
            v-model="form.condition_sql"
            type="textarea"
            :rows="3"
            placeholder="可选 WHERE 条件，如：created_at < '2024-01-01'"
          />
        </el-form-item>
        <el-form-item label="归档表" prop="archive_table">
          <el-input v-model="form.archive_table" placeholder="如 dc_order_log_archive" />
        </el-form-item>
        <el-form-item label="删除源数据">
          <el-switch v-model="form.delete_source" active-text="归档后删除源表数据" />
        </el-form-item>
        <el-form-item label="批大小">
          <el-input-number v-model="form.batch_size" :min="1" :max="100000" controls-position="right" />
        </el-form-item>
        <el-form-item label="调度周期">
          <el-input v-model="form.cron" placeholder="Quartz Cron，如 0 0 3 * * ?（可留空手动执行）" maxlength="64" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 归档日志 -->
    <el-dialog v-model="logDialogVisible" :title="`归档日志：${logQuery.task_name}`" width="720px">
      <el-table v-loading="logLoading" :data="logList" stripe border size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="归档行数" width="100">
          <template #default="{ row }">{{ row.archived_rows ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="删除行数" width="100">
          <template #default="{ row }">{{ row.deleted_rows ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="lastStatusTag[row.status || ''] || 'info'" size="small">
              {{ row.status || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="msg" label="信息" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.msg || '-' }}</template>
        </el-table-column>
        <el-table-column label="执行时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.run_at) }}</template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="logQuery.page"
        v-model:page-size="logQuery.size"
        :total="logTotal"
        layout="total, prev, pager, next"
        small
        class="pager"
        @current-change="fetchLogs"
      />
      <template #footer>
        <el-button @click="logDialogVisible = false">关闭</el-button>
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
.json-pre {
  margin: 0;
  padding: 10px 12px;
  background: #f4f4f5;
  border-radius: 6px;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 260px;
  overflow: auto;
}
</style>
