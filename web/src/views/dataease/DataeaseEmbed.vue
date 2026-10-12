<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  defaultDataeaseUrl,
  executePlatformApi,
  loadDataeaseUrl,
  normalizeUrl,
  probeDataease,
  clearDataeaseUrl,
  saveDataeaseUrl,
  type ExecuteApiResp,
} from '@/api/dataease'
import { ApiError } from '@/api/http'

/**
 * 数据可视化（DataEase 集成）页，两种模式：
 *  a) 仪表板嵌入：iframe 加载 DataEase 地址（建议用「分享 → 公共链接」生成的免登录链接），
 *     地址可存 localStorage / 构建期 VITE_DATAEASE_URL；
 *  b) 部署引导：未部署时展示一键 compose 命令（复制即用）+ 对接平台数据两条路线摘要
 *     + API 数据源签名链路自测。
 *
 * 说明：本页面自包含，不改 router；入口由动态菜单下发
 * （种子 SQL 见 docs/dataease-menu-seed.sql，组件名 dataease/DataeaseEmbed）。
 */

// ---------- 模式与地址 ----------
const mode = ref<string>(loadDataeaseUrl() ? 'embed' : 'guide')

const urlInput = ref(loadDataeaseUrl())
/** 当前实际嵌入的地址（点「加载」后生效，避免输入过程反复重载 iframe） */
const activeUrl = ref(loadDataeaseUrl())
const iframeKey = ref(0)

function loadEmbed() {
  const target = normalizeUrl(urlInput.value)
  if (!target || !/^https?:\/\//.test(target)) {
    ElMessage.warning('请填写 http(s):// 开头的 DataEase 地址')
    return
  }
  urlInput.value = target
  activeUrl.value = target
  iframeKey.value++
  mode.value = 'embed'
  doProbe()
}

function saveDefault() {
  const target = normalizeUrl(urlInput.value)
  if (!target) {
    ElMessage.warning('地址为空，如需回退默认请用「恢复默认」')
    return
  }
  saveDataeaseUrl(target)
  ElMessage.success('已保存为本浏览器默认地址（localStorage）')
}

function resetDefault() {
  clearDataeaseUrl()
  urlInput.value = defaultDataeaseUrl()
  activeUrl.value = urlInput.value
  iframeKey.value++
  ElMessage.success(`已回退默认地址：${activeUrl.value}`)
}

function openNewWindow() {
  const target = normalizeUrl(urlInput.value)
  if (!target) {
    ElMessage.warning('请先填写 DataEase 地址')
    return
  }
  window.open(target, '_blank')
}

// ---------- 连通性探测 ----------
const probing = ref(false)
/** null=未探测；true=网络可达；false=不可达（未部署或地址错误） */
const reachable = ref<boolean | null>(null)

async function doProbe() {
  const target = normalizeUrl(urlInput.value)
  if (!target) return
  probing.value = true
  reachable.value = null
  try {
    reachable.value = await probeDataease(target)
  } finally {
    probing.value = false
  }
}

// ---------- 部署引导：一键命令（与 deploy/dataease、docs/DEPLOY-DATAEASE.md 保持一致） ----------
const CMD_UP = [
  '# 1) 首次部署：生成配置（仓库根目录执行）',
  'cp deploy/dataease/dataease.properties.template deploy/dataease/dataease.properties',
  '# 2) 启动 DataEase（v1.18.27，宿主机端口 8088）',
  'docker compose -f deploy/dataease/docker-compose.dataease.yml up -d',
  '# 3) 观察日志：Flyway 建表 + Tomcat/Spring 启动完成即就绪（首启约 1-3 分钟）',
  'docker compose -f deploy/dataease/docker-compose.dataease.yml logs -f dataease',
].join('\n')

const CMD_DB = [
  '# 平台容器编排（docker-compose.yml，mysql 为服务名）',
  'docker compose exec mysql mysql -uroot -padmin -e "CREATE DATABASE IF NOT EXISTS dataease DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"',
  '# 若 MySQL 是 server.sh 起的独立容器（dc-run-mysql）',
  'docker exec dc-run-mysql mysql -uroot -padmin -e "CREATE DATABASE IF NOT EXISTS dataease DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"',
].join('\n')

const CMD_SEED = [
  '# 「数据可视化」菜单种子（幂等，先按脚本头部说明核对 @beid）',
  'docker compose exec -T mysql mysql -uroot -padmin datacenter < docs/dataease-menu-seed.sql',
].join('\n')

async function copyText(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`${label}已复制`)
  } catch {
    ElMessage.error('复制失败，请手动选择文本复制')
  }
}

