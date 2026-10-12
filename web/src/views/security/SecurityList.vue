<script setup lang="ts">
// 组件 name 与动态路由 name 一致（菜单 /security/security → routerName = SecuritySecurity，keep-alive include 匹配用）
defineOptions({ name: 'SecuritySecurity' })
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  saveSecLevel,
  secLevelList,
  tagSecEntity,
  secEntityTagList,
  saveSecMaskRule,
  deleteSecMaskRule,
  secMaskRuleList,
  maskPreview,
  saveSecGrant,
  deleteSecGrant,
  secGrantList,
  secCheckPerm,
  type SecLevelInfo,
  type SecEntityTag,
  type SecMaskRule,
  type SecGrant,
  type MaskPreviewResp,
  type SecCheckPermResp,
} from '@/api/security'
import { ApiError } from '@/api/http'

// ---------- 公共 ----------
const activeTab = ref('level')

function errText(e: unknown, fallback: string) {
  return e instanceof ApiError ? e.message : fallback
}

function fmtTime(ts?: number) {
  if (!ts) return '-'
  return new Date(ts * (ts < 1e12 ? 1000 : 1)).toLocaleString('zh-CN', { hour12: false })
}

/** 分级三色 tag：L1 公开=绿 / L2 内部=蓝 / L3 敏感=黄 / L4 机密=红，其余灰 */
const levelTagMap: Record<string, 'success' | 'primary' | 'warning' | 'danger'> = {
  L1: 'success',
  L2: 'primary',
  L3: 'warning',
  L4: 'danger',
}
function levelTagOf(code?: string): 'success' | 'primary' | 'warning' | 'danger' | 'info' {
  return (code && levelTagMap[code]) || 'info'
}

/** 分级目录全量（页签 1 表格 + 打标/授权下拉 共用；首次 List 后端自动播种 L1-L4） */
const levelList = ref<SecLevelInfo[]>([])
async function fetchLevels() {
  try {
    const resp = await secLevelList()
    levelList.value = resp.list || []
  } catch (e) {
    ElMessage.error(errText(e, '获取分级目录失败'))
  }
}

/** 脱敏规则下拉源（打标对话框引用），size 放大一次拉全 */
const ruleOptions = ref<SecMaskRule[]>([])
async function fetchRuleOptions() {
  try {
    const resp = await secMaskRuleList({ page: 1, size: 200 })
    ruleOptions.value = resp.list || []
  } catch {
    /* 下拉数据拉取失败不阻断页面 */
  }
}

function onTabChange(name: string | number) {
  const tab = String(name)
  if (tab === 'level') fetchLevels()
  else if (tab === 'entity') {
    if (!levelList.value.length) fetchLevels()
    fetchRuleOptions()
    fetchEntityTags()
  } else if (tab === 'maskrule') fetchRules()
  else if (tab === 'grant') {
    if (!levelList.value.length) fetchLevels()
    fetchGrants()
  }
}

// ============ 页签 1：分级目录 ============
const levelLoading = ref(false)
const levelSaving = ref(false)
const levelDialogVisible = ref(false)
const levelEditingId = ref<number | null>(null)
const levelFormRef = ref<FormInstance>()
const levelForm = reactive({ code: '', name: '', comment: '' })

const levelRules: FormRules = {
  code: [
    { required: true, message: '请输入分级编码', trigger: 'blur' },
    { pattern: /^L[1-9]$/, message: '编码须为 L 加 1-9 数字（如 L1-L4）', trigger: 'blur' },
  ],
  name: [{ required: true, message: '请输入分级名称', trigger: 'blur' }],
}

function openLevelCreate() {
  levelEditingId.value = null
  Object.assign(levelForm, { code: '', name: '', comment: '' })
  levelDialogVisible.value = true
}

function openLevelEdit(row: SecLevelInfo) {
  levelEditingId.value = row.id ?? null
  Object.assign(levelForm, {
    code: row.code || '',
    name: row.name || '',
    comment: row.comment || '',
  })
  levelDialogVisible.value = true
}

