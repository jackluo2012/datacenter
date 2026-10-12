<template>
  <!-- 布局由路由决定：/login 独立页，其余页面嵌套在 MainLayout 内 -->
  <router-view v-slot="{ Component }">
    <transition name="fade-slide" mode="out-in" appear>
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<style>
/* 基础样式与设计令牌统一收敛至 src/styles/index.css（main.ts 引入） */
</style>
