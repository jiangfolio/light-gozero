<script setup lang="ts">
import { message } from 'antdv-next'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { resetDynamicRoutes } from '@/router/guard'
import router from '@/router'
import { useUserStore } from '@/stores/user'

import NotificationPopover from './notification-popover.vue'
import SettingDrawer from './setting-drawer.vue'

const props = defineProps<{ tone?: 'light' | 'dark' }>()

const vueRouter = useRouter()
const userStore = useUserStore()
const settingOpen = ref(false)

const displayName = computed(
  () => userStore.userInfo?.nickname || userStore.userInfo?.username || '用户',
)

const isDark = computed(() => props.tone === 'dark')

async function onLogout() {
  // 先按名字移除动态路由再清会话：logout 内的 permission.reset() 会清空路由名列表，顺序不能反
  resetDynamicRoutes(router)
  await userStore.logout(true)
  message.success('已退出登录')
  await vueRouter.replace('/login')
}

function onMenuClick(info: { key: string | number }) {
  const key = String(info.key)
  if (key === 'logout') {
    onLogout()
    return
  }
  // 个人中心/修改密码：进个人中心页并定位到对应面板
  if (key === 'profile' || key === 'password') {
    vueRouter.push({ path: '/profile', query: key === 'password' ? { tab: 'password' } : {} })
  }
}

const userMenuItems = [
  { key: 'profile', label: '个人中心' },
  { key: 'password', label: '修改密码' },
  { type: 'divider' as const },
  { key: 'logout', label: '退出登录' },
]
</script>

<template>
  <div class="flex h-full items-center gap-0.5 pr-3">
    <NotificationPopover :tone="props.tone" />
    <button
      type="button"
      title="布局设置"
      class="inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-[var(--app-border-radius)]"
      :class="
        isDark
          ? 'text-white/85 hover:bg-white/10 hover:text-white'
          : 'text-black/65 hover:bg-black/5 hover:text-[#1677ff]'
      "
      @click="settingOpen = true"
    >
      <IconParkSettingTwo size="18" />
    </button>
    <ADropdown :menu="{ items: userMenuItems, onClick: onMenuClick }">
      <span
        class="ml-0.5 inline-flex cursor-pointer items-center gap-2 rounded-[var(--app-border-radius)] px-2 py-1"
        :class="isDark ? 'hover:bg-white/10' : 'hover:bg-black/5'"
      >
        <AAvatar :size="28" style="background: #1677ff">{{ displayName.slice(0, 1) }}</AAvatar>
        <span
          class="max-w-[110px] truncate text-sm"
          :class="isDark ? 'text-white/85' : 'text-black/85'"
        >
          {{ displayName }}
        </span>
      </span>
    </ADropdown>
    <SettingDrawer v-model:open="settingOpen" />
  </div>
</template>
