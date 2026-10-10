<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  createTagObject,
  tagObjectList,
  createTagDef,
  tagDefList,
  deleteTagDef,
  tagging,
  objectTags,
  type TagObject,
  type TagDef,
  type TagValueItem,
} from '@/api/tag'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 文案映射 ----------
const objectTypeOptions = [
  { value: 'user', label: '用户 user' },
  { value: 'member', label: '会员 member' },
  { value: 'goods', label: '商品 goods' },
]
const objectTypeText: Record<string, string> = Object.fromEntries(
  objectTypeOptions.map((o) => [o.value, o.label]),
)
/** 行内安全取文案（字段可能缺失） */
function objectTypeLabel(t?: string) {
  return (t && objectTypeText[t]) || t || '-'
}
const tagTypeOptions = [
  { value: 'rule', label: '规则标签' },
  { value: 'fact', label: '事实标签' },
  { value: 'model', label: '模型标签' },
]
const tagTypeText: Record<string, string> = Object.fromEntries(
  tagTypeOptions.map((o) => [o.value, o.label]),
)
/** 行内安全取文案（字段可能缺失） */
function tagTypeLabel(t?: string) {
  return (t && tagTypeText[t]) || t || '-'
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

// ---------- 标签对象（三个页签共用） ----------
const objectLoading = ref(false)
const objectList = ref<TagObject[]>([])
const objectTotal = ref(0)
const objectQuery = reactive({ object_type: '', page: 1, size: 10 })

const objectIdText = computed(() => {
  const map = new Map<number, TagObject>()
  objectList.value.forEach((o) => {
    if (o.id !== undefined) map.set(o.id, o)
  })
  return map
})

async function fetchObjects() {
  objectLoading.value = true
  try {
    const resp = await tagObjectList({
      object_type: objectQuery.object_type || undefined,
      page: objectQuery.page,
      size: objectQuery.size,
    })
    objectList.value = resp.list || []
    objectTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取标签对象失败')
  } finally {
    objectLoading.value = false
  }
}

function onObjectSearch() {
  objectQuery.page = 1
  fetchObjects()
}

const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 100 })
    dsOptions.value = resp.list || []
  } catch {
    // 下拉仅辅助录入，失败不打断页面
  }
}

// ---------- 新建对象对话框 ----------
const objectDialogVisible = ref(false)
const objectSaving = ref(false)
const objectFormEl = ref<FormInstance>()
const objectForm = reactive({
  object_type: 'user',
  name: '',
  ds_id: undefined as number | undefined,
  db_name: '',
  table_name: '',
  pk_column: 'id',
})

const objectRules = reactive<FormRules>({
  object_type: [{ required: true, message: '请选择对象类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入对象名称', trigger: 'blur' }],
  table_name: [{ required: true, message: '请输入物理表名', trigger: 'blur' }],
})

function openObjectCreate() {
  Object.assign(objectForm, {
    object_type: 'user',
    name: '',
    ds_id: undefined,
    db_name: '',
    table_name: '',
    pk_column: 'id',
  })
  fetchDsOptions()
  objectDialogVisible.value = true
}

async function onSaveObject() {
  const valid = await objectFormEl.value?.validate().catch(() => false)
  if (!valid) return
  objectSaving.value = true
  try {
    await createTagObject({
      object_type: objectForm.object_type,
      name: objectForm.name.trim(),
      ds_id: objectForm.ds_id || undefined,
      db_name: objectForm.db_name.trim() || undefined,
      table_name: objectForm.table_name.trim(),
      pk_column: objectForm.pk_column.trim() || undefined,
    })
    ElMessage.success('标签对象创建成功')
    objectDialogVisible.value = false
    fetchObjects()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '创建标签对象失败')
  } finally {
    objectSaving.value = false
  }
}

// ==================== 页签二：标签定义 ====================
const defLoading = ref(false)
const defList = ref<TagDef[]>([])
const defTotal = ref(0)
const defQuery = reactive({ object_id: undefined as number | undefined, page: 1, size: 10 })

async function fetchDefs() {
  defLoading.value = true
  try {
    const resp = await tagDefList({
      object_id: defQuery.object_id,
      page: defQuery.page,
      size: defQuery.size,
    })
    defList.value = resp.list || []
    defTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取标签定义失败')
  } finally {
    defLoading.value = false
  }
}

function onDefSearch() {
  defQuery.page = 1
  fetchDefs()
}

