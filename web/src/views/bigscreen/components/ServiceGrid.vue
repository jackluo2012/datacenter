<script setup lang="ts">
/**
 * 服务状态点阵：中台 17 个服务名 + 在线状态点（脉冲动画）。
 * 探活为示意口径：任一数据接口成功 → 全部在线（网关可达）；全部失败 → 待探活。
 */
import { computed } from 'vue'

const props = defineProps<{ alive: boolean }>()

/** 与仓库后端服务目录对齐（gateway + 16 个 rpc/scheduler） */
const SERVICES: { name: string; desc: string }[] = [
  { name: 'api-gateway', desc: '统一网关' },
  { name: 'sys-rpc', desc: 'RBAC 权限' },
  { name: 'user-rpc', desc: '用户' },
  { name: 'security-rpc', desc: '安全' },
  { name: 'datasource-rpc', desc: '数据源' },
  { name: 'metadata-rpc', desc: '元数据' },
  { name: 'catalog-rpc', desc: '资产目录' },
  { name: 'tag-rpc', desc: '数据标签' },
  { name: 'standard-rpc', desc: '数据标准' },
  { name: 'quality-rpc', desc: '数据质量' },
  { name: 'metric-rpc', desc: '数据指标' },
  { name: 'mdm-rpc', desc: '主数据' },
  { name: 'datasvc-rpc', desc: '数据服务' },
  { name: 'job-scheduler', desc: '调度引擎' },
  { name: 'ide-rpc', desc: 'SQL IDE' },
  { name: 'ops-rpc', desc: '数据运维' },
  { name: 'ai-rpc', desc: 'AI 问数' },
]

const statusText = computed(() => (props.alive ? '在线' : '待探活'))
</script>

<template>
  <div class="sg-wrap">
    <div class="sg-legend">
      <span class="sg-item"><i class="sg-dot on" />{{ statusText }} {{ SERVICES.length }}/{{ SERVICES.length }}</span>
      <span class="sg-note">探活示意 · 以网关连通性为准</span>
    </div>
    <div class="sg-grid">
      <div v-for="s in SERVICES" :key="s.name" class="sg-cell">
        <i class="sg-dot" :class="alive ? 'on' : 'wait'" />
        <div class="sg-meta">
          <span class="sg-name">{{ s.name }}</span>
          <span class="sg-desc">{{ s.desc }}</span>
        </div>
        <span class="sg-status" :class="alive ? 'on' : 'wait'">{{ alive ? '在线' : '待探活' }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sg-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.sg-legend {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  font-size: 12px;
}
.sg-note {
  margin-left: auto;
  font-size: 10px;
  letter-spacing: 1px;
  color: #5a7191;
}
.sg-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #a9c3e2;
}
.sg-grid {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  grid-auto-rows: 1fr;
  gap: 6px;
}
.sg-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 8px;
  border: 1px solid rgba(0, 212, 255, 0.10);
  border-radius: 3px;
  background: rgba(0, 212, 255, 0.03);
  min-width: 0;
}
.sg-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sg-dot.on {
  background: #00ff88;
  box-shadow: 0 0 6px rgba(0, 255, 136, 0.9);
  animation: sg-pulse 2s ease-in-out infinite;
}
.sg-dot.wait {
  background: #ffd855;
  box-shadow: 0 0 6px rgba(255, 216, 85, 0.7);
}
@keyframes sg-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}
.sg-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}
.sg-name {
  font-size: 11px;
  color: #d8ecff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'Consolas', monospace;
}
.sg-desc {
  font-size: 10px;
  color: #5a7191;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sg-status {
  flex-shrink: 0;
  font-size: 10px;
}
.sg-status.on { color: #00ff88; }
.sg-status.wait { color: #ffd855; }
</style>
