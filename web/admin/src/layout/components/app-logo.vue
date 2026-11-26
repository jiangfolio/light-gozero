<script setup lang="ts">
import { useRouter } from 'vue-router'

withDefaults(
  defineProps<{
    collapsed?: boolean
    dark?: boolean
    clickable?: boolean
  }>(),
  { collapsed: false, dark: false, clickable: true },
)

const router = useRouter()

function onClick() {
  router.push('/home')
}
</script>

<template>
  <div
    class="app-logo flex h-10 shrink-0 cursor-pointer items-center justify-center gap-2 px-3"
    :class="{ 'px-0': collapsed }"
    @click="clickable && onClick()"
  >
    <span
      class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[#1677ff] text-sm font-bold text-white shadow-sm"
    >
      L
    </span>
    <span
      v-show="!collapsed"
      class="truncate text-base font-semibold"
      :class="dark ? 'text-white' : 'text-[#001529]'"
    >
      light-gozero
    </span>
  </div>
</template>

<style scoped lang="less">
@import '../styles/layout-variables.less';

.app-logo {
  height: @header-user-height;
  min-height: @header-user-height;
}
</style>