// ---------- 新建标签定义对话框 ----------
const defDialogVisible = ref(false)
const defSaving = ref(false)
const defFormEl = ref<FormInstance>()
const defForm = reactive({
  object_id: undefined as number | undefined,
  tag_key: '',
  tag_name: '',
  tag_type: 'rule',
  rule_sql: '',
  refresh_cron: '',
})

const defRules = reactive<FormRules>({
  object_id: [{ required: true, message: '请选择所属对象', trigger: 'change' }],
  tag_key: [{ required: true, message: '请输入标签 Key', trigger: 'blur' }],
  tag_name: [{ required: true, message: '请输入标签名称', trigger: 'blur' }],
})

function openDefCreate() {
  Object.assign(defForm, {
    object_id: defQuery.object_id,
    tag_key: '',
    tag_name: '',
    tag_type: 'rule',
    rule_sql: '',
    refresh_cron: '',
  })
  defDialogVisible.value = true
}

async function onSaveDef() {
  const valid = await defFormEl.value?.validate().catch(() => false)
  if (!valid) return
  defSaving.value = true
  try {
    await createTagDef({
      object_id: Number(defForm.object_id),
      tag_key: defForm.tag_key.trim(),
      tag_name: defForm.tag_name.trim(),
      tag_type: defForm.tag_type,
      rule_sql: defForm.rule_sql.trim() || undefined,
      refresh_cron: defForm.refresh_cron.trim() || undefined,
    })
    ElMessage.success('标签定义创建成功')
    defDialogVisible.value = false
    fetchDefs()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '创建标签定义失败')
  } finally {
    defSaving.value = false
  }
}

async function onDeleteDef(row: TagDef) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除标签「${row.tag_name || row.tag_key}」吗？已打的历史标签不受影响，但该定义不再刷新。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await deleteTagDef(row.id)
    ElMessage.success('删除成功')
    fetchDefs()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ==================== 页签三：打标与画像 ====================
const taggingObjectId = ref<number | undefined>(undefined)
const taggingRunning = ref(false)
const taggingResult = ref<{ tagged_rows: number; tag_id: number } | null>(null)

async function onTagging() {
  if (!taggingObjectId.value) {
    ElMessage.warning('请先选择要打标的对象')
    return
  }
  const obj = objectIdText.value.get(taggingObjectId.value)
  const ok = await ElMessageBox.confirm(
    `确定对对象「${obj?.name || `#${taggingObjectId.value}`}」执行打标吗？将按规则 SQL 批量写入标签值。`,
    '打标确认',
    { confirmButtonText: '执行打标', cancelButtonText: '取消', type: 'info' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  taggingRunning.value = true
  taggingResult.value = null
  try {
    const resp = await tagging({ object_id: taggingObjectId.value })
    taggingResult.value = { tagged_rows: resp.tagged_rows, tag_id: resp.tag_id }
    ElMessage.success(`打标完成，共写入 ${resp.tagged_rows} 行标签值`)
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '执行打标失败')
  } finally {
    taggingRunning.value = false
  }
}

// ---------- 画像查询 ----------
const profileFormEl = ref<FormInstance>()
const profileLoading = ref(false)
const profileForm = reactive({
  object_id: undefined as number | undefined,
  object_pk: '',
})
const profileRows = ref<TagValueItem[]>([])
const profileQueried = ref(false)

const profileRules = reactive<FormRules>({
  object_id: [{ required: true, message: '请选择对象', trigger: 'change' }],
  object_pk: [{ required: true, message: '请输入对象主键值', trigger: 'blur' }],
})

async function onProfileQuery() {
  const valid = await profileFormEl.value?.validate().catch(() => false)
  if (!valid) return
  profileLoading.value = true
  try {
    const resp = await objectTags({
      object_id: profileForm.object_id,
      object_pk: profileForm.object_pk.trim(),
    })
    profileRows.value = resp.tags || []
    profileQueried.value = true
    if (profileRows.value.length === 0) {
      ElMessage.info('该对象暂无标签值，请先执行打标')
    }
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '画像查询失败')
  } finally {
    profileLoading.value = false
  }
}

// ---------- 页签切换：懒加载 ----------
const activeTab = ref('object')
function onTabChange() {
  if (activeTab.value === 'def' && defList.value.length === 0 && defTotal.value === 0) {
    fetchDefs()
  } else if (activeTab.value === 'profile' && objectList.value.length === 0) {
    fetchObjects()
  }
}

