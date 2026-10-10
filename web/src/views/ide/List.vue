<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  saveScript,
  scriptList,
  scriptInfo,
  deleteScript,
  runScript,
  type Script,
  type RunScriptResp,
} from '@/api/ide'
import { dataSourceList, type DataSourceInfo } from '@/api/datasource'
import { ApiError } from '@/api/http'

// ---------- 左侧脚本列表 ----------
const listLoading = ref(false)
const scripts = ref<Script[]>([])
const scriptTotal = ref(0)
const listQuery = reactive({ keyword: '', page: 1, size: 20 })

async function fetchScripts() {
  listLoading.value = true
  try {
    const resp = await scriptList({
      ...listQuery,
      keyword: listQuery.keyword || undefined,
    })
    scripts.value = resp.list || []
    scriptTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取脚本列表失败')
  } finally {
    listLoading.value = false
  }
}

function onSearch() {
  listQuery.page = 1
  fetchScripts()
}

// ---------- 右侧编辑器 ----------
const dsOptions = ref<DataSourceInfo[]>([])
async function fetchDsOptions() {
  try {
    const resp = await dataSourceList({ page: 1, size: 200 })
    dsOptions.value = resp.list || []
  } catch {
    /* 下拉数据拉取失败不阻断页面 */
  }
}

const currentId = ref<number | null>(null)
const editor = reactive({
  name: '',
  datasource_id: undefined as number | undefined,
  db_name: '',
  sql_text: '',
})
const currentVersion = ref(0)
const dirty = ref(false)
const saving = ref(false)
const running = ref(false)

/** 编辑器任一字段变更即标记未保存（模板事件驱动，避免程序化加载误标） */
function markDirty() {
  dirty.value = true
}

function selectScript(row: Script) {
  if (dirty.value && currentId.value !== (row.id ?? null)) {
    ElMessage.warning('当前脚本有未保存修改')
  }
  loadScript(row.id as number)
}

async function loadScript(id: number) {
  try {
    const s = await scriptInfo(id)
    currentId.value = s.id ?? null
    editor.name = s.name || ''
    editor.datasource_id = s.datasource_id || undefined
    editor.db_name = s.db_name || ''
    editor.sql_text = s.sql_text || ''
    currentVersion.value = s.version || 0
    dirty.value = false
    clearResult()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '加载脚本失败')
  }
}

function newScript() {
  currentId.value = null
  editor.name = ''
  editor.datasource_id = undefined
  editor.db_name = ''
  editor.sql_text = 'SELECT 1'
  currentVersion.value = 0
  dirty.value = false
  clearResult()
}

async function onSave(): Promise<Script | null> {
  if (!editor.name.trim()) {
    ElMessage.warning('请填写脚本名称')
    return null
  }
  if (!editor.datasource_id) {
    ElMessage.warning('请选择执行数据源')
    return null
  }
  if (!editor.sql_text.trim()) {
    ElMessage.warning('SQL 内容不能为空')
    return null
  }
  saving.value = true
  try {
    const resp = await saveScript({
      id: currentId.value ?? undefined,
      name: editor.name.trim(),
      sql_text: editor.sql_text,
      datasource_id: editor.datasource_id,
      db_name: editor.db_name.trim() || undefined,
    })
    currentId.value = resp.id ?? currentId.value
    currentVersion.value = resp.version || 0
    dirty.value = false
    ElMessage.success(
      resp.version && resp.version > 1
        ? `保存成功，版本已更新至 v${resp.version}`
        : '保存成功，当前版本 v1',
    )
    fetchScripts()
    return resp
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '保存脚本失败')
    return null
  } finally {
    saving.value = false
  }
}

async function onDeleteScript(row: Script) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除脚本「${row.name}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteScript(row.id)
    if (currentId.value === row.id) newScript()
    ElMessage.success('删除成功')
    fetchScripts()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '删除脚本失败')
  }
}

