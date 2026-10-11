<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type TreeInstance } from 'element-plus'
import {
  listRoles,
  getRole,
  createRole,
  updateRole,
  deleteRole,
  menuTree,
  type SysRoleInfo,
  type SysRoleSaveReq,
  type SysMenuInfo,
} from '@/api/system'
import { ApiError } from '@/api/http'

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

// ---------- 列表 ----------
const loading = ref(false)
const list = ref<SysRoleInfo[]>([])
const total = ref(0)
const query = reactive({ keyword: '', page: 1, size: 10 })

async function fetchList() {
  loading.value = true
  try {
    const resp = await listRoles({
      keyword: query.keyword || undefined,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取角色列表失败'))
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 菜单权限树 ----------
const treeRef = ref<TreeInstance>()
const menuTreeData = ref<SysMenuInfo[]>([])
const treeLoading = ref(false)

async function fetchMenuTree() {
  if (menuTreeData.value.length) return
  treeLoading.value = true
  try {
    const resp = await menuTree()
    menuTreeData.value = resp.list || []
  } catch (e) {
    ElMessage.error(errText(e, '获取菜单树失败'))
  } finally {
    treeLoading.value = false
  }
}

const typeTag: Record<number, { label: string; type: string }> = {
  1: { label: '目录', type: 'primary' },
  2: { label: '菜单', type: 'success' },
  3: { label: '按钮', type: 'warning' },
}

// ---------- 新建 / 编辑 ----------
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({
  role_key: '',
  role_name: '',
  status: 1,
  remark: '',
})

const rules: FormRules = {
  role_key: [
    { required: true, message: '请输入权限字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]{1,31}$/, message: '字母开头，2-32 位字母/数字/下划线', trigger: 'blur' },
  ],
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

async function openCreate() {
  editingId.value = null
  Object.assign(form, { role_key: '', role_name: '', status: 1, remark: '' })
  dialogVisible.value = true
  await fetchMenuTree()
  await nextTick()
  treeRef.value?.setCheckedKeys([])
}

/** 编辑：表单取列表行，菜单勾选经 getRole 详情回显（列表行不一定带 menu_ids） */
async function openEdit(row: SysRoleInfo) {
  if (row.id === undefined) return
  editingId.value = row.id
  let detail: SysRoleInfo = row
  try {
    const resp = await getRole(row.id)
    if (resp.role) detail = resp.role
  } catch {
    /* 详情失败时退回列表行数据 */
  }
  Object.assign(form, {
    role_key: detail.role_key || '',
    role_name: detail.role_name || '',
    status: detail.status ?? 1,
    remark: detail.remark || '',
  })
  dialogVisible.value = true
  await fetchMenuTree()
  await nextTick()
  treeRef.value?.setCheckedKeys((detail.menu_ids || []).map(Number))
}

async function onSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  const menuIds = (treeRef.value?.getCheckedKeys() || []).map(Number)
  const payload: SysRoleSaveReq = {
    id: editingId.value ?? 0,
    role_key: form.role_key.trim(),
    role_name: form.role_name.trim(),
    status: form.status,
    remark: form.remark.trim() || undefined,
    menu_ids: menuIds,
  }
  saving.value = true
  try {
    if (editingId.value === null) {
      await createRole(payload)
      ElMessage.success('角色创建成功')
    } else {
      await updateRole(payload)
      ElMessage.success('角色更新成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '保存角色失败'))
  } finally {
    saving.value = false
  }
}

// ---------- 删除 ----------
async function onDelete(row: SysRoleInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除角色「${row.role_name || row.role_key}」吗？若仍有用户引用该角色，删除将被拒绝。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(errText(e, '删除角色失败'))
  }
}

onMounted(fetchList)
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          placeholder="按权限字符/角色名搜索"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" plain @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button v-perm="['sys:role:create']" type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>&nbsp;新建角色
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="role_key" label="权限字符" min-width="130" show-overflow-tooltip />
        <el-table-column prop="role_name" label="角色名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button v-perm="['sys:role:update']" link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button v-perm="['sys:role:delete']" link type="danger" @click="onDelete(row)">
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
      :title="editingId === null ? '新建角色' : '编辑角色'"
      width="620px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="权限字符" prop="role_key">
          <el-input
            v-model="form.role_key"
            placeholder="如 operator"
            maxlength="32"
            :disabled="editingId !== null"
          />
          <div v-if="editingId !== null" class="form-tip">权限字符创建后不可修改</div>
        </el-form-item>
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="form.role_name" placeholder="如 运营人员" maxlength="64" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
        <el-form-item label="菜单权限" prop="menu_ids">
          <div v-loading="treeLoading" class="menu-tree-box">
            <el-tree
              ref="treeRef"
              :data="menuTreeData"
              node-key="id"
              show-checkbox
              check-strictly
              default-expand-all
              :props="{ label: 'name', children: 'children' }"
            >
              <template #default="{ data }">
                <span class="tree-node">
                  {{ data.name }}
                  <el-tag v-if="typeTag[data.menu_type]" size="small" :type="typeTag[data.menu_type].type">
                    {{ typeTag[data.menu_type].label }}
                  </el-tag>
                  <span v-if="data.perm" class="tree-perm">{{ data.perm }}</span>
                </span>
              </template>
            </el-tree>
          </div>
          <div class="form-tip">严格模式：父子独立勾选，仅勾选节点会保存为角色菜单</div>
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
.menu-tree-box {
  width: 100%;
  max-height: 320px;
  overflow: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 6px;
}
.tree-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.tree-perm {
  font-size: 12px;
  color: #909399;
}
.form-tip {
  width: 100%;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