async function onSaveLevel() {
  const valid = await levelFormRef.value?.validate().catch(() => false)
  if (!valid) return
  levelSaving.value = true
  try {
    // 后端白名单为大写 L 前缀，小写 l 自动纠正
    await saveSecLevel({
      code: levelForm.code.trim().toUpperCase(),
      name: levelForm.name.trim(),
      comment: levelForm.comment.trim() || undefined,
    })
    ElMessage.success(levelEditingId.value === null ? '分级创建成功' : '分级更新成功')
    levelDialogVisible.value = false
    fetchLevels()
  } catch (e) {
    ElMessage.error(errText(e, '保存分级失败'))
  } finally {
    levelSaving.value = false
  }
}

// ============ 页签 2：实体打标 ============
const etLoading = ref(false)
const etList = ref<SecEntityTag[]>([])
const etTotal = ref(0)
const etQuery = reactive({ level_code: '', page: 1, size: 10 })

async function fetchEntityTags() {
  etLoading.value = true
  try {
    const resp = await secEntityTagList({
      ...etQuery,
      level_code: etQuery.level_code || undefined,
    })
    etList.value = resp.list || []
    etTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取打标列表失败'))
  } finally {
    etLoading.value = false
  }
}

function etSearch() {
  etQuery.page = 1
  fetchEntityTags()
}

const tagDialogVisible = ref(false)
const tagSaving = ref(false)
const tagFormRef = ref<FormInstance>()
const tagForm = reactive({ entity_id: undefined as number | undefined, entity_name: '', level_code: '', mask_rule_code: '' })

const tagRules: FormRules = {
  entity_id: [{ required: true, message: '请输入实体 ID', trigger: 'blur' }],
  level_code: [{ required: true, message: '请选择密级', trigger: 'change' }],
}

function openTagCreate() {
  Object.assign(tagForm, { entity_id: undefined, entity_name: '', level_code: '', mask_rule_code: '' })
  tagDialogVisible.value = true
}

/** 打标为按实体 upsert：回填行数据后可原样覆盖或改绑密级/规则 */
function openTagEdit(row: SecEntityTag) {
  Object.assign(tagForm, {
    entity_id: row.entity_id,
    entity_name: row.entity_name || '',
    level_code: row.level_code || '',
    mask_rule_code: row.mask_rule_code || '',
  })
  tagDialogVisible.value = true
}

async function onSaveTag() {
  const valid = await tagFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!tagForm.entity_id || tagForm.entity_id < 1) {
    ElMessage.warning('实体 ID 须为正整数')
    return
  }
  tagSaving.value = true
  try {
    await tagSecEntity({
      entity_id: tagForm.entity_id,
      entity_name: tagForm.entity_name.trim() || undefined,
      level_code: tagForm.level_code,
      mask_rule_code: tagForm.mask_rule_code || undefined,
    })
    ElMessage.success('打标成功（同实体重复打标为覆盖）')
    tagDialogVisible.value = false
    fetchEntityTags()
  } catch (e) {
    ElMessage.error(errText(e, '打标失败'))
  } finally {
    tagSaving.value = false
  }
}

// ============ 页签 3：脱敏规则 ============
const ruleLoading = ref(false)
const ruleList = ref<SecMaskRule[]>([])
const ruleTotal = ref(0)
const ruleQuery = reactive({ page: 1, size: 10 })

const algorithmOptions = [
  { value: 'mask', label: 'mask（部分掩码）' },
  { value: 'hash', label: 'hash（SHA-256 前 16 位）' },
  { value: 'replace', label: 'replace（整值替换）' },
  { value: 'null', label: 'null（置空）' },
]

async function fetchRules() {
  ruleLoading.value = true
  try {
    const resp = await secMaskRuleList({ ...ruleQuery })
    ruleList.value = resp.list || []
    ruleTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取脱敏规则失败'))
  } finally {
    ruleLoading.value = false
  }
}

const ruleDialogVisible = ref(false)
const ruleSaving = ref(false)
const ruleEditingId = ref<number | null>(null)
const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive({
  rule_code: '',
  rule_name: '',
  algorithm: 'mask',
  // mask 算法参数
  prefix: 3,
  suffix: 4,
  mask_char: '*',
  // replace 算法参数
  replace: '',
})

