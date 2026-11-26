<script setup lang="ts">
import RecursionMenu from './recursion-menu.vue'
import TopMenu from './top-menu.vue'

defineProps<{
  collapsed?: boolean
}>()
</script>

<template>
  <div class="menu-container">
    <TopMenu class="top-menu" />
    <RecursionMenu v-show="!collapsed" class="recursion-menu" />
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.menu-container {
  display: flex;
  height: 100%;

  .top-menu {
    width: @side-expand-top-width;
    flex-shrink: 0;
  }

  .recursion-menu {
    min-width: 126px;
    flex: 1;
  }
}
</style>
