<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'MetadataMetadata' })
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  metaEntityList,
  collectDatabase,
  type MetaEntity,
} from '@/api/metadata'
import { ApiError } from '@/api/http'

// ---------- 列表 ----------
const loading = ref(false)
const list = ref<MetaEntity[]>([])
const total = ref(0)
const query = reactive({
  entity_type: '',
  keyword: '',
  page: 1,
  size: 10,
})

const typeOptions = [
  { value: 'table', label: '表' },
  { value: 'column', label: '字段' },
  { value: 'dashboard', label: '报表' },
]

function fmtRows(n?: number) {
  if (n === undefined || n === null) return '-'
  if (n >= 1e8) return `${(n / 1e8).toFixed(2)} 亿`
  if (n >= 1e4) return `${(n / 1e4).toFixed(1)} 万`
  return String(n)
}

async function fetchList() {
  loading.value = true
  try {
    const resp = await metaEntityList({
      ...query,
      entity_type: query.entity_type || undefined,
      keyword: query.keyword || undefined,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取元数据列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 采集对话框 ----------
const collectVisible = ref(false)
const collecting = ref(false)
const collectForm = reactive({ ds_id: undefined as number | undefined, db_name: '' })
const collectResult = ref<{ tables: number; columns: number } | null>(null)

function openCollect() {
  collectResult.value = null
  collectForm.ds_id = undefined
  collectForm.db_name = ''
  collectVisible.value = true
}

async function onCollect() {
  if (!collectForm.ds_id) {
    ElMessage.warning('请填写数据源 ID')
    return
  }
  collecting.value = true
  try {
    const resp = await collectDatabase({
      ds_id: collectForm.ds_id,
      db_name: collectForm.db_name || undefined,
    })
    collectResult.value = { tables: resp.tables_created, columns: resp.columns_created }
    ElMessage.success(`采集完成：新建表 ${resp.tables_created}、字段 ${resp.columns_created}`)
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '采集失败')
  } finally {
    collecting.value = false
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
          v-model="query.entity_type"
          placeholder="全部类型"
          clearable
          style="width: 130px"
          @change="onSearch"
        >
          <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
        <el-input
          v-model="query.keyword"
          placeholder="按名称/FQN 搜索"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button type="success" @click="openCollect">
          <el-icon><MagicStick /></el-icon>&nbsp;采集元数据
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="72" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.entity_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="fqn" label="FQN" min-width="220" show-overflow-tooltip />
        <el-table-column prop="comment" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column prop="engine" label="引擎" width="110" show-overflow-tooltip />
        <el-table-column label="行数" width="100" align="right">
          <template #default="{ row }">{{ fmtRows(row.rows_cnt) }}</template>
        </el-table-column>
        <el-table-column prop="classification" label="分级" width="90" />
        <el-table-column prop="owner" label="负责人" width="100" />
        <el-table-column prop="ds_id" label="数据源" width="80" />
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

    <!-- 采集对话框 -->
    <el-dialog v-model="collectVisible" title="采集元数据" width="460px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="数据源 ID" required>
          <el-input-number
            v-model="collectForm.ds_id"
            :min="1"
            placeholder="数据源 ID"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="库名">
          <el-input v-model="collectForm.db_name" placeholder="留空则采集默认库" />
        </el-form-item>
      </el-form>
      <el-alert
        v-if="collectResult"
        :title="`采集完成：新建表 ${collectResult.tables} 张、字段 ${collectResult.columns} 个`"
        type="success"
        :closable="false"
        show-icon
      />
      <template #footer>
        <el-button @click="collectVisible = false">关闭</el-button>
        <el-button type="primary" :loading="collecting" @click="onCollect">开始采集</el-button>
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
