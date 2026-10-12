<script setup lang="ts">
// 组件 name 与动态路由 name 一致（keep-alive include 匹配用）
defineOptions({ name: 'SystemMenu' })
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  menuTree,
  saveMenu,
  deleteMenu,
  type SysMenuInfo,
  type SysMenuSaveReq,
} from '@/api/system'
import { ApiError } from '@/api/http'

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

// ---------- 菜单树 ----------
const loading = ref(false)
const tree = ref<SysMenuInfo[]>([])

async function fetchTree() {
  loading.value = true
  try {
    const resp = await menuTree()
    tree.value = resp.list || []
  } catch (e) {
    ElMessage.error(errText(e, '获取菜单树失败'))
  } finally {
    loading.value = false
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

const form = reactive({
  parent_id: 0,
  menu_type: 2,
  name: '',
  path: '',
  component: '',
  perm: '',
  icon: '',
  sort: 0,
  visible: 1,
})

const typeOptions = [
  { value: 1, label: '目录' },
  { value: 2, label: '菜单' },
  { value: 3, label: '按钮' },
]

/** 父节点选项：排除按钮（不能作父级）与自身子树（防环），顶层补“根目录” */
function pruneForParent(nodes: SysMenuInfo[], excludeId: number | null): SysMenuInfo[] {
  const out: SysMenuInfo[] = []
  for (const n of nodes) {
    if (n.menu_type === 3) continue
    if (excludeId !== null && n.id === excludeId) continue
    const children = pruneForParent(n.children || [], excludeId)
    out.push(children.length ? { ...n, children } : { ...n, children: undefined })
  }
  return out
}

const parentOptions = computed<SysMenuInfo[]>(() => [
  {
    id: 0,
    name: '根目录',
    menu_type: 1,
    children: pruneForParent(tree.value, editingId.value),
  },
])

function openCreate(parent?: SysMenuInfo) {
  editingId.value = null
  Object.assign(form, {
    parent_id: parent?.id ?? 0,
    // 按钮父必须是菜单，目录下挂菜单
    menu_type: parent?.menu_type === 2 ? 3 : 2,
    name: '',
    path: '',
    component: '',
    perm: '',
    icon: '',
    sort: 0,
    visible: 1,
  })
  dialogVisible.value = true
}

function openEdit(row: SysMenuInfo) {
  editingId.value = row.id ?? null
  Object.assign(form, {
    parent_id: row.parent_id ?? 0,
    menu_type: row.menu_type ?? 2,
    name: row.name || '',
    path: row.path || '',
    component: row.component || '',
    perm: row.perm || '',
    icon: row.icon || '',
    sort: row.sort ?? 0,
    visible: row.visible === 0 ? 0 : 1,
  })
  dialogVisible.value = true
}

/** 按菜单类型组装保存载荷（更新为整存，不相关字段显式置空防脏数据残留） */
async function onSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  if (form.menu_type !== 3 && !form.path.trim()) {
    ElMessage.warning('请输入路由路径')
    return
  }
  if (form.menu_type === 3 && !form.perm.trim()) {
    ElMessage.warning('请输入权限标识（如 sys:user:create）')
    return
  }
  const payload: SysMenuSaveReq = {
    id: editingId.value ?? 0,
    parent_id: form.parent_id || 0,
    name: form.name.trim(),
    menu_type: form.menu_type,
    sort: form.sort,
    visible: form.menu_type === 3 ? undefined : form.visible,
  }
  if (form.menu_type === 1) {
    payload.path = form.path.trim()
    payload.icon = form.icon.trim()
    payload.component = ''
    payload.perm = ''
  } else if (form.menu_type === 2) {
    payload.path = form.path.trim()
    payload.component = form.component.trim()
    payload.icon = form.icon.trim()
    payload.perm = ''
  } else {
    payload.path = ''
    payload.component = ''
    payload.icon = ''
    payload.perm = form.perm.trim()
  }
  saving.value = true
  try {
    await saveMenu(payload)
    ElMessage.success(editingId.value === null ? '菜单创建成功' : '菜单更新成功')
    dialogVisible.value = false
    fetchTree()
  } catch (e) {
    ElMessage.error(errText(e, '保存菜单失败'))
  } finally {
    saving.value = false
  }
}

// ---------- 删除 ----------
async function onDelete(row: SysMenuInfo) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除菜单「${row.name}」吗？若存在子菜单或已被角色引用，删除将被拒绝。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteMenu(row.id)
    ElMessage.success('删除成功')
    fetchTree()
  } catch (e) {
    ElMessage.error(errText(e, '删除菜单失败'))
  }
}

