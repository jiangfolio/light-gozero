<script setup lang="ts">
import zhCN from 'antdv-next/locale/zh_CN'
</script>

<template>
  <AConfigProvider :locale="zhCN">
    <router-view />
  </AConfigProvider>
</template>
