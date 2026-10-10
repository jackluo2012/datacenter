<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  createMetric,
  updateMetric,
  deleteMetric,
  metricList,
  bindAtoms,
  metricAtoms,
  type MetricInfo,
} from '@/api/metric'
import { ApiError } from '@/api/http'

// ---------- 文案映射 ----------
const metricTypeOptions = [
  { value: 'atom', label: '原子指标' },
  { value: 'derive', label: '派生指标' },
]
const metricTypeText: Record<string, string> = Object.fromEntries(
  metricTypeOptions.map((o) => [o.value, o.label]),
)
const metricTypeTag: Record<string, 'success' | 'warning' | 'info'> = {
  atom: 'success',
  derive: 'warning',
}
const aggTypeOptions = [
  { value: 'sum', label: '求和 sum' },
  { value: 'count', label: '计数 count' },
  { value: 'avg', label: '均值 avg' },
  { value: 'max', label: '最大 max' },
  { value: 'min', label: '最小 min' },
  { value: 'distinct', label: '去重数 distinct' },
]

/** JSON 字段校验：填了必须能 JSON.parse */
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

// ---------- 列表状态 ----------
const loading = ref(false)
const list = ref<MetricInfo[]>([])
const total = ref(0)
const query = reactive({ metric_type: '', keyword: '', page: 1, size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const resp = await metricList({
      metric_type: query.metric_type || undefined,
      keyword: query.keyword.trim() || undefined,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取指标列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 新建 / 编辑对话框 ----------
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formEl = ref<FormInstance>()
const form = reactive({
  name_en: '',
  name_cn: '',
  metric_type: 'atom',
  biz_field: '',
  agg_type: 'sum',
  expr: '',
  dimension_json: '',
  comment: '',
})

const rules = reactive<FormRules>({
  name_en: [{ required: true, message: '请输入英文标识', trigger: 'blur' }],
  name_cn: [{ required: true, message: '请输入中文名称', trigger: 'blur' }],
  metric_type: [{ required: true, message: '请选择指标类型', trigger: 'change' }],
  biz_field: [
    {
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (form.metric_type === 'atom' && !value.trim()) {
          return callback(new Error('原子指标必须填写业务口径字段'))
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  agg_type: [{ required: true, message: '请选择聚合方式', trigger: 'change' }],
  expr: [
    {
      validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
        if (form.metric_type === 'derive' && !value.trim()) {
          return callback(new Error('派生指标必须填写计算表达式'))
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  dimension_json: [{ validator: jsonValidator('dimension_json'), trigger: 'blur' }],
})

function resetForm() {
  Object.assign(form, {
    name_en: '',
    name_cn: '',
    metric_type: 'atom',
    biz_field: '',
    agg_type: 'sum',
    expr: '',
    dimension_json: '',
    comment: '',
  })
}

function openCreate() {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: MetricInfo) {
  editingId.value = row.id ?? null
  Object.assign(form, {
    name_en: row.name_en || '',
    name_cn: row.name_cn || '',
    metric_type: row.metric_type || 'atom',
    biz_field: row.biz_field || '',
    agg_type: row.agg_type || 'sum',
    expr: row.expr || '',
    dimension_json: row.dimension_json || '',
    comment: row.comment || '',
  })
  dialogVisible.value = true
}

async function onSave() {
  const valid = await formEl.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (editingId.value === null) {
      await createMetric({
        name_en: form.name_en.trim(),
        name_cn: form.name_cn.trim(),
        metric_type: form.metric_type,
        biz_field: form.metric_type === 'atom' ? form.biz_field.trim() : undefined,
        agg_type: form.metric_type === 'atom' ? form.agg_type : undefined,
        expr: form.metric_type === 'derive' ? form.expr.trim() : undefined,
        dimension_json: form.dimension_json.trim() || undefined,
        comment: form.comment.trim() || undefined,
      })
      ElMessage.success('指标创建成功')
    } else {
      // 更新接口仅支持 name_cn/expr/dimension_json/comment，其余字段创建后不可改
      await updateMetric({
        id: editingId.value,
        name_cn: form.name_cn.trim(),
        expr: form.metric_type === 'derive' ? form.expr.trim() : undefined,
        dimension_json: form.dimension_json.trim() || undefined,
        comment: form.comment.trim() || undefined,
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

// ---------- 删除 ----------
async function onDelete(row: MetricInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除指标「${row.name_cn || row.name_en}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok) return
  try {
    await deleteMetric(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ---------- 绑定原子指标（派生指标） ----------
const bindVisible = ref(false)
const bindSaving = ref(false)
const bindLoading = ref(false)
const bindTarget = ref<MetricInfo | null>(null)
const bindAtomIds = ref<number[]>([])
const atomOptions = ref<MetricInfo[]>([])

async function openBind(row: MetricInfo) {
  bindTarget.value = row
  bindAtomIds.value = []
  atomOptions.value = []
  bindVisible.value = true
  bindLoading.value = true
  try {
    // 并行拉取：候选原子指标 + 当前已绑定的原子指标（回显）
    const [atomsResp, boundResp] = await Promise.all([
      metricList({ metric_type: 'atom', page: 1, size: 100 }),
      row.id !== undefined ? metricAtoms(row.id) : Promise.resolve({ list: [], total: 0 }),
    ])
    atomOptions.value = atomsResp.list || []
    const boundIds = new Set((boundResp.list || []).map((m) => m.id))
    bindAtomIds.value = [...boundIds].filter((id): id is number => id !== undefined)
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '加载原子指标失败')
  } finally {
    bindLoading.value = false
  }
}

async function onBindAtoms() {
  if (!bindTarget.value?.id) return
  if (bindAtomIds.value.length === 0) {
    ElMessage.warning('请至少选择一个原子指标')
    return
  }
  bindSaving.value = true
  try {
    await bindAtoms({
      derive_metric_id: bindTarget.value.id,
      atom_metric_ids: bindAtomIds.value,
    })
    ElMessage.success(`已为「${bindTarget.value.name_cn || bindTarget.value.name_en}」绑定 ${bindAtomIds.value.length} 个原子指标`)
    bindVisible.value = false
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '绑定失败')
  } finally {
    bindSaving.value = false
  }
}

onMounted(fetchList)
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-select
          v-model="query.metric_type"
          placeholder="全部类型"
          clearable
          style="width: 140px"
          @change="onSearch"
        >
          <el-option v-for="t in metricTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
        <el-input
          v-model="query.keyword"
          placeholder="按名称（中/英文）搜索"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建指标
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="64" />
        <el-table-column prop="name_en" label="英文标识" min-width="130" show-overflow-tooltip />
        <el-table-column prop="name_cn" label="中文名称" min-width="130" show-overflow-tooltip />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="metricTypeTag[row.metric_type] || 'info'">
              {{ metricTypeText[row.metric_type] || row.metric_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="口径字段" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.biz_field || '-' }}</template>
        </el-table-column>
        <el-table-column label="聚合方式" width="90" align="center">
          <template #default="{ row }">{{ row.agg_type || '-' }}</template>
        </el-table-column>
        <el-table-column prop="expr" label="表达式" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.expr || '-' }}</template>
        </el-table-column>
        <el-table-column prop="dimension_json" label="维度" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.dimension_json || '-' }}</template>
        </el-table-column>
        <el-table-column prop="comment" label="备注" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.comment || '-' }}</template>
        </el-table-column>
        <el-table-column prop="owner" label="负责人" width="100" show-overflow-tooltip />
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button
              v-if="row.metric_type === 'derive'"
              link
              type="success"
              @click="openBind(row)"
            >
              绑定原子
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
      :title="editingId === null ? '新建指标' : '编辑指标'"
      width="600px"
      destroy-on-close
    >
      <el-form ref="formEl" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="英文标识" prop="name_en">
          <el-input
            v-model="form.name_en"
            :disabled="editingId !== null"
            placeholder="如 order_amount"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item label="中文名称" prop="name_cn">
          <el-input v-model="form.name_cn" placeholder="如 订单金额" maxlength="64" />
        </el-form-item>
        <el-form-item label="指标类型" prop="metric_type">
          <el-radio-group v-model="form.metric_type" :disabled="editingId !== null">
            <el-radio v-for="t in metricTypeOptions" :key="t.value" :value="t.value">
              {{ t.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <!-- 原子指标：业务口径字段 + 聚合方式 -->
        <template v-if="form.metric_type === 'atom'">
          <el-form-item label="口径字段" prop="biz_field">
            <el-input
              v-model="form.biz_field"
              :disabled="editingId !== null"
              placeholder="如 orders.amount"
            />
          </el-form-item>
          <el-form-item label="聚合方式" prop="agg_type">
            <el-select v-model="form.agg_type" :disabled="editingId !== null" style="width: 100%">
              <el-option v-for="a in aggTypeOptions" :key="a.value" :label="a.label" :value="a.value" />
            </el-select>
          </el-form-item>
        </template>
        <!-- 派生指标：计算表达式 -->
        <el-form-item v-else label="计算表达式" prop="expr">
          <el-input
            v-model="form.expr"
            type="textarea"
            :rows="4"
            placeholder="基于原子指标的表达式，如：sum(orders.amount) / count(distinct orders.user_id)"
          />
        </el-form-item>
        <el-form-item label="维度 JSON" prop="dimension_json">
          <el-input
            v-model="form.dimension_json"
            type="textarea"
            :rows="2"
            placeholder='可选，统计维度，如：["date","city"]'
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.comment" type="textarea" :rows="2" placeholder="可选，口径说明" />
        </el-form-item>
      </el-form>
      <el-alert
        v-if="editingId !== null"
        title="编辑仅可更新中文名称、表达式、维度与备注；类型、口径字段、聚合方式创建后不可修改"
        type="info"
        :closable="false"
        show-icon
      />
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 绑定原子指标对话框 -->
    <el-dialog
      v-model="bindVisible"
      :title="`绑定原子指标 · ${bindTarget?.name_cn || bindTarget?.name_en || ''}`"
      width="520px"
      destroy-on-close
    >
      <el-form v-loading="bindLoading" label-width="100px">
        <el-form-item label="原子指标">
          <el-select
            v-model="bindAtomIds"
            multiple
            filterable
            placeholder="选择该派生指标引用的原子指标（可多选）"
            style="width: 100%"
          >
            <el-option
              v-for="a in atomOptions"
              :key="a.id"
              :label="`#${a.id} ${a.name_cn || a.name_en}（${a.biz_field || '-'} · ${a.agg_type || '-'}）`"
              :value="a.id!"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <el-alert
        title="绑定关系用于口径血缘追踪：派生指标由哪些原子指标计算而来"
        type="info"
        :closable="false"
        show-icon
      />
      <template #footer>
        <el-button @click="bindVisible = false">取消</el-button>
        <el-button type="primary" :loading="bindSaving" @click="onBindAtoms">确认绑定</el-button>
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
