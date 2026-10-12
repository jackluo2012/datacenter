<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'JobJob' })
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  createSyncTask,
  syncTaskList,
  deleteSyncTask,
  deployTask,
  triggerTask,
  taskStatus,
  schedulerInfo,
  type SyncTask,
  type TaskStatusResp,
  type SchedulerInfoResp,
} from '@/api/job'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 列表 ----------
const loading = ref(false)
const list = ref<SyncTask[]>([])
const total = ref(0)
const query = reactive({ status: '', page: 1, size: 10 })

const statusOptions = ['draft', 'deployed', 'running', 'failed']

const statusTag: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger' }> = {
  draft: { label: '草稿', type: 'info' },
  deployed: { label: '已部署', type: 'success' },
  running: { label: '运行中', type: 'warning' },
  failed: { label: '失败', type: 'danger' },
}

function statusOf(s?: string) {
  return statusTag[s || ''] || { label: s || '-', type: 'info' }
}

const modeLabel: Record<string, string> = {
  offline: '离线批量',
  cdc: '实时增量',
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
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

async function fetchList() {
  loading.value = true
  try {
    const resp = await syncTaskList({ ...query, status: query.status || undefined })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取同步任务列表失败'))
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 新建对话框 ----------
const dialogVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  source_ds_id: undefined as number | undefined,
  target_ds_id: undefined as number | undefined,
  mode: 'offline',
  mapping_json: '',
  cron: '',
})
const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  source_ds_id: [{ required: true, message: '请选择源数据源', trigger: 'change' }],
  mode: [{ required: true, message: '请选择同步模式', trigger: 'change' }],
}

function openCreate() {
  Object.assign(form, {
    name: '',
    source_ds_id: undefined,
    target_ds_id: undefined,
    mode: 'offline',
    mapping_json: '',
    cron: '',
  })
  dialogVisible.value = true
}

async function onSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  // mapping_json 允许为空，非空时提交前 parse 校验（须为 JSON 对象，对齐后端）
  let mappingJson = form.mapping_json.trim()
  if (mappingJson) {
    try {
      const v = JSON.parse(mappingJson)
      if (!v || typeof v !== 'object' || Array.isArray(v)) {
        ElMessage.error('mapping_json 须为 JSON 对象（如 {"src":"ods_order","dst":"dwd_order"}）')
        return
      }
      mappingJson = JSON.stringify(v)
    } catch {
      ElMessage.error('mapping_json 不是合法 JSON，请检查格式')
      return
    }
  }
  saving.value = true
  try {
    await createSyncTask({
      name: form.name.trim(),
      source_ds_id: form.source_ds_id as number,
      target_ds_id: form.target_ds_id || undefined,
      mode: form.mode,
      mapping_json: mappingJson || undefined,
      cron: form.cron.trim() || undefined,
    })
    ElMessage.success('任务创建成功，可点击「部署」推送到调度器')
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '创建任务失败'))
  } finally {
    saving.value = false
  }
}

// ---------- 行操作 ----------
async function onDeploy(row: SyncTask) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定将任务「${row.name}」部署到 DolphinScheduler 吗？`,
    '部署确认',
    { confirmButtonText: '部署', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await deployTask(row.id)
    if (resp.ok) {
      ElMessageBox.alert(
        `部署成功${resp.ds_workflow_code ? `，工作流编码 ${resp.ds_workflow_code}` : ''}${
          resp.msg ? `\n\n${resp.msg}` : ''
        }`,
        '部署结果',
        { confirmButtonText: '知道了', type: 'success' },
      ).catch(() => {})
    } else {
      ElMessageBox.alert(resp.msg || '部署失败，未返回原因', '部署结果', {
        confirmButtonText: '知道了',
        type: 'error',
      }).catch(() => {})
    }
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '部署请求失败'))
  }
}

async function onTrigger(row: SyncTask) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定手动触发任务「${row.name}」吗？`,
    '触发确认',
    { confirmButtonText: '触发', cancelButtonText: '取消', type: 'info' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    const resp = await triggerTask(row.id)
    if (resp.ok) {
      ElMessageBox.alert(
        `触发成功${resp.ds_process_instance_id ? `，调度器实例 ID ${resp.ds_process_instance_id}` : ''}${
          resp.msg ? `\n\n${resp.msg}` : ''
        }`,
        '触发结果',
        { confirmButtonText: '知道了', type: 'success' },
      ).catch(() => {})
    } else {
      ElMessageBox.alert(resp.msg || '触发失败，未返回原因', '触发结果', {
        confirmButtonText: '知道了',
        type: 'error',
      }).catch(() => {})
    }
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '触发请求失败'))
  }
}

// 状态弹窗
const statusDialogVisible = ref(false)
const statusLoading = ref(false)
const statusResp = ref<TaskStatusResp | null>(null)
const statusTaskName = ref('')

async function onShowStatus(row: SyncTask) {
  if (row.id === undefined) return
  statusTaskName.value = row.name || `#${row.id}`
  statusDialogVisible.value = true
  statusLoading.value = true
  try {
    statusResp.value = await taskStatus(row.id)
  } catch (e) {
    ElMessage.error(errText(e, '查询任务状态失败'))
    statusDialogVisible.value = false
  } finally {
    statusLoading.value = false
  }
}

