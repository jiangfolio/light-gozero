<script setup lang="ts">
/**
 * 表格/表单超长文本单行省略，悬停展示全文。
 */
defineProps<{
  text?: string | number | null
  empty?: string
}>()
</script>

<template>
  <ATooltip v-if="text !== undefined && text !== null && String(text) !== ''" :title="String(text)">
    <span class="ellipsis-text">{{ text }}</span>
  </ATooltip>
  <span v-else>{{ empty ?? '-' }}</span>
</template>

<style scoped>
.ellipsis-text {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: keep-all;
}
</style>
