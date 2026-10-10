<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { login } from '@/api/auth'
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
    const reply = await login({ mobile: form.mobile, password: form.password })
    const token = reply.jwt_token?.access_token || ''
    if (!token) {
      ElMessage.error('登录响应缺少 access_token')
      return
    }
    auth.setAuth(token, reply.uid, reply.username || reply.nickname || '')
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
    <el-card class="login-card">
      <div class="login-head">
        <el-icon :size="30" color="#409eff"><DataLine /></el-icon>
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
  </div>
</template>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1f3b73 0%, #2f5ca8 50%, #409eff 100%);
}
.login-card {
  width: 380px;
  padding: 12px 8px;
  border-radius: 10px;
}
.login-head {
  text-align: center;
  margin-bottom: 18px;
}
.login-title {
  margin: 10px 0 4px;
  font-size: 20px;
  color: #303133;
}
.login-sub {
  margin: 0;
  font-size: 13px;
  color: #909399;
}
.login-btn {
  width: 100%;
}
.login-tip {
  text-align: center;
  font-size: 12px;
  color: #c0c4cc;
  margin: 4px 0 0;
}
</style>