async function onDelete(row: SyncTask) {
  if (row.id === undefined) return
  const deployed = row.status && row.status !== 'draft'
  const ok = await ElMessageBox.confirm(
    deployed
      ? `任务「${row.name}」当前状态「${statusOf(row.status).label}」，已部署到调度器，删除请求可能被拒绝（需先在调度器侧下线）。确定继续删除吗？`
      : `确定删除任务「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteSyncTask(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '删除任务失败（已部署任务需先在调度器侧下线）'))
  }
}

// ---------- 调度器连接 ----------
const dsInfoDialogVisible = ref(false)
const dsInfoLoading = ref(false)
const dsInfo = ref<SchedulerInfoResp | null>(null)

async function onShowScheduler() {
  dsInfoDialogVisible.value = true
  dsInfoLoading.value = true
  try {
    dsInfo.value = await schedulerInfo()
  } catch (e) {
    ElMessage.error(errText(e, '获取调度器连接信息失败'))
    dsInfoDialogVisible.value = false
  } finally {
    dsInfoLoading.value = false
  }
}

onMounted(() => {
  fetchList()
  fetchDsOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-select
          v-model="query.status"
          placeholder="全部状态"
          clearable
          style="width: 140px"
          @change="onSearch"
        >
          <el-option v-for="s in statusOptions" :key="s" :label="statusOf(s).label" :value="s" />
        </el-select>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建任务
        </el-button>
        <el-button type="success" plain @click="onShowScheduler">
          <el-icon><Connection /></el-icon>&nbsp;调度器连接
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="任务名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="源数据源" width="120">
          <template #default="{ row }">
            {{ dsOptions.find((d) => d.id === row.source_ds_id)?.name || row.source_ds_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="目标数据源" width="120">
          <template #default="{ row }">
            {{ dsOptions.find((d) => d.id === row.target_ds_id)?.name || row.target_ds_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="同步模式" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.mode === 'cdc' ? 'warning' : 'primary'">
              {{ modeLabel[row.mode] || row.mode || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cron" label="调度周期" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.cron || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusOf(row.status).type" size="small">{{ statusOf(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="工作流编码" width="110">
          <template #default="{ row }">{{ row.ds_workflow_code || '-' }}</template>
        </el-table-column>
        <el-table-column label="最近运行" width="170">
          <template #default="{ row }">{{ fmtTime(row.last_run_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click="onDeploy(row)">部署</el-button>
            <el-button
              link
              type="warning"
              :disabled="row.status === 'draft'"
              @click="onTrigger(row)"
            >
              触发
            </el-button>
            <el-button link type="primary" @click="onShowStatus(row)">状态</el-button>
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
    </el-card>

    <!-- 新建任务对话框 -->
    <el-dialog v-model="dialogVisible" title="新建同步任务" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="如：ods_order 每日全量同步" maxlength="128" />
        </el-form-item>
        <el-form-item label="源数据源" prop="source_ds_id">
          <el-select v-model="form.source_ds_id" filterable style="width: 100%">
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（${d.type}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标数据源">
          <el-select
            v-model="form.target_ds_id"
            filterable
            clearable
            placeholder="可选，缺省同源"
            style="width: 100%"
          >
            <el-option
              v-for="d in dsOptions"
              :key="d.id"
              :label="`${d.name}（${d.type}）`"
              :value="d.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="同步模式" prop="mode">
          <el-select v-model="form.mode" style="width: 100%">
            <el-option label="离线批量（offline）" value="offline" />
            <el-option label="实时增量（cdc）" value="cdc" />
          </el-select>
        </el-form-item>
        <el-form-item label="表映射">
          <el-input
            v-model="form.mapping_json"
            type="textarea"
            :rows="5"
            placeholder='可选，JSON 对象，如：{"source_table":"ods_order","target_table":"dwd_order","columns":{"id":"id"}}'
          />
        </el-form-item>
        <el-form-item label="调度周期">
          <el-input v-model="form.cron" placeholder="Quartz Cron 表达式，如 0 0 2 * * ?" maxlength="64" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 任务状态弹窗 -->
    <el-dialog v-model="statusDialogVisible" :title="`任务状态：${statusTaskName}`" width="480px">
      <div v-loading="statusLoading">
        <template v-if="statusResp">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="调度器状态">
              <el-tag :type="statusOf(statusResp.status).type" size="small">
                {{ statusResp.status || '未知' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="调度器实例 ID">
              {{ statusResp.ds_instance_id || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="信息">{{ statusResp.msg || '-' }}</el-descriptions-item>
          </el-descriptions>
        </template>
      </div>
      <template #footer>
        <el-button @click="statusDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 调度器连接信息弹窗 -->
    <el-dialog v-model="dsInfoDialogVisible" title="调度器连接信息（DolphinScheduler）" width="520px">
      <div v-loading="dsInfoLoading">
        <template v-if="dsInfo">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="已配置">
              <el-tag :type="dsInfo.configured ? 'success' : 'info'" size="small">
                {{ dsInfo.configured ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="连接状态">
              <el-tag :type="dsInfo.connected ? 'success' : 'danger'" size="small">
                {{ dsInfo.connected ? '已连通' : '不可达' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="地址">{{ dsInfo.base_url || '-' }}</el-descriptions-item>
            <el-descriptions-item label="信息">{{ dsInfo.msg || '-' }}</el-descriptions-item>
          </el-descriptions>
        </template>
      </div>
      <template #footer>
        <el-button @click="dsInfoDialogVisible = false">关闭</el-button>
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
</style>
