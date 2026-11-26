<script setup lang="ts">
import { computed } from 'vue'

import { LAYOUT_ENUM } from '@/constants/layout'
import { useAppConfigStore } from '@/stores/app-config'

import SideExpandLayout from './side-expand-layout.vue'
import SideLayout from './side-layout.vue'
import TopExpandLayout from './top-expand-layout.vue'
import TopLayout from './top-layout.vue'

const appConfig = useAppConfigStore()
const layout = computed(() => appConfig.state.layout)
</script>

<template>
  <SideLayout v-if="layout === LAYOUT_ENUM.SIDE" />
  <SideExpandLayout v-else-if="layout === LAYOUT_ENUM.SIDE_EXPAND" />
  <TopLayout v-else-if="layout === LAYOUT_ENUM.TOP" />
  <TopExpandLayout v-else />
</template>
