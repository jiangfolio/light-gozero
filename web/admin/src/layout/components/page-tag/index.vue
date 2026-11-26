<script setup lang="ts">
import { computed } from 'vue'

import { PAGE_TAG_STYLE } from '@/constants/layout'
import { useAppConfigStore } from '@/stores/app-config'

import AntdvTab from './antdv-tab.vue'
import ChromeTab from './chrome-tab.vue'

const appConfig = useAppConfigStore()

const styleMode = computed(() => appConfig.state.pageTagStyle)
</script>

<template>
  <div v-if="appConfig.state.pageTagFlag" class="page-tag-wrap" :data-style="styleMode">
    <!-- key 强制切换时整树重建，避免样式/遮罩残留 -->
    <AntdvTab v-if="styleMode === PAGE_TAG_STYLE.ANTDV" :key="'antdv'" />
    <ChromeTab v-else :key="'chrome'" />
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.page-tag-wrap {
  display: flex;
  width: 100%;
  min-width: 0;
  height: @page-tag-height;
  min-height: @page-tag-height;
  /* Chrome 必须贴底，弧形才能齐底边；其它样式垂直居中 */
  align-items: center;
  overflow: visible;
  border-bottom: 1px solid #f0f0f0;

  &[data-style='chrome'] {
    align-items: flex-end;
  }
}
</style>