onMounted(fetchTree)
</script>

<template>
  <div>
    <el-card shadow="never">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-button v-perm="['sys:menu:create']" type="primary" @click="openCreate()">
          <el-icon><Plus /></el-icon>&nbsp;新建菜单
        </el-button>
        <el-button @click="fetchTree">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
      </div>

      <!-- 树形表格 -->
      <el-table
        v-loading="loading"
        :data="tree"
        row-key="id"
        :tree-props="{ children: 'children' }"
        default-expand-all
        stripe
        border
      >
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <el-icon v-if="row.icon" class="menu-icon"><component :is="row.icon" /></el-icon>
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag v-if="typeTag[row.menu_type]" :type="typeTag[row.menu_type].type" size="small">
              {{ typeTag[row.menu_type].label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="perm" label="权限标识" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.perm || '-' }}</template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.path || '-' }}</template>
        </el-table-column>
        <el-table-column prop="component" label="组件" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.component || '-' }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="可见" width="80">
          <template #default="{ row }">
            <span v-if="row.menu_type === 3" class="muted">-</span>
            <el-tag v-else :type="row.visible === 1 ? 'success' : 'info'" size="small">
              {{ row.visible === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.menu_type !== 3"
              v-perm="['sys:menu:create']"
              link
              type="success"
              @click="openCreate(row)"
            >
              子级
            </el-button>
            <el-button v-perm="['sys:menu:update']" link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button v-perm="['sys:menu:delete']" link type="danger" @click="onDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="editingId === null ? '新建菜单' : '编辑菜单'"
      width="560px"
      destroy-on-close
    >
      <el-form :model="form" label-width="90px">
        <el-form-item label="父级">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            node-key="id"
            check-strictly
            default-expand-all
            :render-after-expand="false"
            :props="{ label: 'name', children: 'children' }"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-radio-group v-model="form.menu_type">
            <el-radio-button v-for="t in typeOptions" :key="t.value" :value="t.value">
              {{ t.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="菜单名称" maxlength="64" />
        </el-form-item>
        <el-form-item v-if="form.menu_type !== 3" label="路由路径" required>
          <el-input
            v-model="form.path"
            :placeholder="form.menu_type === 1 ? '以 / 开头，如 /system' : '相对父级，如 user'"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item v-if="form.menu_type === 2" label="组件">
          <el-input
            v-model="form.component"
            placeholder="相对 src/views 的路径（不含 .vue），如 system/UserList"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item v-if="form.menu_type === 3" label="权限标识" required>
          <el-input v-model="form.perm" placeholder="模块:动作，如 sys:user:create" maxlength="64" />
        </el-form-item>
        <el-form-item v-if="form.menu_type !== 3" label="图标">
          <el-input v-model="form.icon" placeholder="Element Plus 图标名，如 Setting" maxlength="32" />
        </el-form-item>
        <el-form-item v-if="form.menu_type !== 3" label="是否可见">
          <el-switch v-model="form.visible" :active-value="1" :inactive-value="0" />
          <span class="form-tip">隐藏后仍注册路由，仅不出现在侧边栏</span>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
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
.menu-icon {
  vertical-align: -2px;
  margin-right: 6px;
}
.muted {
  color: #909399;
}
.form-tip {
  margin-left: 10px;
  font-size: 12px;
  color: #909399;
}
</style>
