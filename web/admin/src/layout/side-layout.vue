<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAppConfigStore } from '@/stores/app-config'

import HeaderUserSpace from './components/header-user-space.vue'
import LayoutFooter from './components/layout-footer.vue'
import MenuLocationBreadcrumb from './components/menu-location-breadcrumb.vue'
import PageContent from './components/page-content.vue'
import PageTag from './components/page-tag/index.vue'
import SideMenu from './components/side-menu.vue'
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
  // 1. 用户行 40；内容上方页签再加 40；页脚再加 40
  let due = 40
  if (tagInCenter.value) due += 40
  if (appConfig.state.footerFlag) due += 40
  return due
})
</script>

<template>
  <ALayout class="side-layout">
    <ALayoutSider
      class="side-menu"
      :width="appConfig.state.menuWidth"
      :collapsed="collapsed"
      :collapsed-width="64"
      :theme="appConfig.state.menuTheme"
      collapsible
      :trigger="null"
    >
      <SideMenu :collapsed="collapsed" />
    </ALayoutSider>

    <ALayout class="side-layout-main">
      <ALayoutHeader class="side-layout-header">
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
          <div class="header-right">
            <HeaderUserSpace />
          </div>
        </div>
        <PageTag v-if="tagInCenter" />
      </ALayoutHeader>

      <ALayoutContent
        class="side-layout-content"
        :style="{ height: `calc(100% - ${dueHeight}px)` }"
      >
        <Watermark>
          <div class="content-inner">
            <PageContent />
          </div>
        </Watermark>
      </ALayoutContent>

      <ALayoutFooter v-if="appConfig.state.footerFlag" class="side-layout-footer">
        <LayoutFooter />
      </ALayoutFooter>

      <AFloatBackTop :visibility-height="80" />
    </ALayout>
  </ALayout>
</template>

<style scoped lang="less">
@import './styles/layout-variables.less';

.side-layout {
  min-height: 100vh;
}

.side-menu {
  z-index: 10;
  height: 100vh;
  overflow-x: hidden;
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.2);
  }
}

.side-layout-main {
  height: 100vh;
  overflow: hidden;
  background: #f0f2f5;
}

.side-layout-header {
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
  height: @header-user-height;
  min-width: 0;
  flex: 1;
  align-items: center;
}

.icon-btn {
  display: inline-flex;
  height: @header-user-height;
  cursor: pointer;
  align-items: center;
  margin-left: 10px;
  border: 0;
  background: transparent;
  padding: 0 6px;
  color: rgba(0, 0, 0, 0.65);

  &:hover {
    color: var(--app-primary, #1677ff);
  }
}

.location-breadcrumb {
  margin-left: 8px;
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  overflow: hidden;
  padding-right: 8px;
}

.header-right {
  display: flex;
  height: @header-user-height;
  flex-shrink: 0;
  align-items: center;
}

.side-layout-content {
  position: relative;
  min-height: 0;
  overflow: auto;
  background: inherit;
  padding: 5px 10px 0;
}

.content-inner {
  /* 定高：页面根 h-full 生效，由页面自己控制内部滚动（双卡片骨架） */
  height: 100%;
}

.side-layout-footer {
  display: flex;
  justify-content: center;
  background: transparent !important;
  padding: 6px 0 !important;
}
</style>
