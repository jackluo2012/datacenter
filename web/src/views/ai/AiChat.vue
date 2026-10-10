<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ask, askAndRun, type Candidate } from '@/api/ai'
import { ApiError } from '@/api/http'

/** 聊天记录条目 */
interface ChatMsg {
  role: 'user' | 'assistant'
  question?: string
  answer?: string
  candidates?: Candidate[]
  generatedSql?: string
  llmEnabled?: boolean
  costMs?: number
  // ask_and_run 执行结果
  executed?: boolean
  columns?: string[]
  rows?: Record<string, unknown>[]
  rowCount?: number
  runError?: string
}

const messages = ref<ChatMsg[]>([])
const question = ref('')
const sending = ref(false)

// 自动执行开关：调 /ai/ask_and_run 并展示结果表格
const autoRun = ref(false)
const runConfig = reactive({ datasource_id: undefined as number | undefined, max_rows: 20 })

const chatBody = ref<HTMLDivElement>()

function scrollToBottom() {
  nextTick(() => {
    chatBody.value?.scrollTo({ top: chatBody.value.scrollHeight, behavior: 'smooth' })
  })
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

async function onSend() {
  const q = question.value.trim()
  if (!q || sending.value) return
  messages.value.push({ role: 'user', question: q })
  question.value = ''
  scrollToBottom()

  sending.value = true
  try {
    if (autoRun.value) {
      if (!runConfig.datasource_id) {
        ElMessage.warning('自动执行需先填写数据源 ID')
        sending.value = false
        return
      }
      const resp = await askAndRun({
        question: q,
        auto_run: true,
        datasource_id: runConfig.datasource_id,
        max_rows: runConfig.max_rows || 20,
      })
      messages.value.push({
        role: 'assistant',
        question: q,
        answer: resp.answer,
        candidates: resp.candidates,
        generatedSql: resp.generated_sql,
        llmEnabled: resp.llm_enabled,
        costMs: resp.cost_ms,
        executed: resp.executed,
        columns: parseJsonArray<string>(resp.columns_json),
        rows: parseJsonArray<Record<string, unknown>>(resp.rows_json),
        rowCount: resp.row_count,
        runError: resp.run_error,
      })
    } else {
      const resp = await ask({ question: q })
      messages.value.push({
        role: 'assistant',
        question: q,
        answer: resp.answer,
        candidates: resp.candidates,
        generatedSql: resp.generated_sql,
        llmEnabled: resp.llm_enabled,
        costMs: resp.cost_ms,
      })
    }
  } catch (e) {
    messages.value.push({
      role: 'assistant',
      question: q,
      answer: `出错了：${e instanceof ApiError ? e.message : '请求失败，请稍后重试'}`,
    })
  } finally {
    sending.value = false
    scrollToBottom()
  }
}
</script>

<template>
  <el-card shadow="never" class="chat-card" :body-style="{ height: '100%', display: 'flex', flexDirection: 'column' }">
    <!-- 工具条 -->
    <div class="chat-toolbar">
      <el-switch v-model="autoRun" active-text="自动执行 SQL" />
      <template v-if="autoRun">
        <el-input-number
          v-model="runConfig.datasource_id"
          :min="1"
          placeholder="数据源 ID"
          style="width: 150px"
        />
        <el-input-number
          v-model="runConfig.max_rows"
          :min="1"
          :max="100"
          placeholder="最大行数"
          style="width: 130px"
        />
      </template>
      <span class="toolbar-tip">候选对象来自元数据与指标语义检索，全程留审计</span>
    </div>

    <!-- 聊天区 -->
    <div ref="chatBody" class="chat-body">
      <el-empty v-if="messages.length === 0" description="输入自然语言问题，例如：昨日订单量是多少？" />

      <div v-for="(m, i) in messages" :key="i" class="msg" :class="m.role">
        <div class="bubble">
          <div v-if="m.role === 'user'" class="msg-text">{{ m.question }}</div>

          <template v-else>
            <div class="msg-text">{{ m.answer }}</div>

            <div v-if="m.generatedSql" class="sql-block">
              <div class="sql-label">生成 SQL（{{ m.llmEnabled ? 'LLM' : '规则' }}）</div>
              <pre class="sql-pre">{{ m.generatedSql }}</pre>
            </div>

            <div v-if="m.candidates && m.candidates.length" class="candidates">
              <div class="sql-label">候选对象</div>
              <el-tag
                v-for="(c, ci) in m.candidates"
                :key="ci"
                size="small"
                :type="c.type === 'metric' ? 'warning' : 'primary'"
                class="candidate-tag"
              >
                [{{ c.type }}] {{ c.name }}
              </el-tag>
            </div>

            <!-- 执行结果表格（动态列） -->
            <template v-if="m.executed">
              <el-alert
                v-if="m.runError"
                :title="`执行失败：${m.runError}`"
                type="error"
                :closable="false"
                show-icon
                class="run-alert"
              />
              <el-table v-else-if="m.rows && m.rows.length" :data="m.rows" size="small" border max-height="320" class="run-table">
                <el-table-column
                  v-for="col in m.columns"
                  :key="col"
                  :prop="col"
                  :label="col"
                  min-width="110"
                  show-overflow-tooltip
                />
              </el-table>
              <el-alert
                v-else-if="!m.runError"
                title="SQL 执行成功，无返回行"
                type="success"
                :closable="false"
                show-icon
                class="run-alert"
              />
              <div v-if="m.rowCount !== undefined" class="run-meta">
                共 {{ m.rowCount }} 行 · 耗时 {{ m.costMs }}ms
              </div>
            </template>
            <div v-else-if="m.costMs !== undefined" class="run-meta">耗时 {{ m.costMs }}ms</div>
          </template>
        </div>
      </div>

      <div v-if="sending" class="msg assistant">
        <div class="bubble typing">思考中…</div>
      </div>
    </div>

    <!-- 输入区 -->
    <div class="chat-input">
      <el-input
        v-model="question"
        type="textarea"
        :rows="2"
        resize="none"
        placeholder="输入你的问题，Enter 发送（Shift+Enter 换行）"
        @keydown.enter.exact.prevent="onSend"
      />
      <el-button
        type="primary"
        :loading="sending"
        :disabled="!question.trim()"
        class="send-btn"
        @click="onSend"
      >
        发送
      </el-button>
    </div>
  </el-card>
</template>

<style scoped>
.chat-card {
  height: calc(100vh - 88px - 32px);
  min-height: 480px;
}
.chat-toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}
.toolbar-tip {
  font-size: 12px;
  color: #c0c4cc;
}
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 4px;
}
.msg {
  display: flex;
  margin-bottom: 14px;
}
.msg.user {
  justify-content: flex-end;
}
.msg.assistant {
  justify-content: flex-start;
}
.bubble {
  max-width: 82%;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.6;
}
.msg.user .bubble {
  background: #409eff;
  color: #fff;
  border-top-right-radius: 2px;
}
.msg.assistant .bubble {
  background: #f4f4f5;
  color: #303133;
  border-top-left-radius: 2px;
}
.bubble.typing {
  color: #909399;
}
.sql-block {
  margin-top: 10px;
}
.sql-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
.sql-pre {
  margin: 0;
  padding: 8px 10px;
  background: #282c34;
  color: #98c379;
  border-radius: 6px;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.candidates {
  margin-top: 10px;
}
.candidate-tag {
  margin: 0 6px 6px 0;
}
.run-alert {
  margin-top: 10px;
}
.run-table {
  margin-top: 10px;
}
.run-meta {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}
.chat-input {
  display: flex;
  gap: 10px;
  align-items: flex-end;
  padding-top: 10px;
  border-top: 1px solid #ebeef5;
}
.send-btn {
  height: 54px;
}
</style>
