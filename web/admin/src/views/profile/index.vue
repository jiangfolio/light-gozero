<script setup lang="ts">
/**
 * 个人中心：左侧功能菜单 + 右侧内容面板。
 * 满高布局：页面撑满内容区，表单类面板内容滚动，日志类面板仅表体滚动。
 */
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import MenuIcon from '@/layout/components/menu-icon.vue'

import LoginLogPanel from './panels/login-log-panel.vue'
import OperLogPanel from './panels/oper-log-panel.vue'
import PasswordPanel from './panels/password-panel.vue'
import PlaceholderPanel from './panels/placeholder-panel.vue'
import ProfilePanel from './panels/profile-panel.vue'

defineOptions({ name: 'Profile' })

type PanelKey = 'profile' | 'password' | 'message' | 'notice' | 'login-log' | 'oper-log'

const route = useRoute()

const menus: { key: PanelKey; label: string; icon: string }[] = [
  { key: 'profile', label: '个人中心', icon: 'user' },
  { key: 'password', label: '修改密码', icon: 'lock' },
  { key: 'message', label: '我的消息', icon: 'mail' },
  { key: 'notice', label: '通知公告', icon: 'volume-notice' },
  { key: 'login-log', label: '登录日志', icon: 'log' },
  { key: 'oper-log', label: '操作日志', icon: 'history' },
]

/** 合法面板 key；非法值回退个人中心 */
function normalizeTab(raw: unknown): PanelKey {
  const hit = menus.find((m) => m.key === raw)
  return hit ? hit.key : 'profile'
}

const activeKey = ref<PanelKey>(normalizeTab(route.query.tab))

/** 已在个人中心时再点头像菜单，仅靠 query 变化切换面板 */
watch(
  () => route.query.tab,
  (tab) => {
    if (tab != null) activeKey.value = normalizeTab(tab)
  },
)

const activeMenu = computed(() => menus.find((m) => m.key === activeKey.value))
</script>

<template>
  <div class="profile-page">
    <ACard size="small" :bordered="false" class="profile-sider">
      <div class="sider-title">个人中心</div>
      <div
        v-for="m in menus"
        :key="m.key"
        class="sider-item"
        :class="{ active: m.key === activeKey }"
        @click="activeKey = m.key"
      >
        <MenuIcon :name="m.icon" :size="15" />
        <span class="sider-label">{{ m.label }}</span>
      </div>
    </ACard>

    <div class="profile-main">
      <ProfilePanel v-if="activeKey === 'profile'" />
      <PasswordPanel v-else-if="activeKey === 'password'" />
      <PlaceholderPanel
        v-else-if="activeKey === 'message'"
        title="我的消息"
        description="消息中心暂未接入"
      />
      <PlaceholderPanel
        v-else-if="activeKey === 'notice'"
        title="通知公告"
        description="通知公告暂未接入"
      />
      <LoginLogPanel v-else-if="activeKey === 'login-log'" />
      <OperLogPanel v-else-if="activeKey === 'oper-log'" />
      <AEmpty v-else :description="activeMenu?.label" />
    </div>
  </div>
</template>

<style scoped lang="less">
.profile-page {
  display: flex;
  height: 100%;
  min-height: 0;
  gap: 10px;
  padding-bottom: 10px;
}

.profile-sider {
  width: 180px;
  flex-shrink: 0;

  :deep(.ant-card-body) {
    padding: 8px;
  }
}

.sider-title {
  padding: 6px 10px 10px;
  font-size: 14px;
  font-weight: 600;
}

.sider-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  cursor: pointer;
  border-radius: var(--app-border-radius, 6px);
  color: #555;
  font-size: 13px;

  &:hover {
    background: rgba(0, 0, 0, 0.04);
  }

  &.active {
    background: var(--ant-color-primary-bg, #e6f4ff);
    color: var(--ant-color-primary, #1677ff);
    font-weight: 600;
  }
}

.profile-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}
</style>
