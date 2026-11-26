<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAppConfigStore } from '@/stores/app-config'

import HeaderUserSpace from './components/header-user-space.vue'
import LayoutFooter from './components/layout-footer.vue'
import MenuLocationBreadcrumb from './components/menu-location-breadcrumb.vue'
import PageContent from './components/page-content.vue'
import PageTag from './components/page-tag/index.vue'
import TopExpandMenu from './components/top-expand-menu/index.vue'
import { hasSideMenu } from './components/top-expand-menu/shared'
import Watermark from './components/watermark.vue'

const collapsed = ref(false)
const router = useRouter()
const appConfig = useAppConfigStore()

const tagInTop = computed(
  () => appConfig.state.pageTagFlag && appConfig.state.pageTagLocation === 'top',
)
const tagInCenter = computed(
  () => appConfig.state.pageTagFlag && appConfig.state.pageTagLocation === 'center',
)
const showBreadcrumb = computed(
  () => appConfig.state.breadcrumbFlag && appConfig.state.pageTagLocation === 'center',
)

const dueHeight = computed(() => {
  let due = 40
  if (tagInCenter.value) due += 40
  if (appConfig.state.footerFlag) due += 40
  return due
})

/** 无二级时 sider 宽为 0，仅保留覆盖层顶栏。 */
const siderWidth = computed(() => {
  if (!hasSideMenu.value) return 0
  return collapsed.value ? 80 : appConfig.state.menuWidth
})
</script>

<template>
  <ALayout class="top-expand-layout">
    <ALayoutSider
      theme="light"
      class="side-menu"
      :class="{ 'side-menu--hidden': !hasSideMenu }"
      :collapsed="collapsed && hasSideMenu"
      :collapsed-width="hasSideMenu ? 80 : 0"
      :width="siderWidth || 1"
      :trigger="null"
    >
      <TopExpandMenu :collapsed="collapsed" />
    </ALayoutSider>

    <ALayout class="top-expand-main">
      <ALayoutHeader class="top-expand-header">
        <div class="header-user">
          <div class="header-left">
            <button
              v-if="hasSideMenu"
              type="button"
              class="icon-btn"
              @click="collapsed = !collapsed"
            >
              <IconParkMenuUnfoldOne v-if="collapsed" size="18" />
              <IconParkMenuFoldOne v-else size="18" />
            </button>
            <ATooltip title="首页" placement="bottom">
              <button type="button" class="icon-btn" @click="router.push('/home')">
                <IconParkHome size="18" />
              </button>
            </ATooltip>
            <div class="location-breadcrumb">
              <PageTag v-if="tagInTop" />
              <MenuLocationBreadcrumb v-if="showBreadcrumb" />
            </div>
          </div>
          <HeaderUserSpace />
        </div>
        <PageTag v-if="tagInCenter" />
      </ALayoutHeader>

      <ALayoutContent
        class="top-expand-content"
        :style="{ height: `calc(100% - ${dueHeight}px)` }"
      >
        <Watermark>
          <PageContent />
        </Watermark>
      </ALayoutContent>

      <ALayoutFooter v-if="appConfig.state.footerFlag" class="top-expand-footer">
        <LayoutFooter />
      </ALayoutFooter>
      <AFloatBackTop :visibility-height="80" />
    </ALayout>
  </ALayout>
</template>

<style scoped lang="less">
@import './styles/layout-variables.less';

.top-expand-layout {
  min-height: 100vh;
  position: relative;
  background: #f0f2f5;
}

.side-menu {
  z-index: 10;
  height: 100vh;
  overflow: visible !important;

  &--hidden {
    max-width: 0 !important;
    min-width: 0 !important;
    flex: 0 0 0 !important;
  }
}

.top-expand-main {
  min-width: 0;
  height: 100vh;
  overflow: hidden;
  background: #f0f2f5;
  /* 给顶层覆盖菜单留空 */
  padding-top: @top-menu-height;
}

.top-expand-header {
  z-index: 21;
  height: auto !important;
  background: #fff !important;
  padding: 0 !important;
  line-height: normal !important;
}

.header-user {
  display: flex;
  height: @header-user-height;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0f0f0;
}

.header-left {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
}

.icon-btn {
  display: inline-flex;
  height: @header-user-height;
  cursor: pointer;
  align-items: center;
  border: 0;
  background: transparent;
  padding: 0 8px;
  color: rgba(0, 0, 0, 0.65);

  &:hover {
    color: var(--app-primary, #1677ff);
  }
}

.location-breadcrumb {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  padding-right: 8px;
}

.top-expand-content {
  min-height: 0;
  overflow: auto;
  background: inherit;
  padding: 5px 10px 0;
}

.top-expand-footer {
  background: transparent !important;
  padding: 6px 0 !important;
}
</style>
