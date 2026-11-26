<script setup lang="ts">
import { computed } from 'vue'

import { useAppConfigStore } from '@/stores/app-config'

import AppLogo from './components/app-logo.vue'
import HeaderUserSpace from './components/header-user-space.vue'
import LayoutFooter from './components/layout-footer.vue'
import MenuLocationBreadcrumb from './components/menu-location-breadcrumb.vue'
import PageContent from './components/page-content.vue'
import PageTag from './components/page-tag/index.vue'
import TopMenu from './components/top-menu.vue'
import Watermark from './components/watermark.vue'

const appConfig = useAppConfigStore()

/** 顶栏模式：页签/面包屑固定在菜单条下方，不跟 location 拆装。 */
const showPageTag = computed(() => appConfig.state.pageTagFlag)
const showBreadcrumb = computed(() => appConfig.state.breadcrumbFlag)

const dueHeight = computed(() => {
  let due = 48
  if (showPageTag.value) due += 40
  if (showBreadcrumb.value) due += 40
  if (appConfig.state.footerFlag) due += 40
  return due
})
</script>

<template>
  <ALayout class="top-layout">
    <ALayoutHeader class="top-layout-menu">
      <AppLogo dark class="logo" />
      <TopMenu />
      <div class="top-layout-user">
        <HeaderUserSpace tone="dark" />
      </div>
    </ALayoutHeader>

    <div v-if="showPageTag" class="page-tag-bar">
      <PageTag />
    </div>

    <div v-if="showBreadcrumb" class="bread-crumb-bar">
      <MenuLocationBreadcrumb />
    </div>

    <ALayoutContent
      class="top-layout-content"
      :style="{ height: `calc(100vh - ${dueHeight}px)` }"
    >
      <Watermark>
        <PageContent />
      </Watermark>
    </ALayoutContent>

    <ALayoutFooter v-if="appConfig.state.footerFlag" class="top-layout-footer">
      <LayoutFooter />
    </ALayoutFooter>
  </ALayout>
</template>

<style scoped lang="less">
@import './styles/layout-variables.less';

.top-layout {
  min-height: 100vh;
  background: #f0f2f5;
}

.top-layout-menu {
  display: flex !important;
  height: @top-menu-height !important;
  align-items: center !important;
  gap: 12px;
  background: #001529 !important;
  padding: 0 16px !important;
  line-height: @top-menu-height !important;
  color: #fff;
  overflow: hidden;
}

.logo {
  display: flex !important;
  height: @top-menu-height !important;
  min-height: @top-menu-height !important;
  align-items: center !important;
  padding-left: 0 !important;
  padding-right: 8px !important;
}

.top-layout-user {
  display: flex;
  height: @top-menu-height;
  flex-shrink: 0;
  align-items: center;
  margin-left: auto;
}

.page-tag-bar {
  display: flex;
  height: @page-tag-height;
  min-height: @page-tag-height;
  align-items: flex-end;
  overflow: visible;
  background: #fff;
}

.bread-crumb-bar {
  display: flex;
  height: @header-user-height;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;
  padding: 0 12px;
}

.top-layout-content {
  /* 定高 + 内部滚动：页面根 h-full 生效，双卡片骨架只在表体内滚动 */
  min-height: 0;
  overflow: auto;
  background: inherit;
  padding: 5px 10px 0;
}

.top-layout-footer {
  background: transparent !important;
  padding: 6px 0 !important;
}
</style>