const ruleRules: FormRules = {
  rule_code: [{ required: true, message: '请输入规则编码', trigger: 'blur' }],
  rule_name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  algorithm: [{ required: true, message: '请选择算法', trigger: 'change' }],
}

function openRuleCreate() {
  ruleEditingId.value = null
  Object.assign(ruleForm, {
    rule_code: '',
    rule_name: '',
    algorithm: 'mask',
    prefix: 3,
    suffix: 4,
    mask_char: '*',
    replace: '',
  })
  ruleDialogVisible.value = true
}

/** 编辑回填：解析 params_json 到动态表单字段，脏数据回退默认值 */
function openRuleEdit(row: SecMaskRule) {
  ruleEditingId.value = row.id ?? null
  Object.assign(ruleForm, {
    rule_code: row.rule_code || '',
    rule_name: row.rule_name || '',
    algorithm: row.algorithm || 'mask',
    prefix: 3,
    suffix: 4,
    mask_char: '*',
    replace: '',
  })
  if (row.params_json) {
    try {
      const p = JSON.parse(row.params_json) as Record<string, unknown>
      if (typeof p.prefix === 'number') ruleForm.prefix = p.prefix
      if (typeof p.suffix === 'number') ruleForm.suffix = p.suffix
      if (typeof p.mask_char === 'string') ruleForm.mask_char = p.mask_char
      if (typeof p.replace === 'string') ruleForm.replace = p.replace
    } catch {
      /* params_json 非法时忽略，保留默认参数 */
    }
  }
  ruleDialogVisible.value = true
}

/**
 * 按算法序列化动态表单为 params_json：
 * mask→{prefix,suffix,mask_char?}（后端要求必填）；replace→{replace}（键必填，允许空串）；
 * hash/null→不带（后端允许为空）。序列化失败返回 null 并已提示。
 */
function buildParamsJson(): string | undefined | null {
  if (ruleForm.algorithm === 'mask') {
    const ch = ruleForm.mask_char.trim()
    if (ch && [...ch].length !== 1) {
      ElMessage.warning('掩码字符须为单个字符（如 *）')
      return null
    }
    const obj: Record<string, unknown> = {
      prefix: ruleForm.prefix ?? 0,
      suffix: ruleForm.suffix ?? 0,
    }
    if (ch) obj.mask_char = ch
    return JSON.stringify(obj)
  }
  if (ruleForm.algorithm === 'replace') {
    // replace 键必须存在（整值替换为空串也合法），总是序列化该键
    return JSON.stringify({ replace: ruleForm.replace })
  }
  return undefined
}

async function onSaveRule() {
  const valid = await ruleFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const paramsJson = buildParamsJson()
  if (paramsJson === null) return
  ruleSaving.value = true
  try {
    await saveSecMaskRule({
      rule_code: ruleForm.rule_code.trim(),
      rule_name: ruleForm.rule_name.trim(),
      algorithm: ruleForm.algorithm,
      params_json: paramsJson,
    })
    ElMessage.success(ruleEditingId.value === null ? '规则创建成功' : '规则更新成功')
    ruleDialogVisible.value = false
    fetchRules()
    fetchRuleOptions()
  } catch (e) {
    ElMessage.error(errText(e, '保存脱敏规则失败'))
  } finally {
    ruleSaving.value = false
  }
}

