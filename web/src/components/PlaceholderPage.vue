<script setup lang="ts">
/**
 * 占位页统一骨架：功能提示 + 表格骨架 + 分页占位
 * 后续批次用真实列表页替换各模块的 List.vue
 */
export interface PlaceholderColumn {
  prop: string
  label: string
  width?: number | string
  minWidth?: number | string
}

defineProps<{
  title: string
  description: string
  endpoints: string
  columns: PlaceholderColumn[]
}>()
</script>

<template>
  <div>
    <el-card shadow="never">
      <div class="toolbar">
        <div>
          <h3 class="title">{{ title }}</h3>
          <p class="desc">{{ description }}</p>
        </div>
        <el-tag type="warning" effect="light" size="large">
          <el-icon><Loading /></el-icon>&nbsp;功能对接中 · 本批次为骨架占位
        </el-tag>
      </div>

      <el-alert
        :title="`规划接口：${endpoints}`"
        type="info"
        :closable="false"
        show-icon
        class="endpoint-alert"
      />

      <!-- 表格骨架：仅展示列头，数据待接口上线 -->
      <el-table :data="[]" stripe border>
        <el-table-column
          v-for="c in columns"
          :key="c.prop"
          :prop="c.prop"
          :label="c.label"
          :width="c.width"
          :min-width="c.minWidth"
        />
        <template #empty>
          <el-empty description="接口待上线，敬请期待" :image-size="72" />
        </template>
      </el-table>

      <el-pagination
        :total="0"
        :page-size="10"
        layout="total, prev, pager, next"
        disabled
        class="pager"
      />
    </el-card>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 12px;
}
.title {
  margin: 0 0 6px;
  font-size: 17px;
  color: #303133;
}
.desc {
  margin: 0;
  font-size: 13px;
  color: #909399;
}
.endpoint-alert {
  margin-bottom: 14px;
}
.pager {
  margin-top: 14px;
  justify-content: flex-end;
}
</style>
