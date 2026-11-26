<script setup lang="ts">
import { computed } from 'vue'

import { useAppConfigStore } from '@/stores/app-config'
import { useUserStore } from '@/stores/user'

const appConfig = useAppConfigStore()
const userStore = useUserStore()

const content = computed(
  () => userStore.userInfo?.username || userStore.userInfo?.nickname || 'light-gozero',
)
</script>

<template>
  <AWatermark
    v-if="appConfig.state.watermarkFlag"
    class="h-full"
    :content="content"
    :font="{ color: 'rgba(0,0,0,0.06)', fontSize: 14 }"
    :gap="[140, 120]"
    :rotate="-22"
  >
    <slot />
  </AWatermark>
  <slot v-else />
</template>
