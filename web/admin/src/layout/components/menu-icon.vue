<script setup lang="ts">
/**
 * 菜单 icon 字段存 IconPark outline 名（如 setting-two），按名取 SVG 渲染。
 */
import { computed } from 'vue'

import { resolveIconPark } from '@/lib/icon-park'

const props = defineProps<{ name?: string; size?: number | string }>()

const resolved = computed(() => resolveIconPark(props.name))

const px = computed(() => {
  const n = Number(props.size)
  return Number.isFinite(n) && n > 0 ? n : 16
})
</script>

<template>
  <svg
    class="menu-icon"
    xmlns="http://www.w3.org/2000/svg"
    :width="px"
    :height="px"
    :viewBox="`0 0 ${resolved.w} ${resolved.h}`"
    v-html="resolved.body"
  />
</template>

<style scoped>
.menu-icon {
  display: inline-block;
  vertical-align: middle;
  flex-shrink: 0;
  fill: currentColor;
}
</style>
