<script setup lang="ts">
import { computed } from 'vue'

import { useUserStore } from '@/stores/user'

defineOptions({ name: 'Home' })

const userStore = useUserStore()
const displayName = computed(
  () => userStore.userInfo?.nickname || userStore.userInfo?.username || '管理员',
)
</script>

<template>
  <div class="min-h-full">
    <ACard variant="borderless" class="rounded-lg shadow-sm" title="工作台">
      <p class="mb-4 text-[15px] text-black/85">
        你好，{{ displayName }}，欢迎使用 light-gozero 管理端。
      </p>
      <AAlert
        type="info"
        show-icon
        message="左侧菜单由登录后的 /user/resources 动态生成；按钮权限可用 v-privilege 控制。"
      />
    </ACard>
  </div>
</template>
