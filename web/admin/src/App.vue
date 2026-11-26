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

watchEffect(() => {
  const root = document.documentElement
  root.style.setProperty('--app-primary', appConfig.state.themeColor)
  root.style.setProperty('--app-primary-bg', mixWithWhite(appConfig.state.themeColor, 0.88))
  // 暴露给非 antdv 的自定义元素（头部图标 hover 底等），保证圆角与主题 token 一致
  root.style.setProperty('--app-border-radius', `${appConfig.state.borderRadius}px`)
})
</script>

<template>
  <AConfigProvider :locale="zhCN" :theme="themeConfig">
    <router-view />
  </AConfigProvider>
</template>