// ---------- 执行与结果 ----------
const maxRows = ref<number>(100)
const result = ref<RunScriptResp | null>(null)
const runError = ref('')

function clearResult() {
  result.value = null
  runError.value = ''
}

function parseJsonArray<T>(raw?: string): T[] {
  if (!raw) return []
  try {
    const v = JSON.parse(raw)
    return Array.isArray(v) ? (v as T[]) : []
  } catch {
    return []
  }
}

const resultColumns = computed(() => parseJsonArray<string>(result.value?.columns_json))
const resultRows = computed(() =>
  parseJsonArray<Record<string, unknown>>(result.value?.rows_json),
)

async function onRun() {
  if (currentId.value === null) {
    ElMessage.warning('请先保存脚本后再执行（执行基于已保存版本）')
    return
  }
  let runId = currentId.value
  if (dirty.value) {
    const ok = await ElMessageBox.confirm(
      '脚本有未保存修改，执行基于已保存版本。是否先保存再执行？',
      '未保存提醒',
      { confirmButtonText: '保存并执行', cancelButtonText: '取消', type: 'warning' },
    )
      .then(() => true)
      .catch(() => false)
    if (!ok) return
    const saved = await onSave()
    if (!saved || saved.id === undefined) return
    runId = saved.id
  }
  running.value = true
  clearResult()
  try {
    const resp = await runScript({
      id: runId,
      max_rows: maxRows.value || undefined,
    })
    result.value = resp
    if (resp.truncated) {
      ElMessage.warning(`结果超出 ${maxRows.value || '上限'} 行已截断，仅展示部分数据`)
    }
  } catch (e) {
    runError.value = e instanceof ApiError ? e.message : '执行失败，请稍后重试'
  } finally {
    running.value = false
  }
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => {
  fetchScripts()
  fetchDsOptions()
  newScript()
})
</script>