async function onDeleteRule(row: SecMaskRule) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除脱敏规则「${row.rule_name || row.rule_code}」吗？被实体打标引用的规则将被拒绝删除。`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteSecMaskRule(row.id)
    ElMessage.success('删除成功')
    fetchRules()
    fetchRuleOptions()
  } catch (e) {
    ElMessage.error(errText(e, '删除脱敏规则失败'))
  }
}

// 脱敏预览（每行"预览"按钮）
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewRule = ref<SecMaskRule | null>(null)
const previewSample = ref('')
const previewResult = ref<MaskPreviewResp | null>(null)

function openPreview(row: SecMaskRule) {
  previewRule.value = row
  previewSample.value = ''
  previewResult.value = null
  previewVisible.value = true
}

async function onPreview() {
  if (!previewSample.value) {
    ElMessage.warning('请输入样例值')
    return
  }
  if (previewRule.value?.id === undefined) return
  previewLoading.value = true
  try {
    previewResult.value = await maskPreview({
      rule_id: previewRule.value.id,
      sample: previewSample.value,
    })
  } catch (e) {
    ElMessage.error(errText(e, '脱敏预览失败'))
  } finally {
    previewLoading.value = false
  }
}

function algorithmLabelOf(alg?: string) {
  return algorithmOptions.find((a) => a.value === alg)?.label || alg || '-'
}

// ============ 页签 4：授权矩阵 ============
const grLoading = ref(false)
const grList = ref<SecGrant[]>([])
const grTotal = ref(0)
const grQuery = reactive({ subject: '', resource: '', page: 1, size: 10 })

async function fetchGrants() {
  grLoading.value = true
  try {
    const resp = await secGrantList({
      ...grQuery,
      subject: grQuery.subject.trim() || undefined,
      resource: grQuery.resource.trim() || undefined,
    })
    grList.value = resp.list || []
    grTotal.value = resp.total || 0
  } catch (e) {
    ElMessage.error(errText(e, '获取授权列表失败'))
  } finally {
    grLoading.value = false
  }
}

function grSearch() {
  grQuery.page = 1
  fetchGrants()
}

/** 授权主体格式：type:id，type ∈ user/app/role，id 段须含非空白字符（对齐后端 ValidSubject） */
const subjectPattern = /^(user|app|role):.*\S/

const grantDialogVisible = ref(false)
const grantSaving = ref(false)
const grantFormRef = ref<FormInstance>()
const grantForm = reactive({ subject: '', resource: '', resource_id: '', level_code: '', expire_date: '' })

const grantRules: FormRules = {
  subject: [
    { required: true, message: '请输入授权主体', trigger: 'blur' },
    { pattern: subjectPattern, message: '格式：user:1 / app:key / role:admin', trigger: 'blur' },
  ],
  resource: [{ required: true, message: '请输入资源类型', trigger: 'blur' }],
  level_code: [{ required: true, message: '请选择许可密级', trigger: 'change' }],
}

function openGrantCreate() {
  Object.assign(grantForm, { subject: '', resource: '', resource_id: '', level_code: '', expire_date: '' })
  grantDialogVisible.value = true
}

async function onSaveGrant() {
  const valid = await grantFormRef.value?.validate().catch(() => false)
  if (!valid) return
  grantSaving.value = true
  try {
    await saveSecGrant({
      subject: grantForm.subject.trim(),
      resource: grantForm.resource.trim(),
      resource_id: grantForm.resource_id.trim() || undefined,
      level_code: grantForm.level_code,
      // expire_date 为毫秒时间戳字符串（value-format="x"），0/缺省=永久
      expire_at: grantForm.expire_date ? Math.floor(Number(grantForm.expire_date) / 1000) : undefined,
    })
    ElMessage.success('授权保存成功')
    grantDialogVisible.value = false
    fetchGrants()
  } catch (e) {
    ElMessage.error(errText(e, '保存授权失败'))
  } finally {
    grantSaving.value = false
  }
}

async function onDeleteGrant(row: SecGrant) {
  if (row.id === undefined) return
  const ok = await ElMessageBox.confirm(
    `确定删除授权（主体 ${row.subject || '-'} → ${row.resource || '-'}）吗？`,
    '删除确认',
    { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
  )
    .then(() => true)
    .catch(() => false)
  if (!ok) return
  try {
    await deleteSecGrant(row.id)
    ElMessage.success('删除成功')
    fetchGrants()
  } catch (e) {
    ElMessage.error(errText(e, '删除授权失败'))
  }
}

// 权限检查器
const checkForm = reactive({ subject: '', resource: '', resource_id: '', level_code: '' })
const checking = ref(false)
const checkResult = ref<SecCheckPermResp | null>(null)

async function onCheckPerm() {
  if (!checkForm.subject.trim() || !checkForm.resource.trim()) {
    ElMessage.warning('请至少填写主体与资源')
    return
  }
  if (!subjectPattern.test(checkForm.subject.trim())) {
    ElMessage.warning('主体格式须为 user:1 / app:key / role:admin')
    return
  }
  checking.value = true
  checkResult.value = null
  try {
    checkResult.value = await secCheckPerm({
      subject: checkForm.subject.trim(),
      resource: checkForm.resource.trim(),
      resource_id: checkForm.resource_id.trim() || undefined,
      level_code: checkForm.level_code || undefined,
    })
  } catch (e) {
    ElMessage.error(errText(e, '权限检查失败'))
  } finally {
    checking.value = false
  }
}

onMounted(fetchLevels)
</script>

<template>
  <div>
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <!-- ===== 分级目录 ===== -->
        <el-tab-pane label="分级目录" name="level">
          <el-alert
            title="数据密级定义（L1 公开 / L2 内部 / L3 敏感 / L4 机密，L5-L9 留给租户自定义）。目录为空时首次查询自动播种内置四级。"
            type="info"
            :closable="false"
            show-icon
            class="tab-tip"
          />
          <div class="toolbar">
            <el-button type="primary" @click="openLevelCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建分级
            </el-button>
            <el-button @click="fetchLevels">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="levelLoading" :data="levelList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column label="分级编码" width="120">
              <template #default="{ row }">
                <el-tag :type="levelTagOf(row.code)" size="small">{{ row.code || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="分级名称" min-width="140" show-overflow-tooltip />
            <el-table-column prop="comment" label="备注" min-width="220" show-overflow-tooltip />
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openLevelEdit(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- ===== 实体打标 ===== -->
        <el-tab-pane label="实体打标" name="entity">
          <div class="toolbar">
            <el-select
              v-model="etQuery.level_code"
              placeholder="全部密级"
              clearable
              style="width: 150px"
              @change="etSearch"
            >
              <el-option
                v-for="l in levelList"
                :key="l.id"
                :label="`${l.code || ''} ${l.name || ''}`"
                :value="l.code || ''"
              />
            </el-select>
            <el-button type="primary" @click="openTagCreate">
              <el-icon><Plus /></el-icon>&nbsp;实体打标
            </el-button>
            <el-button @click="fetchEntityTags">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="etLoading" :data="etList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="entity_id" label="实体 ID" width="110" />
            <el-table-column prop="entity_name" label="实体名称" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.entity_name || '-' }}</template>
            </el-table-column>
            <el-table-column label="密级" width="110">
              <template #default="{ row }">
                <el-tag :type="levelTagOf(row.level_code)" size="small">{{ row.level_code || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="脱敏规则" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">{{ row.mask_rule_code || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openTagEdit(row)">重新打标</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="etQuery.page"
            v-model:page-size="etQuery.size"
            :total="etTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchEntityTags"
            @size-change="etSearch"
          />
        </el-tab-pane>

        <!-- ===== 脱敏规则 ===== -->
        <el-tab-pane label="脱敏规则" name="maskrule">
          <div class="toolbar">
            <el-button type="primary" @click="openRuleCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建规则
            </el-button>
            <el-button @click="fetchRules">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="ruleLoading" :data="ruleList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="rule_code" label="规则编码" min-width="130" show-overflow-tooltip />
            <el-table-column prop="rule_name" label="规则名称" min-width="150" show-overflow-tooltip />
            <el-table-column label="算法" width="190">
              <template #default="{ row }">
                <el-tag size="small" type="warning">{{ algorithmLabelOf(row.algorithm) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="params_json" label="参数" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ row.params_json || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="170" fixed="right">
              <template #default="{ row }">
                <el-button link type="success" @click="openPreview(row)">预览</el-button>
                <el-button link type="primary" @click="openRuleEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="onDeleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="ruleQuery.page"
            v-model:page-size="ruleQuery.size"
            :total="ruleTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchRules"
            @size-change="fetchRules"
          />
        </el-tab-pane>

        <!-- ===== 授权矩阵 ===== -->
        <el-tab-pane label="授权矩阵" name="grant">
          <!-- 权限检查器 -->
          <el-card shadow="never" class="perm-card">
            <template #header>
              <span class="perm-title">权限检查器</span>
            </template>
            <el-form inline @submit.prevent>
              <el-form-item label="主体">
                <el-input
                  v-model="checkForm.subject"
                  placeholder="user:1 / app:key / role:admin"
                  style="width: 210px"
                  @keyup.enter="onCheckPerm"
                />
              </el-form-item>
              <el-form-item label="资源">
                <el-input
                  v-model="checkForm.resource"
                  placeholder="资源类型，如 table"
                  style="width: 150px"
                  @keyup.enter="onCheckPerm"
                />
              </el-form-item>
              <el-form-item label="资源 ID">
                <el-input
                  v-model="checkForm.resource_id"
                  placeholder="可选"
                  clearable
                  style="width: 130px"
                  @keyup.enter="onCheckPerm"
                />
              </el-form-item>
              <el-form-item label="数据密级">
                <el-select
                  v-model="checkForm.level_code"
                  placeholder="可选"
                  clearable
                  style="width: 130px"
                >
                  <el-option
                    v-for="l in levelList"
                    :key="l.id"
                    :label="l.code || ''"
                    :value="l.code || ''"
                  />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="checking" @click="onCheckPerm">
                  <el-icon><CircleCheck /></el-icon>&nbsp;检查权限
                </el-button>
              </el-form-item>
            </el-form>
            <div v-if="checkResult" class="perm-result" :class="checkResult.allowed ? 'is-allowed' : 'is-denied'">
              <span class="perm-verdict">{{ checkResult.allowed ? '✓ 允许访问' : '✗ 拒绝访问' }}</span>
              <span class="perm-reason">{{ checkResult.reason || '-' }}</span>
            </div>
          </el-card>

          <div class="toolbar">
            <el-input
              v-model="grQuery.subject"
              placeholder="按主体搜索（user:1）"
              clearable
              style="width: 190px"
              @keyup.enter="grSearch"
              @clear="grSearch"
            />
            <el-input
              v-model="grQuery.resource"
              placeholder="按资源类型搜索"
              clearable
              style="width: 160px"
              @keyup.enter="grSearch"
              @clear="grSearch"
            />
            <el-button type="primary" plain @click="grSearch">
              <el-icon><Search /></el-icon>&nbsp;查询
            </el-button>
            <el-button type="primary" @click="openGrantCreate">
              <el-icon><Plus /></el-icon>&nbsp;新建授权
            </el-button>
            <el-button @click="fetchGrants">
              <el-icon><Refresh /></el-icon>&nbsp;刷新
            </el-button>
          </div>

          <el-table v-loading="grLoading" :data="grList" stripe border>
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="subject" label="主体" min-width="140" show-overflow-tooltip />
            <el-table-column prop="resource" label="资源" min-width="110" show-overflow-tooltip />
            <el-table-column label="资源 ID" min-width="110" show-overflow-tooltip>
              <template #default="{ row }">{{ row.resource_id || '-' }}</template>
            </el-table-column>
            <el-table-column label="许可密级" width="110">
              <template #default="{ row }">
                <el-tag :type="levelTagOf(row.level_code)" size="small">{{ row.level_code || '-' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="过期时间" width="170">
              <template #default="{ row }">{{ row.expire_at ? fmtTime(row.expire_at) : '永久' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" @click="onDeleteGrant(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="grQuery.page"
            v-model:page-size="grQuery.size"
            :total="grTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pager"
            @current-change="fetchGrants"
            @size-change="grSearch"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 分级新建/编辑 -->
    <el-dialog
      v-model="levelDialogVisible"
      :title="levelEditingId === null ? '新建分级' : '编辑分级'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="levelFormRef" :model="levelForm" :rules="levelRules" label-width="90px">
        <el-form-item label="分级编码" prop="code">
          <el-input v-model="levelForm.code" placeholder="L1-L9（内置 L1-L4）" maxlength="3" />
        </el-form-item>
        <el-form-item label="分级名称" prop="name">
          <el-input v-model="levelForm.name" placeholder="如：公开 / 内部 / 敏感 / 机密" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="levelForm.comment" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="levelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="levelSaving" @click="onSaveLevel">保存</el-button>
      </template>
    </el-dialog>

    <!-- 实体打标（upsert） -->
    <el-dialog v-model="tagDialogVisible" title="实体打标" width="560px" destroy-on-close>
      <el-form ref="tagFormRef" :model="tagForm" :rules="tagRules" label-width="100px">
        <el-form-item label="实体 ID" prop="entity_id">
          <el-input-number
            v-model="tagForm.entity_id"
            :min="1"
            :step="1"
            :precision="0"
            controls-position="right"
            placeholder="表/字段/资产等实体主键"
            style="width: 100%"
          />
          <div class="form-tip">同一实体重复打标等价覆盖</div>
        </el-form-item>
        <el-form-item label="实体名称">
          <el-input v-model="tagForm.entity_name" placeholder="如：dwd_order.user_phone" maxlength="128" />
        </el-form-item>
        <el-form-item label="密级" prop="level_code">
          <el-select v-model="tagForm.level_code" placeholder="从分级目录选择" style="width: 100%">
            <el-option
              v-for="l in levelList"
              :key="l.id"
              :label="`${l.code || ''} ${l.name || ''}`"
              :value="l.code || ''"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="脱敏规则">
          <el-select
            v-model="tagForm.mask_rule_code"
            placeholder="可选，绑定默认脱敏规则"
            clearable
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="r in ruleOptions"
              :key="r.id"
              :label="`${r.rule_code || ''} ${r.rule_name || ''}`"
              :value="r.rule_code || ''"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tagDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="tagSaving" @click="onSaveTag">保存</el-button>
      </template>
    </el-dialog>

    <!-- 脱敏规则新建/编辑（params_json 动态表单） -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="ruleEditingId === null ? '新建脱敏规则' : '编辑脱敏规则'"
      width="580px"
      destroy-on-close
    >
      <el-form ref="ruleFormRef" :model="ruleForm" :rules="ruleRules" label-width="100px">
        <el-form-item label="规则编码" prop="rule_code">
          <el-input v-model="ruleForm.rule_code" placeholder="如：mask_phone" maxlength="64" />
        </el-form-item>
        <el-form-item label="规则名称" prop="rule_name">
          <el-input v-model="ruleForm.rule_name" placeholder="如：手机号掩码" maxlength="64" />
        </el-form-item>
        <el-form-item label="算法" prop="algorithm">
          <el-select v-model="ruleForm.algorithm" style="width: 100%">
            <el-option v-for="a in algorithmOptions" :key="a.value" :label="a.label" :value="a.value" />
          </el-select>
        </el-form-item>

        <!-- mask：保留前 prefix 后 suffix，中间以 mask_char 填充 -->
        <template v-if="ruleForm.algorithm === 'mask'">
          <el-form-item label="保留前缀">
            <el-input-number v-model="ruleForm.prefix" :min="0" :precision="0" controls-position="right" />
            <span class="form-tip">保留前 prefix 个字符</span>
          </el-form-item>
          <el-form-item label="保留后缀">
            <el-input-number v-model="ruleForm.suffix" :min="0" :precision="0" controls-position="right" />
            <span class="form-tip">保留后 suffix 个字符；长度不足时整值掩码</span>
          </el-form-item>
          <el-form-item label="掩码字符">
            <el-input v-model="ruleForm.mask_char" placeholder="默认 *" maxlength="4" style="width: 120px" />
            <span class="form-tip">单个字符，缺省为 *</span>
          </el-form-item>
        </template>

        <!-- replace：整值替换 -->
        <el-form-item v-if="ruleForm.algorithm === 'replace'" label="替换值">
          <el-input v-model="ruleForm.replace" placeholder="整值替换为该值（允许空串）" maxlength="128" />
        </el-form-item>

        <!-- hash / null：无参数 -->
        <el-form-item v-if="ruleForm.algorithm === 'hash'" label="参数">
          <span class="form-tip">hash 算法无需参数（SHA-256 十六进制前 16 位）</span>
        </el-form-item>
        <el-form-item v-if="ruleForm.algorithm === 'null'" label="参数">
          <span class="form-tip">null 算法无需参数（返回空串）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSaving" @click="onSaveRule">保存</el-button>
      </template>
    </el-dialog>

    <!-- 脱敏预览 -->
    <el-dialog v-model="previewVisible" title="脱敏预览" width="560px" destroy-on-close>
      <el-descriptions v-if="previewRule" :column="1" border size="small" class="preview-desc">
        <el-descriptions-item label="规则">
          {{ previewRule.rule_code || '-' }}（{{ previewRule.rule_name || '-' }}）
        </el-descriptions-item>
        <el-descriptions-item label="算法">
          {{ algorithmLabelOf(previewRule.algorithm) }}
        </el-descriptions-item>
      </el-descriptions>
      <el-form label-width="100px" @submit.prevent>
        <el-form-item label="样例值">
          <el-input
            v-model="previewSample"
            placeholder="输入待脱敏的样例值，如 13812345678"
            @keyup.enter="onPreview"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="previewLoading" @click="onPreview">
            <el-icon><View /></el-icon>&nbsp;预览脱敏
          </el-button>
        </el-form-item>
      </el-form>
      <div v-if="previewResult" class="preview-result">
        <div class="preview-label">
          脱敏结果
          <el-tag v-if="previewResult.rule_code" size="small" type="info" class="preview-rule-tag">
            命中规则 {{ previewResult.rule_code }}
          </el-tag>
        </div>
        <div class="preview-masked">{{ previewResult.masked || '（空串）' }}</div>
      </div>
      <template #footer>
        <el-button @click="previewVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 授权新建 -->
    <el-dialog v-model="grantDialogVisible" title="新建授权" width="560px" destroy-on-close>
      <el-form ref="grantFormRef" :model="grantForm" :rules="grantRules" label-width="100px">
        <el-form-item label="授权主体" prop="subject">
          <el-input v-model="grantForm.subject" placeholder="user:1 / app:key / role:admin" maxlength="64" />
        </el-form-item>
        <el-form-item label="资源类型" prop="resource">
          <el-input v-model="grantForm.resource" placeholder="如 table / api / dataset" maxlength="64" />
        </el-form-item>
        <el-form-item label="资源 ID">
          <el-input v-model="grantForm.resource_id" placeholder="可选，留空对整类资源生效" maxlength="64" />
        </el-form-item>
        <el-form-item label="许可密级" prop="level_code">
          <el-select v-model="grantForm.level_code" placeholder="可访问的最高密级" style="width: 100%">
            <el-option
              v-for="l in levelList"
              :key="l.id"
              :label="`${l.code || ''} ${l.name || ''}`"
              :value="l.code || ''"
            />
          </el-select>
          <div class="form-tip">数据密级 ≤ 许可密级才放行</div>
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="grantForm.expire_date"
            type="date"
            placeholder="可选，缺省永久有效"
            value-format="x"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="grantSaving" @click="onSaveGrant">保存</el-button>
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
.tab-tip {
  margin-bottom: 14px;
}
.form-tip {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}
.perm-card {
  margin-bottom: 16px;
}
.perm-title {
  font-weight: 600;
}
.perm-result {
  display: flex;
  align-items: baseline;
  gap: 14px;
  padding: 12px 16px;
  border-radius: 6px;
}
.perm-result.is-allowed {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
}
.perm-result.is-denied {
  background: #fef0f0;
  border: 1px solid #fde2e2;
}
.perm-verdict {
  font-size: 22px;
  font-weight: 700;
}
.is-allowed .perm-verdict {
  color: #67c23a;
}
.is-denied .perm-verdict {
  color: #f56c6c;
}
.perm-reason {
  font-size: 13px;
  color: #606266;
}
.preview-desc {
  margin-bottom: 16px;
}
.preview-result {
  margin-top: 4px;
}
.preview-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 13px;
  color: #909399;
}
.preview-rule-tag {
  font-weight: 400;
}
.preview-masked {
  font-size: 20px;
  font-weight: 600;
  padding: 12px 16px;
  background: #282c34;
  color: #98c379;
  border-radius: 6px;
  word-break: break-all;
}
</style>
