<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  dataSourceList,
  createDataSource,
  updateDataSource,
  deleteDataSource,
  dataSourceTest,
  type DataSourceInfo,
} from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 列表状态 ----------
const loading = ref(false)
const list = ref<DataSourceInfo[]>([])
const total = ref(0)
const query = reactive({ type: '', page: 1, size: 10 })

const typeOptions = ['mysql', 'clickhouse', 'postgresql', 'hive', 'doris', 'starrocks']

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

async function fetchList() {
  loading.value = true
  try {
    const resp = await dataSourceList({ ...query, type: query.type || undefined })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取数据源列表失败')
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
const formRef = reactive({
  name: '',
  type: 'mysql',
  host: '',
  port: 3306,
  db_name: '',
  username: '',
  password: '',
  owner: '',
})

function openCreate() {
  editingId.value = null
  Object.assign(formRef, {
    name: '',
    type: 'mysql',
    host: '',
    port: 3306,
    db_name: '',
    username: '',
    password: '',
    owner: '',
  })
  dialogVisible.value = true
}

function openEdit(row: DataSourceInfo) {
  editingId.value = row.id ?? null
  Object.assign(formRef, {
    name: row.name || '',
    type: row.type || 'mysql',
    host: row.host || '',
    port: row.port || 3306,
    db_name: row.db_name || '',
    username: row.username || '',
    password: '',
    owner: row.owner || '',
  })
  dialogVisible.value = true
}

async function onSave() {
  if (!formRef.name || !formRef.type) {
    ElMessage.warning('请填写名称与类型')
    return
  }
  saving.value = true
  try {
    if (editingId.value === null) {
      await createDataSource({ ...formRef })
      ElMessage.success('创建成功')
    } else {
      // 编辑时不回填密码，留空则不更新密码字段
      const { password, ...rest } = formRef
      await updateDataSource({ id: editingId.value, ...rest, password: password || undefined })
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
async function onDelete(row: DataSourceInfo) {
  const ok = await ElMessageBox.confirm(
    `确定删除数据源「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  ).then(() => true).catch(() => false)
  if (!ok || row.id === undefined) return
  try {
    await deleteDataSource(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

// ---------- 测试连接 ----------
const testingId = ref<number | null>(null)
async function onTest(row: DataSourceInfo) {
  if (row.id === undefined) return
  testingId.value = row.id
  try {
    const resp = await dataSourceTest(row.id)
    if (resp.ok) {
      ElMessage.success(`连接成功，延迟 ${resp.latency_ms}ms`)
    } else {
      ElMessage.error(`连接失败：${resp.err_msg || '未知错误'}`)
    }
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '测试连接失败')
  } finally {
    testingId.value = null
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
          v-model="query.type"
          placeholder="全部类型"
          clearable
          style="width: 160px"
          @change="onSearch"
        >
          <el-option v-for="t in typeOptions" :key="t" :label="t" :value="t" />
        </el-select>
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建数据源
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="64" />
        <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="type" label="类型" width="110">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="地址" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.host }}:{{ row.port }}</template>
        </el-table-column>
        <el-table-column prop="db_name" label="数据库" min-width="120" show-overflow-tooltip />
        <el-table-column prop="username" label="用户名" width="110" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="owner" label="负责人" width="100" />
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button
              link
              type="success"
              :loading="testingId === row.id"
              @click="onTest(row)"
            >
              测试
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
      :title="editingId === null ? '新建数据源' : '编辑数据源'"
      width="560px"
      destroy-on-close
    >
      <el-form :model="formRef" label-width="90px">
        <el-form-item label="名称" required>
          <el-input v-model="formRef.name" placeholder="如：数仓 MySQL 主库" maxlength="64" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="formRef.type" style="width: 100%">
            <el-option v-for="t in typeOptions" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机">
          <el-input v-model="formRef.host" placeholder="127.0.0.1" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="formRef.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="数据库">
          <el-input v-model="formRef.db_name" placeholder="db 名称" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="formRef.username" autocomplete="off" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="formRef.password"
            type="password"
            show-password
            autocomplete="new-password"
            :placeholder="editingId === null ? '请输入密码' : '留空则不修改'"
          />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="formRef.owner" placeholder="如：data-team" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
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
