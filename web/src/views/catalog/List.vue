<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  createCategory,
  categoryList,
  tagAsset,
  assetList,
  type AssetCategory,
  type Asset,
} from '@/api/catalog'
import { ApiError } from '@/api/http'

// ---------- 分类树 ----------
interface CategoryNode {
  id: number
  parent_id: number
  name: string
  children: CategoryNode[]
}

const treeLoading = ref(false)
const treeData = ref<CategoryNode[]>([])
const currentCategory = ref<CategoryNode | null>(null)

/** 扁平分类列表 → 树；parent 不在列表内的节点挂到根级 */
function buildTree(list: AssetCategory[]): CategoryNode[] {
  const nodes = new Map<number, CategoryNode>()
  const items = list.filter((c) => c.id !== undefined)
  items.forEach((c) => {
    nodes.set(c.id!, {
      id: c.id!,
      parent_id: c.parent_id || 0,
      name: c.name || `分类${c.id}`,
      children: [],
    })
  })
  const roots: CategoryNode[] = []
  nodes.forEach((node) => {
    const parent = node.parent_id > 0 ? nodes.get(node.parent_id) : undefined
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  })
  return roots
}

async function fetchTree() {
  treeLoading.value = true
  try {
    const resp = await categoryList()
    treeData.value = buildTree(resp.list || [])
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取分类失败')
  } finally {
    treeLoading.value = false
  }
}

function onNodeClick(data: CategoryNode) {
  currentCategory.value = data
  query.category_id = data.id
  onSearch()
}

function onResetFilter() {
  currentCategory.value = null
  query.category_id = undefined
  onSearch()
}

// ---------- 资产列表 ----------
const loading = ref(false)
const list = ref<Asset[]>([])
const total = ref(0)
const query = reactive({
  category_id: undefined as number | undefined,
  keyword: '',
  page: 1,
  size: 10,
})

/** 分类 id → 名称（表格列展示用） */
const categoryName = computed(() => {
  const map = new Map<number, string>()
  const walk = (nodes: CategoryNode[]) => {
    nodes.forEach((n) => {
      map.set(n.id, n.name)
      walk(n.children)
    })
  }
  walk(treeData.value)
  return map
})

const classText: Record<string, string> = {
  L1: 'L1 公开',
  L2: 'L2 内部',
  L3: 'L3 敏感',
  L4: 'L4 核心',
}

async function fetchList() {
  loading.value = true
  try {
    const resp = await assetList({
      category_id: query.category_id,
      keyword: query.keyword.trim() || undefined,
      page: query.page,
      size: query.size,
    })
    list.value = resp.list || []
    total.value = resp.total || 0
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '获取资产列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  fetchList()
}

// ---------- 新建分类对话框 ----------
const cateVisible = ref(false)
const cateSaving = ref(false)
const cateFormEl = ref<FormInstance>()
const cateForm = reactive({ parent_id: undefined as number | undefined, name: '' })

const cateRules = reactive<FormRules>({
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
})

function openCategory(parent?: CategoryNode) {
  Object.assign(cateForm, { parent_id: parent?.id, name: '' })
  cateVisible.value = true
}

async function onSaveCategory() {
  const valid = await cateFormEl.value?.validate().catch(() => false)
  if (!valid) return
  cateSaving.value = true
  try {
    await createCategory({
      parent_id: cateForm.parent_id || 0,
      name: cateForm.name.trim(),
    })
    ElMessage.success('分类创建成功')
    cateVisible.value = false
    fetchTree()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '创建分类失败')
  } finally {
    cateSaving.value = false
  }
}

// ---------- 登记资产对话框 ----------
const assetVisible = ref(false)
const assetSaving = ref(false)
const assetFormEl = ref<FormInstance>()
const assetForm = reactive({
  entity_id: undefined as number | undefined,
  category_id: undefined as number | undefined,
  name: '',
  entity_type: 'table',
  classification: 'L2',
  visibility: 'public',
  comment: '',
})

const assetRules = reactive<FormRules>({
  entity_id: [{ required: true, message: '请输入资产实体 ID（元数据实体）', trigger: 'blur' }],
  name: [{ required: true, message: '请输入资产名称', trigger: 'blur' }],
  classification: [{ required: true, message: '请选择安全分级', trigger: 'change' }],
  visibility: [{ required: true, message: '请选择可见性', trigger: 'change' }],
})

function openAsset() {
  Object.assign(assetForm, {
    entity_id: undefined,
    category_id: currentCategory.value?.id,
    name: '',
    entity_type: 'table',
    classification: 'L2',
    visibility: 'public',
    comment: '',
  })
  assetVisible.value = true
}

async function onSaveAsset() {
  const valid = await assetFormEl.value?.validate().catch(() => false)
  if (!valid) return
  assetSaving.value = true
  try {
    await tagAsset({
      entity_id: Number(assetForm.entity_id),
      category_id: assetForm.category_id || undefined,
      name: assetForm.name.trim(),
      entity_type: assetForm.entity_type.trim() || undefined,
      classification: assetForm.classification,
      visibility: assetForm.visibility,
      comment: assetForm.comment.trim() || undefined,
    })
    ElMessage.success('资产登记成功')
    assetVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '登记资产失败')
  } finally {
    assetSaving.value = false
  }
}

onMounted(() => {
  fetchTree()
  fetchList()
})
</script>