// ---------- API 数据源签名链路自测（对照 datasvc 签名机制） ----------
const appKey = ref('')
const appSecret = ref('')
const apiPath = ref('/api/v1/demo/orders')
const paramsJson = ref('{"user_id":19,"limit":10}')
const executing = ref(false)
const execResult = ref<ExecuteApiResp | null>(null)

function prettyRows(raw?: string): string {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function runSelfTest() {
  if (!appKey.value.trim() || !appSecret.value.trim()) {
    ElMessage.warning('请先填写 app_key / app_secret（数据服务 → 应用管理 注册获得）')
    return
  }
  executing.value = true
  execResult.value = null
  try {
    execResult.value = await executePlatformApi(
      appKey.value.trim(),
      appSecret.value.trim(),
      apiPath.value.trim(),
      paramsJson.value,
    )
    ElMessage.success('执行成功，签名链路可用，可按文档把该 API 配为 DataEase 数据源')
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '执行失败（检查 API 是否已发布/授权、ts 是否偏差超 300 秒）')
  } finally {
    executing.value = false
  }
}

onMounted(() => {
  if (mode.value === 'embed') doProbe()
})
</script>

<template>
  <div class="dataease-page">
    <!-- 头部：模式切换 + 地址工具条 -->
    <el-card shadow="never" class="head-card">
      <div class="head-row">
        <el-radio-group v-model="mode">
          <el-radio-button value="embed">仪表板嵌入</el-radio-button>
          <el-radio-button value="guide">部署引导</el-radio-button>
        </el-radio-group>

        <template v-if="mode === 'embed'">
          <el-input
            v-model="urlInput"
            class="url-input"
            placeholder="DataEase 地址或仪表板公共链接，如 http://localhost:8088"
            clearable
            @keyup.enter="loadEmbed"
          />
          <el-button type="primary" @click="loadEmbed">加载</el-button>
          <el-button @click="doProbe" :loading="probing">测连通</el-button>
          <el-button @click="openNewWindow">新窗口</el-button>
          <el-button @click="saveDefault">存为默认</el-button>
          <el-button @click="resetDefault">恢复默认</el-button>
          <el-tag v-if="reachable === true" type="success" class="probe-tag">网络可达</el-tag>
          <el-tag v-else-if="reachable === false" type="danger" class="probe-tag">
            不可达（未部署或地址不对，见部署引导）
          </el-tag>
        </template>
      </div>
      <div v-if="mode === 'embed'" class="head-tip">
        建议嵌入 DataEase「仪表板 → 分享 → 开启公共链接」生成的免登录链接；嵌整站地址时可在框内直接登录。
        地址保存在本浏览器 localStorage，或构建期经 <code>VITE_DATAEASE_URL</code> 指定。
      </div>
    </el-card>

    <!-- 模式 a：iframe 嵌入 -->
    <el-card v-if="mode === 'embed'" shadow="never" class="body-card">
      <iframe
        v-if="activeUrl"
        :key="iframeKey"
        :src="activeUrl"
        class="de-frame"
        frameborder="0"
        allow="fullscreen"
      />
      <el-empty v-else description="尚未配置 DataEase 地址，请切换到「部署引导」" />
    </el-card>

    <!-- 模式 b：部署引导 -->
    <template v-else>
      <el-alert type="info" :closable="false" class="guide-alert">
        <template #title>
          从零到出图的完整步骤见仓库文档 <b>docs/DEPLOY-DATAEASE.md</b>（compose 启动 → 建库 → 初始化 → 登录
          admin/dataease → 对接平台数据）。以下命令与本仓库 deploy/dataease 编排一致，复制即用。
        </template>
      </el-alert>

      <el-collapse class="guide-collapse">
        <el-collapse-item title="第 1 步：一键启动 DataEase（compose）" name="up">
          <p class="cmd-desc">复用平台 MySQL 存元数据，数据持久化在仓库根 ./data/dataease：</p>
          <pre class="cmd-block">{{ CMD_UP }}</pre>
          <el-button type="primary" size="small" @click="copyText(CMD_UP, '启动命令')">复制启动命令</el-button>
        </el-collapse-item>

        <el-collapse-item title="第 2 步：创建 dataease 元数据库（一次性）" name="db">
          <p class="cmd-desc">
            v1.18 官方镜像不内嵌元数据库，复用平台 MySQL，需先建库（漏建会导致容器反复重启，补建后自动恢复）：
          </p>
          <pre class="cmd-block">{{ CMD_DB }}</pre>
          <el-button type="primary" size="small" @click="copyText(CMD_DB, '建库命令')">复制建库命令</el-button>
        </el-collapse-item>

        <el-collapse-item title="第 3 步：登录" name="login">
          <p class="cmd-desc">
            浏览器打开 <code>http://localhost:8088</code>，默认账号 <b>admin / dataease</b>（首登后修改）；
            系统内新建用户的初始密码为 DataEase123456（dataease.init_password 可改）。
          </p>
        </el-collapse-item>

        <el-collapse-item title="第 4 步：对接平台数据（直连 / API 两条路线）" name="data">
          <p class="cmd-desc">
            <b>直连路线（推荐起步）</b>：DataEase「数据源 → 新建 → MySQL」，主机名
            <code>mysql</code>、端口 <code>3306</code>、数据库 <code>datacenter</code>、账号
            <code>root/admin</code>（与平台同网络，服务名直连）→ 建 SQL 数据集 → 仪表板出图。
          </p>
          <p class="cmd-desc">
            <b>API 路线</b>：把数据服务 API（<code>POST /datasvc/execute</code>，app_key + HMAC 签名）配为数据源。
            注意 v1.18 未内置 API 数据源（v2 支持），v1.18 下建议先用右侧「API 链路自测」验证签名，再按文档
            走「API 结果落表后直连」；字段级配置见 docs/DEPLOY-DATAEASE.md 路线 B。
          </p>
        </el-collapse-item>

        <el-collapse-item title="第 5 步：把「数据可视化」加进左侧菜单" name="menu">
          <p class="cmd-desc">
            本页面不改 router，入口由动态菜单下发（种子 SQL 幂等可重跑，注意脚本内 @beid 与现有菜单一致）：
          </p>
          <pre class="cmd-block">{{ CMD_SEED }}</pre>
          <el-button type="primary" size="small" @click="copyText(CMD_SEED, '菜单种子命令')">复制种子命令</el-button>
        </el-collapse-item>
      </el-collapse>

      <!-- API 数据源签名链路自测 -->
      <el-card shadow="never" class="selftest-card">
        <template #header>API 数据源链路自测（对照 datasvc 签名机制）</template>
        <el-form label-width="110px" class="selftest-form">
          <el-form-item label="app_key">
            <el-input v-model="appKey" placeholder="数据服务 → 应用管理 注册返回" clearable />
          </el-form-item>
          <el-form-item label="app_secret">
            <el-input
              v-model="appSecret"
              type="password"
              placeholder="仅注册时返回一次"
              show-password
              clearable
            />
          </el-form-item>
          <el-form-item label="API 路径">
            <el-input v-model="apiPath" placeholder="/api/v1/demo/orders" clearable />
          </el-form-item>
          <el-form-item label="参数 JSON">
            <el-input v-model="paramsJson" placeholder='{"user_id":19,"limit":10}' />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="executing" @click="runSelfTest">
              签名执行（POST /datasvc/execute）
            </el-button>
            <span class="selftest-hint">
              前端按 hex(hmac-sha256(app_secret, app_key+ts)) 现算签名，ts 为当前 unix 秒（服务端容差 ±300s）
            </span>
          </el-form-item>
        </el-form>
        <template v-if="execResult">
          <el-descriptions :column="3" size="small" border class="selftest-meta">
            <el-descriptions-item label="耗时">{{ execResult.cost_ms }} ms</el-descriptions-item>
            <el-descriptions-item label="命中缓存">{{ execResult.cache_hit ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="受影响行数">{{ execResult.affected }}</el-descriptions-item>
          </el-descriptions>
          <pre class="cmd-block">{{ prettyRows(execResult.rows_json) }}</pre>
        </template>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.dataease-page {
  padding: 12px;
}

.head-card {
  margin-bottom: 12px;
}

.head-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.url-input {
  width: 380px;
  max-width: 100%;
}

.probe-tag {
  margin-left: 4px;
}

.head-tip {
  margin-top: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
}

.body-card :deep(.el-card__body) {
  padding: 0;
}

.de-frame {
  display: block;
  width: 100%;
  height: calc(100vh - 220px);
  min-height: 480px;
  border: 0;
}

.guide-alert {
  margin-bottom: 12px;
}

.guide-collapse {
  margin-bottom: 12px;
}

.cmd-desc {
  margin: 0 0 8px;
  font-size: 13px;
  line-height: 1.8;
  color: var(--el-text-color-regular);
}

.cmd-block {
  margin: 0 0 10px;
  padding: 10px 12px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-x: auto;
}

.selftest-card {
  margin-bottom: 12px;
}

.selftest-form {
  max-width: 640px;
}

.selftest-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.selftest-meta {
  margin-bottom: 10px;
}
</style>
