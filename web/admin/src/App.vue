<script setup lang="ts">
import { theme } from 'antdv-next'
import zhCN from 'antdv-next/locale/zh_CN'
import { computed, watchEffect } from 'vue'

import { mixWithWhite } from '@/lib/color'
import { useAppConfigStore } from '@/stores/app-config'

const appConfig = useAppConfigStore()

const themeConfig = computed(() => ({
  algorithm: appConfig.state.compact ? theme.compactAlgorithm : theme.defaultAlgorithm,
  token: {
    colorPrimary: appConfig.state.themeColor,
    borderRadius: appConfig.state.borderRadius,
  },
}))

/** 全局弹窗：垂直居中；内容区定高纵向滚动；禁止横向滚动（Row gutter 负边距易撑出） */
const modalConfig = {
  centered: true,
  styles: {
    body: {
      maxHeight: 'calc(100vh - 200px)',
      overflowX: 'hidden' as const,
      overflowY: 'auto' as const,
    },
  },
}

watchEffect(() => {
  const root = document.documentElement
  root.style.setProperty('--app-primary', appConfig.state.themeColor)
  root.style.setProperty('--app-primary-bg', mixWithWhite(appConfig.state.themeColor, 0.88))
  // 暴露给非 antdv 的自定义元素（头部图标 hover 底等），保证圆角与主题 token 一致
  root.style.setProperty('--app-border-radius', `${appConfig.state.borderRadius}px`)
})
</script>

<template>
  <AConfigProvider :locale="zhCN" :theme="themeConfig" :modal="modalConfig">
    <router-view />
  </AConfigProvider>
</template>

