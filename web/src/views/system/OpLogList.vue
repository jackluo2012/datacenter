<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { auditLogList, type OpsAuditLog } from '@/api/ops'
import { ApiError } from '@/api/http'

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

// ---------- 列表 ----------
const loading = ref(false)
const list = ref<OpsAuditLog[]>([])
const total = ref(0)
const query = reactive({
  resource: '',
  uid: undefined as number | undefined,
  page: 1,
  size: 10,
})

async function fetchList() {
  loading.value = true
  try {
    const resp = await auditLogList({
      resource: query.resource || undefined,
      uid: query.uid || undefined,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取审计日志失败'))
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

onMounted(fetchList)
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-input
          v-model="query.resource"
          placeholder="按资源过滤（如 /sys/user/create）"
          clearable
          style="width: 240px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-input-number
          v-model="query.uid"
          :min="1"
          controls-position="right"
          placeholder="用户 ID"
          style="width: 150px"
        />
        <el-button type="primary" plain @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表（展开行看 detail_json） -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="detail-box">
              <pre v-if="row.detail_json" class="json-pre">{{ prettyJson(row.detail_json) }}</pre>
              <span v-else class="muted">无请求详情</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="uid" label="用户ID" width="90">
          <template #default="{ row }">{{ row.uid ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户" min-width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="action" label="动作" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.action || '-' }}</template>
        </el-table-column>
        <el-table-column prop="resource" label="资源" min-width="190" show-overflow-tooltip>
          <template #default="{ row }">{{ row.resource || '-' }}</template>
        </el-table-column>
        <el-table-column prop="resource_id" label="资源ID" min-width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.resource_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="130">
          <template #default="{ row }">{{ row.ip || '-' }}</template>
        </el-table-column>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
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
.detail-box {
  padding: 4px 16px 10px 48px;
}
.json-pre {
  margin: 0;
  max-height: 320px;
  overflow: auto;
  padding: 10px 12px;
  background: #282c34;
  color: #98c379;
  border-radius: 6px;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
.muted {
  color: #909399;
}
</style>