<template>
  <div class="ide-wrap">
    <!-- 左侧脚本列表 -->
    <el-card shadow="never" class="left-panel" :body-style="{ padding: '12px' }">
      <div class="left-toolbar">
        <el-input
          v-model="listQuery.keyword"
          placeholder="搜索脚本名称"
          clearable
          @keyup.enter="onSearch"
          @clear="onSearch"
        >
          <template #append>
            <el-button @click="onSearch">
              <el-icon><Search /></el-icon>
            </el-button>
          </template>
        </el-input>
        <el-button type="primary" class="new-btn" @click="newScript">
          <el-icon><Plus /></el-icon>&nbsp;新建
        </el-button>
      </div>

      <div v-loading="listLoading" class="script-list">
        <div
          v-for="s in scripts"
          :key="s.id"
          class="script-item"
          :class="{ active: s.id === currentId }"
          @click="selectScript(s)"
        >
          <div class="script-name">
            {{ s.name || '未命名脚本' }}
            <el-tag v-if="s.version" size="small" type="info" class="ver-tag">v{{ s.version }}</el-tag>
          </div>
          <div class="script-meta">
            <span>数据源 {{ s.datasource_id || '-' }}</span>
            <span v-if="s.db_name"> · {{ s.db_name }}</span>
            <span> · {{ fmtTime(s.last_run_at) }}</span>
          </div>
          <el-button
            link
            type="danger"
            class="del-btn"
            @click.stop="onDeleteScript(s)"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <el-empty v-if="!listLoading && scripts.length === 0" description="暂无脚本" :image-size="60" />
      </div>

      <el-pagination
        v-model:current-page="listQuery.page"
        v-model:page-size="listQuery.size"
        :total="scriptTotal"
        layout="total, prev, pager, next"
        small
        class="left-pager"
        @current-change="fetchScripts"
      />
    </el-card>

    <!-- 右侧编辑器与结果 -->
    <el-card shadow="never" class="right-panel">
      <!-- 脚本信息 -->
      <div class="editor-head">
        <el-input
          v-model="editor.name"
          placeholder="脚本名称"
          maxlength="64"
          style="width: 240px"
          @input="markDirty"
        />
        <el-select
          v-model="editor.datasource_id"
          filterable
          placeholder="执行数据源"
          style="width: 220px"
          @change="markDirty"
        >
          <el-option
            v-for="d in dsOptions"
            :key="d.id"
            :label="`${d.name}（${d.type}）`"
            :value="d.id as number"
          />
        </el-select>
        <el-input
          v-model="editor.db_name"
          placeholder="数据库（可选）"
          style="width: 160px"
          @input="markDirty"
        />
        <el-tag v-if="currentId !== null" size="small" type="info">
          v{{ currentVersion || 1 }}{{ dirty ? ' · 未保存修改' : '' }}
        </el-tag>
        <el-tag v-else size="small" type="warning">新脚本（未保存）</el-tag>
        <el-button type="primary" :loading="saving" @click="onSave">
          <el-icon><DocumentChecked /></el-icon>&nbsp;保存
        </el-button>
      </div>

      <!-- SQL 编辑器 -->
      <el-input
        v-model="editor.sql_text"
        type="textarea"
        :rows="18"
        class="sql-editor"
        placeholder="输入 SELECT 语句（仅支持查询，带行数保护）"
        spellcheck="false"
        @input="markDirty"
      />

      <!-- 执行栏 -->
      <div class="run-bar">
        <span class="run-label">最大返回行数</span>
        <el-input-number v-model="maxRows" :min="1" :max="1000" controls-position="right" />
        <el-button type="success" :loading="running" @click="onRun">
          <el-icon><VideoPlay /></el-icon>&nbsp;运行
        </el-button>
        <span class="run-tip">仅允许 SELECT 语句；后端超限自动截断</span>
      </div>

      <!-- 结果区 -->
      <div class="result-area">
        <template v-if="runError">
          <el-alert :title="`执行失败：${runError}`" type="error" :closable="false" show-icon />
        </template>
        <template v-else-if="result">
          <div class="result-meta">
            <el-tag size="small" type="success">共 {{ result.row_count }} 行</el-tag>
            <el-tag size="small" type="info">耗时 {{ result.cost_ms }}ms</el-tag>
            <el-tag v-if="result.truncated" size="small" type="warning">结果已截断</el-tag>
          </div>
          <el-table
            v-if="resultRows.length"
            :data="resultRows"
            size="small"
            border
            stripe
            max-height="360"
          >
            <el-table-column
              v-for="col in resultColumns"
              :key="col"
              :prop="col"
              :label="col"
              min-width="110"
              show-overflow-tooltip
            />
          </el-table>
          <el-empty v-else description="执行成功，无返回行" :image-size="70" />
        </template>
        <el-empty v-else description="点击「运行」查看查询结果" :image-size="70" />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.ide-wrap {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.left-panel {
  width: 300px;
  flex-shrink: 0;
}
.left-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}
.new-btn {
  flex-shrink: 0;
}
.script-list {
  min-height: 320px;
  max-height: calc(100vh - 320px);
  overflow-y: auto;
}
.script-item {
  position: relative;
  padding: 8px 30px 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 4px;
  border: 1px solid transparent;
}
.script-item:hover {
  background: #f5f7fa;
}
.script-item.active {
  background: #ecf5ff;
  border-color: #d9ecff;
}
.script-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}
.ver-tag {
  flex-shrink: 0;
}
.script-meta {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
.del-btn {
  position: absolute;
  right: 6px;
  top: 10px;
}
.left-pager {
  margin-top: 10px;
  justify-content: center;
}
.right-panel {
  flex: 1;
  min-width: 0;
}
.editor-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.sql-editor :deep(.el-textarea__inner) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
}
.run-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 12px 0;
}
.run-label {
  font-size: 13px;
  color: #606266;
}
.run-tip {
  font-size: 12px;
  color: #c0c4cc;
}
.result-area {
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
}
.result-meta {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}
</style>
