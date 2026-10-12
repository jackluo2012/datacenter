<script setup lang="ts">
/** 资产 / 标签 / 标准 计数卡：三枚小体积指标块（图标 + 翻牌数字） */
import FlipCounter from './FlipCounter.vue'
import BsEmpty from './BsEmpty.vue'

interface AssetTriple {
  asset: number | null
  tag: number | null
  std: number | null
}

defineProps<{
  counts: AssetTriple
  loading?: boolean
  failed?: boolean
}>()
</script>

<template>
  <div class="ac-wrap">
    <div class="ac-cell">
      <el-icon :size="22" class="ac-icon cyan"><Coin /></el-icon>
      <div class="ac-meta">
        <FlipCounter :value="counts.asset" />
        <span class="ac-label">资产目录</span>
      </div>
    </div>
    <div class="ac-cell">
      <el-icon :size="22" class="ac-icon green"><PriceTag /></el-icon>
      <div class="ac-meta">
        <FlipCounter :value="counts.tag" class="green" />
        <span class="ac-label">数据标签</span>
      </div>
    </div>
    <div class="ac-cell">
      <el-icon :size="22" class="ac-icon yellow"><Document /></el-icon>
      <div class="ac-meta">
        <FlipCounter :value="counts.std" class="yellow" />
        <span class="ac-label">数据标准</span>
      </div>
    </div>
    <BsEmpty v-if="loading || failed" :loading="loading" :failed="failed" />
  </div>
</template>

<style scoped>
.ac-wrap {
  position: relative;
  height: 100%;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.ac-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border: 1px solid rgba(0, 212, 255, 0.12);
  border-radius: 3px;
  background: linear-gradient(160deg, rgba(0, 212, 255, 0.06), rgba(0, 212, 255, 0.01));
  min-width: 0;
}
.ac-icon {
  flex-shrink: 0;
  padding: 8px;
  border-radius: 50%;
  box-sizing: border-box;
}
.ac-icon.cyan {
  color: #00d4ff;
  background: rgba(0, 212, 255, 0.10);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.25);
}
.ac-icon.green {
  color: #00ff88;
  background: rgba(0, 255, 136, 0.08);
  box-shadow: 0 0 10px rgba(0, 255, 136, 0.22);
}
.ac-icon.yellow {
  color: #ffd855;
  background: rgba(255, 216, 85, 0.08);
  box-shadow: 0 0 10px rgba(255, 216, 85, 0.22);
}
.ac-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.ac-meta :deep(.fc-num) {
  font-size: 22px;
  line-height: 1.15;
}
.ac-meta :deep(.fc-num.green) { color: #00ff88; text-shadow: 0 0 12px rgba(0, 255, 136, 0.5); }
.ac-meta :deep(.fc-num.yellow) { color: #ffd855; text-shadow: 0 0 12px rgba(255, 216, 85, 0.5); }
.ac-label {
  font-size: 11px;
  letter-spacing: 1px;
  color: #7f9bbf;
  white-space: nowrap;
}
</style>
