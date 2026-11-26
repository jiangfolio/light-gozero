<script setup lang="ts">
import RecursionMenu from './recursion-menu.vue'
import TopMenu from './top-menu.vue'

defineProps<{
  collapsed?: boolean
}>()
</script>

<template>
  <div class="menu-container">
    <TopMenu :collapsed="collapsed" class="top-menu-layer" />
    <RecursionMenu :collapsed="collapsed" class="recursion-menu" />
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.menu-container {
  height: 100%;
  position: relative;
}
</style>