<template>
  <div class="catalog-page">
    <!-- 左侧分类树 -->
    <el-card shadow="never" class="tree-card" v-loading="treeLoading">
      <template #header>
        <div class="tree-header">
          <span>资产分类</span>
          <div>
            <el-button link type="primary" @click="openCategory()">新建分类</el-button>
          </div>
        </div>
      </template>
      <el-button
        size="small"
        class="all-btn"
        :type="currentCategory === null ? 'primary' : ''"
        @click="onResetFilter"
      >
        全部资产
      </el-button>
      <el-tree
        :data="treeData"
        :props="{ label: 'name', children: 'children' }"
        node-key="id"
        highlight-current
        default-expand-all
        :expand-on-click-node="false"
        empty-text="暂无分类，点击右上角新建"
        @node-click="onNodeClick"
      >
        <template #default="{ data }">
          <span class="tree-node">
            <el-icon><Folder /></el-icon>
            <span class="tree-name">{{ data.name }}</span>
            <el-button
              link
              type="primary"
              size="small"
              class="tree-add"
              @click.stop="openCategory(data)"
            >
              <el-icon><Plus /></el-icon>
            </el-button>
          </span>
        </template>
      </el-tree>
    </el-card>

    <!-- 右侧资产列表 -->
    <el-card shadow="never" class="table-card">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          placeholder="按资产名称搜索"
          clearable
          style="width: 220px"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-button type="primary" @click="onSearch">
          <el-icon><Search /></el-icon>&nbsp;查询
        </el-button>
        <el-button type="success" @click="openAsset">
          <el-icon><Plus /></el-icon>&nbsp;登记资产
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>&nbsp;刷新
        </el-button>
        <span class="filter-tip">
          {{ currentCategory ? `分类：${currentCategory.name}` : '全部分类' }}
        </span>
      </div>

      <el-table v-loading="loading" :data="list" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="资产名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="entity_id" label="实体 ID" width="90" align="center" />
        <el-table-column prop="entity_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.entity_type || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="所属分类" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ categoryName.get(row.category_id) || row.category_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="安全分级" width="110" align="center">
          <template #default="{ row }">
            <el-tag
              v-if="row.classification"
              size="small"
              :type="row.classification === 'L4' ? 'danger' : row.classification === 'L3' ? 'warning' : 'success'"
            >
              {{ classText[row.classification] || row.classification }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="可见性" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.visibility === 'private' ? 'danger' : 'success'">
              {{ row.visibility === 'private' ? '私有' : '公开' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="comment" label="备注" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.comment || '-' }}</template>
        </el-table-column>
      </el-table>

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

    <!-- 新建分类对话框 -->
    <el-dialog v-model="cateVisible" title="新建分类" width="460px" destroy-on-close>
      <el-form ref="cateFormEl" :model="cateForm" :rules="cateRules" label-width="90px">
        <el-form-item label="上级分类">
          <el-tree-select
            v-model="cateForm.parent_id"
            :data="treeData"
            :props="{ label: 'name', children: 'children', value: 'id' }"
            node-key="id"
            value-key="id"
            check-strictly
            clearable
            placeholder="不选则为顶级分类"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="cateForm.name" placeholder="如：交易域 / 用户域" maxlength="64" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cateVisible = false">取消</el-button>
        <el-button type="primary" :loading="cateSaving" @click="onSaveCategory">保存</el-button>
      </template>
    </el-dialog>

    <!-- 登记资产对话框 -->
    <el-dialog v-model="assetVisible" title="登记资产" width="560px" destroy-on-close>
      <el-form ref="assetFormEl" :model="assetForm" :rules="assetRules" label-width="100px">
        <el-form-item label="实体 ID" prop="entity_id">
          <el-input-number
            v-model="assetForm.entity_id"
            :min="1"
            placeholder="元数据实体 ID（/metadata 实体）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="资产名称" prop="name">
          <el-input v-model="assetForm.name" placeholder="如：订单明细表" maxlength="128" />
        </el-form-item>
        <el-form-item label="所属分类">
          <el-tree-select
            v-model="assetForm.category_id"
            :data="treeData"
            :props="{ label: 'name', children: 'children', value: 'id' }"
            node-key="id"
            value-key="id"
            check-strictly
            clearable
            placeholder="选择分类（可空）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="资产类型">
          <el-select v-model="assetForm.entity_type" style="width: 100%">
            <el-option label="表 table" value="table" />
            <el-option label="报表 dashboard" value="dashboard" />
            <el-option label="接口 api" value="api" />
            <el-option label="指标 metric" value="metric" />
          </el-select>
        </el-form-item>
        <el-form-item label="安全分级" prop="classification">
          <el-select v-model="assetForm.classification" style="width: 100%">
            <el-option label="L1 公开数据" value="L1" />
            <el-option label="L2 内部数据" value="L2" />
            <el-option label="L3 敏感数据" value="L3" />
            <el-option label="L4 核心数据" value="L4" />
          </el-select>
        </el-form-item>
        <el-form-item label="可见性" prop="visibility">
          <el-radio-group v-model="assetForm.visibility">
            <el-radio value="public">公开</el-radio>
            <el-radio value="private">私有</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="assetForm.comment" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assetVisible = false">取消</el-button>
        <el-button type="primary" :loading="assetSaving" @click="onSaveAsset">登记</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.catalog-page {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}
.tree-card {
  width: 280px;
  flex-shrink: 0;
}
.tree-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.all-btn {
  margin-bottom: 8px;
  width: 100%;
}
.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}
.tree-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tree-add {
  margin-left: auto;
  visibility: hidden;
}
.tree-node:hover .tree-add {
  visibility: visible;
}
.table-card {
  flex: 1;
  min-width: 0;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
  align-items: center;
}
.filter-tip {
  margin-left: auto;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
</style>
