<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  listUsers,
  createUser,
  updateUser,
  deleteUser,
  resetUserPwd,
  listRoles,
  type SysUserInfo,
  type SysRoleInfo,
} from '@/api/system'
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
const list = ref<SysUserInfo[]>([])
const total = ref(0)
// status：-1 全部（后端 >=0 才参与过滤，默认必须传 -1，否则会只查出停用用户）
const query = reactive({ keyword: '', status: -1, page: 1, size: 10 })

const statusOptions = [
  { value: -1, label: '全部状态' },
  { value: 1, label: '启用' },
  { value: 0, label: '停用' },
]

async function fetchList() {
  loading.value = true
  try {
    const resp = await listUsers({
      keyword: query.keyword || undefined,
      status: query.status,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取用户列表失败'))
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 角色映射（role_ids → 角色名） ----------
const roleOptions = ref<SysRoleInfo[]>([])
const roleMap = computed(() => {
  const m = new Map<number, string>()
  for (const r of roleOptions.value) {
    if (r.id !== undefined) m.set(r.id, r.role_name || r.role_key || `角色#${r.id}`)
  }
  return m
})

function roleNames(row: SysUserInfo): string[] {
  return (row.role_ids || []).map((id) => roleMap.value.get(id) ?? `#${id}`)
}

async function fetchRoleOptions() {
  try {
    const resp = await listRoles({ page: 1, size: 200 })
    roleOptions.value = resp.list || []
  } catch {
    /* 角色下拉拉取失败不阻断页面 */
  }
}

// ---------- 状态开关（整存更新：带全字段防误清空） ----------
async function onStatusChange(row: SysUserInfo) {
  if (row.id === undefined) return
  const prev = row.status === 1 ? 0 : 1
  try {
    await updateUser({
      id: row.id,
      nickname: row.nickname || '',
      mobile: row.mobile || '',
      email: row.email || '',
      remark: row.remark || '',
      status: row.status,
      role_ids: row.role_ids || [],
    })
    ElMessage.success(row.status === 1 ? '已启用' : '已停用')
  } catch (e) {
    row.status = prev
    ElMessage.error(errText(e, '状态更新失败'))
  }
}

// ---------- 新建 / 编辑 ----------
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  mobile: '',
  email: '',
  remark: '',
  status: 1,
  role_ids: [] as number[],
})

// 编辑时密码非必填（留空不改）
const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: '请输入登录名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]{2,32}$/, message: '2-32 位字母/数字/下划线', trigger: 'blur' },
  ],
  password:
    editingId.value === null
      ? [{ required: true, message: '请输入初始密码', trigger: 'blur' }]
      : [{ pattern: /^\S{6,32}$/, message: '6-32 位且不含空格，留空则不修改', trigger: 'blur' }],
  mobile: [{ pattern: /^1\d{10}$/, message: '手机号格式不正确', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
}))

function openCreate() {
  editingId.value = null
  Object.assign(form, {
    username: '',
    password: '',
    nickname: '',
    mobile: '',
    email: '',
    remark: '',
    status: 1,
    role_ids: [],
  })
  dialogVisible.value = true
}

function openEdit(row: SysUserInfo) {
  editingId.value = row.id ?? null
  Object.assign(form, {
    username: row.username || '',
    password: '',
    nickname: row.nickname || '',
    mobile: row.mobile || '',
    email: row.email || '',
    remark: row.remark || '',
    status: row.status ?? 1,
    role_ids: [...(row.role_ids || [])],
  })
  dialogVisible.value = true
}

async function onSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    if (editingId.value === null) {
      await createUser({
        username: form.username.trim(),
        password: form.password,
        nickname: form.nickname.trim() || undefined,
        mobile: form.mobile.trim() || undefined,
        email: form.email.trim() || undefined,
        remark: form.remark.trim() || undefined,
        role_ids: form.role_ids,
      })
      ElMessage.success('用户创建成功')
    } else {
      // 密码留空 = 不修改；username 后端不支持改，不提交
      await updateUser({
        id: editingId.value,
        nickname: form.nickname.trim(),
        mobile: form.mobile.trim(),
        email: form.email.trim(),
        status: form.status,
        remark: form.remark.trim(),
        password: form.password || undefined,
        role_ids: form.role_ids,
      })
      ElMessage.success('用户更新成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '保存用户失败'))
  } finally {
    saving.value = false
  }
}

// ---------- 重置密码 ----------
async function onResetPwd(row: SysUserInfo) {
  if (row.id === undefined) return
  try {
    const { value } = await ElMessageBox.prompt(
      `为用户「${row.username || row.id}」设置新密码（6-32 位，不含空格）`,
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'password',
        inputPattern: /^\S{6,32}$/,
        inputErrorMessage: '密码须为 6-32 位且不含空格',
      },
    )
    await resetUserPwd(row.id, value)
    ElMessage.success('密码已重置')
  } catch (e) {
    if (typeof e === 'string') return // cancel / close
    ElMessage.error(errText(e, '重置密码失败'))
  }
}

// ---------- 删除 ----------
async function onDelete(row: SysUserInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除用户「${row.username}」吗？删除后不可恢复。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '删除用户失败'))
  }
}

onMounted(() => {
  fetchList()
  fetchRoleOptions()
})
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-select v-model="query.status" style="width: 130px" @change="onSearch">
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-input
          v-model="query.keyword"
          placeholder="用户名/昵称/手机号"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" plain @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button v-perm="['sys:user:create']" type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建用户
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="登录名" min-width="120" show-overflow-tooltip />
        <el-table-column prop="nickname" label="昵称" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="mobile" label="手机号" width="130">
          <template #default="{ row }">{{ row.mobile || '-' }}</template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.email || '-' }}</template>
        </el-table-column>
        <el-table-column label="角色" min-width="150">
          <template #default="{ row }">
            <template v-if="roleNames(row).length">
              <el-tag v-for="n in roleNames(row)" :key="n" size="small" class="role-tag">
                {{ n }}
              </el-tag>
            </template>
            <span v-else class="muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch
              v-perm="['sys:user:update']"
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="onStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button v-perm="['sys:user:update']" link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button
              v-perm="['sys:user:resetPwd']"
              link
              type="warning"
              @click="onResetPwd(row)"
            >
              重置密码
            </el-button>
            <el-button v-perm="['sys:user:delete']" link type="danger" @click="onDelete(row)">
              删除
            </el-button>
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
      :title="editingId === null ? '新建用户' : '编辑用户'"
      width="560px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="登录名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="字母/数字/下划线"
            maxlength="32"
            :disabled="editingId !== null"
          />
          <div v-if="editingId !== null" class="form-tip">登录名创建后不可修改</div>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            autocomplete="new-password"
            :placeholder="editingId === null ? '请输入初始密码' : '留空则不修改'"
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" maxlength="64" />
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form.mobile" maxlength="11" placeholder="11 位手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" maxlength="64" />
        </el-form-item>
        <el-form-item label="角色" prop="role_ids">
          <el-select
            v-model="form.role_ids"
            multiple
            filterable
            style="width: 100%"
            placeholder="选择角色（可多选）"
          >
            <el-option
              v-for="r in roleOptions"
              :key="r.id"
              :label="r.role_name || r.role_key"
              :value="r.id as number"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editingId !== null" label="状态" prop="status">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
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
  flex-wrap: wrap;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
.role-tag {
  margin-right: 6px;
}
.muted {
  color: #909399;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}
</style>
