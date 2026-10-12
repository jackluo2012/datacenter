<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { sysLogin } from '@/api/auth'
import { ApiError } from '@/api/http'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({
  mobile: '13800138000',
  password: '123456',
})

const rules: FormRules = {
  mobile: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1\d{10}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' },
  ],
}

async function onSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    // 管理台登录走 sys_user（RBAC 账号，种子内置 admin/admin123）
    const reply = await sysLogin({ username: form.mobile, password: form.password })
    if (!reply.token) {
      ElMessage.error('登录响应缺少 token')
      return
    }
    auth.setAuth(reply.token, reply.uid, reply.username || reply.nickname || '')
    await auth.fetchInfo().catch(() => undefined) // 拉角色/权限/菜单（失败降级静态菜单）
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    ElMessage.error(e instanceof ApiError ? e.message : '登录失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- 左侧品牌区：渐变底 + 产品名 + 三大价值点 -->
    <div class="login-brand">
      <div class="brand-logo">
        <span class="brand-logo-icon"><el-icon :size="22"><DataLine /></el-icon></span>
        <span class="brand-logo-text">数据中台 DataCenter</span>
      </div>

      <h1 class="brand-title">让企业数据<br />资产化 · 服务化 · 智能化</h1>
      <p class="brand-sub">一站式数据中台：从数据接入、治理到价值释放，一个平台全部搞定</p>

      <ul class="brand-points">
        <li class="brand-point">
          <span class="point-icon"><el-icon :size="19"><Coin /></el-icon></span>
          <div class="point-text">
            <b>数据资产化</b>
            <span>元数据采集 · 资产目录 · 健康分治理</span>
          </div>
        </li>
        <li class="brand-point">
          <span class="point-icon"><el-icon :size="19"><Connection /></el-icon></span>
          <div class="point-text">
            <b>数据服务化</b>
            <span>SQL 即服务 · 细粒度授权 · 稳定 API</span>
          </div>
        </li>
        <li class="brand-point">
          <span class="point-icon"><el-icon :size="19"><MagicStick /></el-icon></span>
          <div class="point-text">
            <b>数据智能化</b>
            <span>AI 问数 · 智能标签 · 指标口径统一</span>
          </div>
        </li>
      </ul>

      <div class="brand-tags">
        <span>元数据</span><span>数据质量</span><span>数据服务</span><span>指标中心</span><span>AI 问数</span>
      </div>
    </div>

    <!-- 右侧登录卡片 -->
    <div class="login-panel">
      <el-card class="login-card" shadow="always">
        <div class="login-head">
          <span class="login-logo"><el-icon :size="20"><DataLine /></el-icon></span>
          <h2 class="login-title">数据中台控制台</h2>
          <p class="login-sub">统一数据资产管理入口</p>
        </div>
        <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="onSubmit">
          <el-form-item prop="mobile">
            <el-input v-model.trim="form.mobile" placeholder="手机号" maxlength="11">
              <template #prefix><el-icon><Iphone /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model.trim="form.password"
              type="password"
              show-password
              placeholder="密码"
            >
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" class="login-btn" :loading="loading" @click="onSubmit">
              登 录
            </el-button>
          </el-form-item>
        </el-form>
        <p class="login-tip">演示账号：13800138000 / 123456</p>
      </el-card>
      <p class="login-copyright">© 2026 DataCenter · 一体化数据中台平台</p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  overflow: hidden;
}

/* ========== 左侧品牌区 ========== */
.login-brand {
  flex: 1.15;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 6vh 6vw;
  position: relative;
  color: #fff;
  background: linear-gradient(160deg, #1e3a8a 0%, #2563eb 48%, #06b6d4 100%);
  overflow: hidden;
}
/* 装饰光斑 */
.login-brand::before,
.login-brand::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.login-brand::before {
  width: 460px;
  height: 460px;
  right: -140px;
  top: -160px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.16) 0%, transparent 68%);
}
.login-brand::after {
  width: 380px;
  height: 380px;
  left: -120px;
  bottom: -140px;
  background: radial-gradient(circle, rgba(6, 182, 212, 0.35) 0%, transparent 70%);
}
.login-brand > * {
  position: relative;
  z-index: 1;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6vh;
  font-size: 17px;
  font-weight: 600;
  letter-spacing: 0.5px;
}
.brand-logo-icon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(4px);
}

.brand-title {
  margin: 0;
  font-size: 38px;
  line-height: 1.32;
  font-weight: 700;
  letter-spacing: 1px;
}
.brand-sub {
  margin: 16px 0 0;
  max-width: 460px;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.82);
}

.brand-points {
  list-style: none;
  margin: 5vh 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 22px;
}
.brand-point {
  display: flex;
  align-items: center;
  gap: 14px;
}
.point-icon {
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.22);
  backdrop-filter: blur(4px);
  transition: var(--dc-transition);
}
.brand-point:hover .point-icon {
  background: rgba(255, 255, 255, 0.28);
  transform: translateY(-2px);
}
.point-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.point-text b {
  font-size: 15px;
  font-weight: 600;
}
.point-text span {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.72);
}

.brand-tags {
  margin-top: 5vh;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.brand-tags span {
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  color: rgba(255, 255, 255, 0.85);
}

/* ========== 右侧登录面板 ========== */
.login-panel {
  flex: 1;
  min-width: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  background-color: var(--dc-gray-50);
  /* 细点阵纹理，弱化空旷感 */
  background-image: radial-gradient(rgba(37, 99, 235, 0.08) 1px, transparent 1px);
  background-size: 22px 22px;
}
.login-card {
  width: 400px;
  border: none;
  border-radius: var(--dc-radius-xl);
  box-shadow: var(--dc-shadow-lg);
}
.login-card :deep(.el-card__body) {
  padding: 38px 36px 28px;
}

.login-head {
  text-align: center;
  margin-bottom: 24px;
}
.login-logo {
  width: 46px;
  height: 46px;
  margin: 0 auto;
  border-radius: 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: var(--dc-grad-brand);
  box-shadow: var(--dc-shadow-primary);
}
.login-title {
  margin: 14px 0 6px;
  font-size: 22px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.login-sub {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.login-btn {
  width: 100%;
  height: 42px;
  font-size: 15px;
  letter-spacing: 6px;
}

/* 演示账号提示：弱化为浅色胶囊 */
.login-tip {
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  background: var(--el-fill-color-light);
  border-radius: 999px;
  padding: 5px 0;
  margin: 2px 0 0;
}

.login-copyright {
  position: absolute;
  bottom: 20px;
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

/* ========== 响应式 & 暗色 ========== */
@media (max-width: 960px) {
  .login-brand {
    display: none;
  }
  .login-panel {
    min-width: 0;
    width: 100%;
  }
  .login-card {
    width: min(400px, calc(100vw - 48px));
  }
}

html.dark .login-panel {
  background-color: var(--el-bg-color-page);
  background-image: radial-gradient(rgba(59, 130, 246, 0.14) 1px, transparent 1px);
}
html.dark .login-title {
  color: var(--el-text-color-primary);
}
</style>