onMounted(() => {
  fetchObjects()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ==================== 对象 ==================== -->
        <el-tab-pane label="标签对象" name="object">
          <div class="toolbar">
            <el-select
              v-model="objectQuery.object_type"
              placeholder="全部类型"
              clearable
              style="width: 150px"
              @change="onObjectSearch"
            >
              <el-option v-for="t in objectTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
            <el-button type="primary" @click="openObjectCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建对象
            </el-button>
            <el-button @click="fetchObjects">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="objectLoading" :data="objectList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="name" label="对象名称" min-width="140" show-overflow-tooltip />
            <el-table-column label="对象类型" width="130">
              <template #default="{ row }">
                <el-tag size="small" type="info">{{ objectTypeLabel(row.object_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ds_id" label="数据源" width="80" align="center">
              <template #default="{ row }">{{ row.ds_id || '-' }}</template>
            </el-table-column>
            <el-table-column label="数据库" width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ row.db_name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="table_name" label="物理表" min-width="150" show-overflow-tooltip />
            <el-table-column prop="pk_column" label="主键列" width="110" show-overflow-tooltip />
          </el-table>

          <el-pagination
            v-model:current-page="objectQuery.page"
            v-model:page-size="objectQuery.size"
            :total="objectTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchObjects"
            @size-change="onObjectSearch"
          />
        </el-tab-pane>

        <!-- ==================== 标签定义 ==================== -->
        <el-tab-pane label="标签定义" name="def">
          <div class="toolbar">
            <el-select
              v-model="defQuery.object_id"
              placeholder="全部对象"
              clearable
              filterable
              style="width: 220px"
              @change="onDefSearch"
            >
              <el-option
                v-for="o in objectList"
                :key="o.id"
                :label="`#${o.id} ${o.name}`"
                :value="o.id!"
              />
            </el-select>
            <el-button type="primary" @click="openDefCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建标签
            </el-button>
            <el-button @click="fetchDefs">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="defLoading" :data="defList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="所属对象" min-width="130" show-overflow-tooltip>
              <template #default="{ row }">
                <template v-if="objectIdText.get(row.object_id)">
                  #{{ row.object_id }} {{ objectIdText.get(row.object_id)?.name }}
                </template>
                <template v-else>{{ row.object_id || '-' }}</template>
              </template>
            </el-table-column>
            <el-table-column prop="tag_key" label="标签 Key" min-width="120" show-overflow-tooltip />
            <el-table-column prop="tag_name" label="标签名称" min-width="120" show-overflow-tooltip />
            <el-table-column label="标签类型" width="110" align="center">
              <template #default="{ row }">
                <el-tag size="small">{{ tagTypeLabel(row.tag_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="rule_sql" label="规则 SQL" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ row.rule_sql || '-' }}</template>
            </el-table-column>
            <el-table-column label="刷新周期" width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.refresh_cron || '-' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteDef(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="defQuery.page"
            v-model:page-size="defQuery.size"
            :total="defTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchDefs"
            @size-change="onDefSearch"
          />
        </el-tab-pane>

        <!-- ==================== 打标与画像 ==================== -->
        <el-tab-pane label="打标与画像" name="profile">
          <el-row :gutter="14">
            <!-- 左：执行打标 -->
            <el-col :span="10">
              <div class="panel-title">执行打标</div>
              <el-form label-width="90px" class="tagging-form">
                <el-form-item label="打标对象" required>
                  <el-select
                    v-model="taggingObjectId"
                    filterable
                    placeholder="选择要打标的标签对象"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="o in objectList"
                      :key="o.id"
                      :label="`#${o.id} ${o.name}（${objectTypeLabel(o.object_type)}）`"
                      :value="o.id!"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="primary"
                    :loading="taggingRunning"
                    :disabled="!taggingObjectId"
                    @click="onTagging"
                  >
                    <el-icon><VideoPlay /></el-icon>&nbsp;执行打标
                  </el-button>
                </el-form-item>
              </el-form>
              <el-alert
                v-if="taggingResult"
                :title="`打标完成：对象 #${taggingResult.tag_id} 共写入 ${taggingResult.tagged_rows} 行标签值`"
                type="success"
                :closable="false"
                show-icon
              />
              <el-alert
                v-else
                title="按标签定义的规则 SQL 批量计算并写入标签值，仅更新该对象下所有标签"
                type="info"
                :closable="false"
                show-icon
              />
            </el-col>

            <!-- 右：画像查询 -->
            <el-col :span="14">
              <div class="panel-title">对象画像查询</div>
              <el-form
                ref="profileFormEl"
                :model="profileForm"
                :rules="profileRules"
                label-width="90px"
                class="tagging-form"
              >
                <el-form-item label="对象" prop="object_id">
                  <el-select
                    v-model="profileForm.object_id"
                    filterable
                    placeholder="选择标签对象"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="o in objectList"
                      :key="o.id"
                      :label="`#${o.id} ${o.name}`"
                      :value="o.id!"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item label="主键值" prop="object_pk">
                  <el-input
                    v-model="profileForm.object_pk"
                    placeholder="输入对象主键值，如用户 ID：10001"
                    @keyup.enter="onProfileQuery"
                  />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :loading="profileLoading" @click="onProfileQuery">
                    <el-icon><Search /></el-icon>&nbsp;查询画像
                  </el-button>
                </el-form-item>
              </el-form>

              <el-table
                v-if="profileQueried"
                v-loading="profileLoading"
                :data="profileRows"
                stripe
                border
                max-height="380"
                empty-text="该对象暂无标签值"
              >
                <el-table-column prop="tag_key" label="标签 Key" min-width="110" show-overflow-tooltip />
                <el-table-column prop="tag_name" label="标签名称" min-width="110" show-overflow-tooltip />
                <el-table-column label="标签值" min-width="130" show-overflow-tooltip>
                  <template #default="{ row }">
                    <el-tag size="small" type="warning">{{ row.tag_value ?? '-' }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="打标时间" width="170">
                  <template #default="{ row }">{{ fmtTime(row.tagged_at) }}</template>
                </el-table-column>
              </el-table>
            </el-col>
          </el-row>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新建对象对话框 -->
    <el-dialog v-model="objectDialogVisible" title="新建标签对象" width="560px" destroy-on-close>
      <el-form ref="objectFormEl" :model="objectForm" :rules="objectRules" label-width="100px">
        <el-form-item label="对象类型" prop="object_type">
          <el-select v-model="objectForm.object_type" style="width: 100%">
            <el-option v-for="t in objectTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="对象名称" prop="name">
          <el-input v-model="objectForm.name" placeholder="如：平台用户" maxlength="64" />
        </el-form-item>
        <el-form-item label="数据源">
          <el-select
            v-model="objectForm.ds_id"
            filterable
            clearable
            placeholder="选择对象所在数据源"
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
        <el-form-item label="数据库">
          <el-input v-model="objectForm.db_name" placeholder="可选，如 dwd" />
        </el-form-item>
        <el-form-item label="物理表" prop="table_name">
          <el-input v-model="objectForm.table_name" placeholder="如 dwd_user_info" />
        </el-form-item>
        <el-form-item label="主键列">
          <el-input v-model="objectForm.pk_column" placeholder="默认 id" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="objectDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="objectSaving" @click="onSaveObject">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新建标签定义对话框 -->
    <el-dialog v-model="defDialogVisible" title="新建标签定义" width="600px" destroy-on-close>
      <el-form ref="defFormEl" :model="defForm" :rules="defRules" label-width="100px">
        <el-form-item label="所属对象" prop="object_id">
          <el-select
            v-model="defForm.object_id"
            filterable
            placeholder="选择标签对象"
            style="width: 100%"
          >
            <el-option
              v-for="o in objectList"
              :key="o.id"
              :label="`#${o.id} ${o.name}（${objectTypeLabel(o.object_type)}）`"
              :value="o.id!"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签 Key" prop="tag_key">
          <el-input v-model="defForm.tag_key" placeholder="如 consume_level" maxlength="64" />
        </el-form-item>
        <el-form-item label="标签名称" prop="tag_name">
          <el-input v-model="defForm.tag_name" placeholder="如 消费等级" maxlength="64" />
        </el-form-item>
        <el-form-item label="标签类型">
          <el-select v-model="defForm.tag_type" style="width: 100%">
            <el-option v-for="t in tagTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则 SQL">
          <el-input
            v-model="defForm.rule_sql"
            type="textarea"
            :rows="4"
            placeholder="规则标签的计算 SQL，如：case when amount > 10000 then '高价值' when amount > 1000 then '中价值' else '普通' end"
          />
        </el-form-item>
        <el-form-item label="刷新周期">
          <el-input v-model="defForm.refresh_cron" placeholder="可选，如 0 2 * * *" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="defDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="defSaving" @click="onSaveDef">保存</el-button>
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
.panel-title {
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--el-text-color-primary);
}
.tagging-form {
  margin-bottom: 12px;
}
</style>
