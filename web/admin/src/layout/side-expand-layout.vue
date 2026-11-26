<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAppConfigStore } from '@/stores/app-config'

import HeaderUserSpace from './components/header-user-space.vue'
import LayoutFooter from './components/layout-footer.vue'
import MenuLocationBreadcrumb from './components/menu-location-breadcrumb.vue'
import PageContent from './components/page-content.vue'
import PageTag from './components/page-tag/index.vue'
import SideExpandMenu from './components/side-expand-menu/index.vue'
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
  () => appConfig.state.breadcrumbFlag && appConfig.state.pageTagLocation !== 'top',
)

const dueHeight = computed(() => {
  let due = 40
  if (tagInCenter.value) due += 40
  if (appConfig.state.footerFlag) due += 40
  return due
})

const siderWidth = computed(() => {
  // 1. 折叠仅保留一级 114；展开 = 114 + 二级宽度（menuWidth 作为二级+一级总量近似）
  if (collapsed.value) return 114
  return Math.max(appConfig.state.menuWidth, 240)
})
</script>

<template>
  <ALayout class="side-expand-layout">
    <ALayoutSider
      class="side-menu"
      :theme="appConfig.state.menuTheme"
      :collapsed="collapsed"
      :collapsed-width="114"
      :width="siderWidth"
      :trigger="null"
      collapsible
    >
      <SideExpandMenu :collapsed="collapsed" />
    </ALayoutSider>

    <ALayout class="side-expand-main">
      <ALayoutHeader class="side-expand-header">
        <div class="header-user">
          <div class="header-left">
            <button type="button" class="icon-btn" @click="collapsed = !collapsed">
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
        class="side-expand-content"
        :style="{ height: `calc(100% - ${dueHeight}px)` }"
      >
        <Watermark>
          <PageContent />
        </Watermark>
      </ALayoutContent>

      <ALayoutFooter v-if="appConfig.state.footerFlag" class="side-expand-footer">
        <LayoutFooter />
      </ALayoutFooter>
      <AFloatBackTop :visibility-height="80" />
    </ALayout>
  </ALayout>
</template>

<style scoped lang="less">
@import './styles/layout-variables.less';

.side-expand-layout {
  min-height: 100vh;
}

.side-menu {
  overflow: auto;
  height: 100vh;
}

.side-expand-main {
  min-width: 0;
  height: 100vh;
  overflow: hidden;
  background: #f0f2f5;
}

.side-expand-header {
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

.side-expand-content {
  min-width: 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  background: inherit;
  padding: 5px 10px 0;
}

.side-expand-footer {
  background: transparent !important;
  padding: 6px 0 !important;
}
</style>
