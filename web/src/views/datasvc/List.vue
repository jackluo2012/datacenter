<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  apiList,
  createApi,
  updateApi,
  deleteApi,
  publishApi,
  offlineApi,
  registerApp,
  grantApi,
  type ApiInfo,
  type AppResp,
} from '@/api/datasvc'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

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
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
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
</style>
