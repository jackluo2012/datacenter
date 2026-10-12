<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'SystemLoginlog' })
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { loginLogList, type OpsLoginLog } from '@/api/ops'
import { ApiError } from '@/api/http'

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

// ---------- 列表 ----------
const loading = ref(false)
const list = ref<OpsLoginLog[]>([])
const total = ref(0)
// success_filter：0 全部 / 1 成功 / 2 失败
const query = reactive({ username: '', success_filter: 0, page: 1, size: 10 })

const successOptions = [
  { value: 0, label: '全部结果' },
  { value: 1, label: '成功' },
  { value: 2, label: '失败' },
]

async function fetchList() {
  loading.value = true
  try {
    const resp = await loginLogList({
      username: query.username || undefined,
      success_filter: query.success_filter,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取登录日志失败'))
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
        <el-select v-model="query.success_filter" style="width: 130px" @change="onSearch">
          <el-option v-for="s in successOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-input
          v-model="query.username"
          placeholder="按用户名搜索"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" plain @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column prop="uid" label="用户ID" width="90">
          <template #default="{ row }">{{ row.uid ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="结果" width="90">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" size="small">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="140">
          <template #default="{ row }">{{ row.ip || '-' }}</template>
        </el-table-column>
        <el-table-column prop="msg" label="说明" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.msg || '-' }}</template>
        </el-table-column>
        <el-table-column prop="user_agent" label="User-Agent" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.user_agent || '-' }}</template>
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
</style>
