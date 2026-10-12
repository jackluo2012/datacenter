import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
// Element Plus 暗色变量基础（html.dark 生效）+ 全局设计系统（品牌色/圆角/阴影/动效）
import 'element-plus/theme-chalk/dark/css-vars.css'
import './styles/index.css'

import App from './App.vue'
import router from './router'
import { perm } from './directive/perm'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// 按钮权限指令：v-perm="'sys:user:create'"（perms 集合判断，无权限移除元素）
app.directive('perm', perm)

// 全量注册图标，模板中 <el-icon><Coin /></el-icon> 或 :is="'Coin'" 直接可用
for (const [name, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(name, component)
}

app.mount('#app')
