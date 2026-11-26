<script setup lang="ts">
import { computed } from 'vue'

import MenuIcon from '@/layout/components/menu-icon.vue'

import { usePageTag } from './use-page-tag'

const {
  pageTagFlag,
  tabsStore,
  activeKey,
  primary,
  primaryBg,
  borderRadius,
  selectTab,
  closeTab,
  operateMenu,
} = usePageTag()

/** 用 CSS 变量驱动主题色，避免 Less 编译吃掉 v-bind 导致默认样式不起作用 */
const cssVars = computed(() => ({
  '--pt-primary': primary.value,
  '--pt-primary-bg': primaryBg.value,
  '--pt-radius': borderRadius.value,
}))
</script>

<template>
  <div v-show="pageTagFlag" class="page-tag-row" :style="cssVars">
    <div class="page-tag-main">
      <ADropdown :trigger="['contextmenu']" :menu="operateMenu">
        <div class="page-tag">
          <ATabs
            v-model:active-key="activeKey"
            tab-position="top"
            size="small"
            @tab-click="selectTab"
          >
            <ATabPane v-for="item in tabsStore.tabs" :key="item.path">
              <template #tab>
                <span class="page-tag-label">
                  <MenuIcon
                    v-if="item.icon || item.affix"
                    class="page-tag-icon"
                    :name="item.icon || 'Home'"
                    :size="12"
                  />
                  <span class="page-tag-title">{{ item.title }}</span>
                  <IconParkCloseSmall
                    v-if="!item.affix"
                    class="page-tag-close"
                    :size="12"
                    @click.stop="closeTab(item.path, $event)"
                  />
                </span>
              </template>
            </ATabPane>
          </ATabs>
        </div>
      </ADropdown>
    </div>

    <ADropdown :menu="operateMenu">
      <div class="page-tag-operate">
        <div class="page-tag-operate-icon">
          <IconParkAllApplication :size="18" />
        </div>
      </div>
    </ADropdown>
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.page-tag-row {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  min-height: @page-tag-height;
}

.page-tag-main {
  flex: 1;
  min-width: 0;
  width: calc(100% - @page-tag-operate-width);
}

.page-tag-operate {
  box-sizing: border-box;
  display: flex;
  flex-shrink: 0;
  width: @page-tag-operate-width;
  height: @page-tag-height;
  align-items: center;
  justify-content: center;
  padding-right: 10px;
  font-size: 17px;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.65);

  .page-tag-operate-icon {
    display: inline-flex;
    width: 20px;
    height: 20px;
    align-items: center;
    justify-content: center;
  }

  &:hover {
    color: var(--pt-primary);
  }
}

.page-tag {
  position: relative;
  box-sizing: border-box;
  display: flex;
  width: 100%;
  min-height: @page-tag-height;
  align-items: center;
  padding-right: 20px;
  padding-left: 20px;
  user-select: none;

  .page-tag-label {
    position: relative;
    display: inline-flex;
    box-sizing: border-box;
    min-width: 100px;
    align-items: center;
    padding-right: 16px;
    white-space: nowrap;
    line-height: 1;
  }

  .page-tag-icon {
    display: inline-flex;
    flex-shrink: 0;
    margin-right: 4px;
    color: #666;
    line-height: 1;

    :deep(svg) {
      display: block;
    }
  }

  .page-tag-title {
    /* 超长标题省略，不允许撑破页签或换行 */
    max-width: 120px;
    overflow: hidden;
    line-height: 1;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-tag-close {
    position: absolute;
    top: 50%;
    right: 0;
    display: inline-flex;
    color: #666;
    line-height: 1;
    cursor: pointer;
    transform: translateY(-50%);

    :deep(svg) {
      display: block;
    }
  }

  :deep(.ant-tabs) {
    width: 100%;
  }

  :deep(.ant-tabs-nav) {
    /* 撑满整行高度，激活页签淡蓝底与底部 ink-bar 才能贴到行底灰线 */
    height: @page-tag-height;
    margin: 0 !important;
    padding: 0 !important;

    &::before {
      border-bottom: none !important;
    }
  }

  :deep(.ant-tabs-content-holder) {
    display: none !important;
  }

  :deep(.ant-tabs-tab) {
    height: @page-tag-height !important;
    margin: 0 0 0 5px !important;
    padding: 0 12px !important;
    border: none !important;
    border-radius: var(--pt-radius) !important;
    background: transparent !important;
    transition: background-color 0.2s;
  }

  :deep(.ant-tabs-tab-btn) {
    display: inline-flex !important;
    width: 100%;
    align-items: center !important;
    line-height: 1 !important;
    color: inherit !important;
  }

  :deep(.ant-tabs-tab-active) {
    background-color: var(--pt-primary-bg) !important;

    .page-tag-icon,
    .page-tag-close,
    .page-tag-title {
      color: var(--pt-primary) !important;
    }
  }

  :deep(.ant-tabs-tab:hover) {
    background-color: var(--pt-primary-bg) !important;

    .page-tag-icon,
    .page-tag-close {
      color: var(--pt-primary) !important;
    }
  }
}
</style>
