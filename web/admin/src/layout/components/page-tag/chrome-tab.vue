<script setup lang="ts">
import { computed } from 'vue'

import activeBg from '@/assets/images/nav/active_bg2.svg'
import activeBgHover from '@/assets/images/nav/active_bg2_default.svg'
import MenuIcon from '@/layout/components/menu-icon.vue'

import { usePageTag } from './use-page-tag'

const {
  pageTagFlag,
  tabsStore,
  activeKey,
  primary,
  primaryBg,
  selectTab,
  closeTab,
  operateMenu,
} = usePageTag()

const borderRadius = '8px'

const cssVars = computed(() => ({
  '--pt-primary': primary.value,
  '--pt-primary-bg': primaryBg.value,
  '--pt-mask-image': `url("${activeBg}")`,
  '--pt-mask-hover-image': `url("${activeBgHover}")`,
  '--pt-radius': borderRadius,
}))
</script>

<template>
  <!-- 结构对齐 smart-admin chrome-tab：左侧 tabs + 右侧 operate -->
  <div v-show="pageTagFlag" class="page-tag-row" :style="cssVars">
    <ADropdown :trigger="['contextmenu']" :menu="operateMenu">
      <div class="page-tag">
        <ATabs
          v-model:active-key="activeKey"
          style="width: 100%"
          tab-placement="top"
          size="small"
          @tab-click="selectTab"
        >
          <ATabPane v-for="item in tabsStore.tabs" :key="item.path">
            <template #tab>
              <span class="page-tag-content">
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

/*
 * Chrome 页签：按 smart-admin 移植，几何上有两处修正。
 * 1) 整行贴底（flex-end），弧形才能贴齐底边线，居中会留白缝并造成「往上顶」错觉。
 * 2) 关闭钮留在文档流内、margin-left:auto 顶到内容区右缘；内容区 padding-right:13px 让关闭钮
 *    右缘内收，保证在统一的 -20px 页签重叠下：
 *    - 前邻页签的 × 不会被本页签激活/hover 的左弧形盖住（留 7px 空隙）；
 *    - 后邻页签的盒子不会罩住本页签的 ×（否则 × 点击落空、图标压字）；
 *    - 后邻图标恰好起始于右弧形收尾处（与真 Chrome 一致）。
 * 3) 参考里 active+1 有 margin-left:-50px 的特殊重叠，但那会让切换激活页时后续页签整体位移；
 *    这里全部页签统一 margin:0 -10px，切换零位移。
 * 4) 父级白底时弧形不能用 z-index:-1，改 isolation + 正层级。
 */
.page-tag-row {
  position: relative;
  display: flex;
  width: 100%;
  height: @page-tag-height;
  min-height: @page-tag-height;
  align-items: flex-end;
}

.page-tag-operate {
  box-sizing: border-box;
  display: flex;
  width: @page-tag-operate-width;
  height: @page-tag-height;
  flex-shrink: 0;
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
  align-items: flex-end;
  justify-content: space-between;
  height: @page-tag-height;
  min-height: @page-tag-height;
  padding-right: 20px;
  padding-left: 20px;
  user-select: none;
  width: calc(100% - @page-tag-operate-width);

  .page-tag-icon {
    display: inline-flex;
    flex-shrink: 0;
    margin-right: 5px;
    color: #666;
    line-height: 1;
    vertical-align: middle;

    :deep(.menu-icon),
    :deep(svg) {
      display: block;
      line-height: 1;
    }
  }

  .page-tag-title {
    /* 超长标题省略，不允许撑破页签或换行 */
    max-width: 120px;
    overflow: hidden;
    line-height: 1;
    text-overflow: ellipsis;
    vertical-align: middle;
    white-space: nowrap;
  }

  /* 关闭钮留在流内，margin-left:auto 顶到内容区右缘，不占额外宽度 */
  .page-tag-close {
    display: block;
    flex-shrink: 0;
    /* auto 顶到右缘；padding-left 保证长标题时与文字至少 5px 间隔，总占位与参考一致 */
    margin-left: auto;
    padding-left: 5px;
    color: #666;
    line-height: 1;
    cursor: pointer;
  }

  :deep(.ant-tabs) {
    width: 100%;
  }

  :deep(.ant-tabs-content-holder) {
    display: none;
  }

  :deep(.ant-tabs-nav) {
    margin: 0;
    min-height: 35px;
    min-width: 120px;
    box-sizing: border-box;

    &::before {
      border-bottom: none !important;
    }
  }

  /* 与参考一致的 padding；禁止 background shorthand */
  :deep(.ant-tabs-nav .ant-tabs-tab) {
    position: relative;
    padding: 5px 18px 3px 24px !important;
    border-radius: var(--pt-radius) var(--pt-radius) 0 0;
    margin: 0 -10px !important;
    border: none !important;
    line-height: 1 !important;

    &:nth-child(1) {
      margin-left: 0 !important;
    }

    &:nth-last-child(2) {
      margin-right: 0 !important;
    }
  }

  :deep(.ant-tabs-tab-btn) {
    display: inline-flex !important;
    align-items: center !important;
    line-height: 1 !important;
    background: transparent !important;
  }

  .page-tag-content {
    position: relative;
    display: inline-flex;
    /* border-box 下 min-width 含 padding，短标题页签总宽仍与参考一致（142px） */
    box-sizing: border-box;
    min-width: 100px;
    align-items: center;
    /* 关闭钮右缘内收 13px，是给前后邻页签留的重叠安全距离，见文件头注释 */
    padding-right: 13px;
    /* 内收后 min-width 剩余宽度放不下「图标+4字标题+×」，禁止 CJK 换行，宁可盒子自然撑宽 */
    white-space: nowrap;
    line-height: 1;

    &::after {
      content: '';
      position: absolute;
      top: 50%;
      right: -9px;
      z-index: -2;
      width: 1px;
      height: 16px;
      margin-top: -8px;
      background: #eeeeee;
    }
  }

  :deep(.ant-tabs-tab-active) {
    position: relative;
    z-index: 0;
    isolation: isolate;
    background-color: transparent !important;
    background-image: linear-gradient(var(--pt-primary-bg), var(--pt-primary-bg)) !important;
    /* 中间填充锚底且与两翼同高 35px；写 100% 会高出 3px，形成「中间突出」 */
    background-position: center bottom !important;
    background-size: 60% 35px !important;
    background-repeat: no-repeat !important;

    &::before,
    &::after {
      content: '';
      position: absolute;
      bottom: 0;
      z-index: 0;
      width: 50%;
      height: 35px;
      pointer-events: none;
      background-color: var(--pt-primary-bg);
      -webkit-mask-image: var(--pt-mask-image);
      -webkit-mask-repeat: no-repeat;
      -webkit-mask-position: left;
      -webkit-mask-size: auto;
      mask-image: var(--pt-mask-image);
      mask-repeat: no-repeat;
      mask-position: left;
      mask-size: auto;
    }

    &::before {
      left: -4px;
    }

    &::after {
      right: -4px;
      transform: scaleX(-1);
    }

    .page-tag-content {
      z-index: 1;

      &::after {
        display: none;
      }
    }

    .page-tag-close,
    .page-tag-icon {
      color: var(--pt-primary);
    }
  }

  :deep(.ant-tabs-ink-bar) {
    display: none;
  }

  :deep(.ant-tabs-nav .ant-tabs-tab:hover) {
    &:not(.ant-tabs-tab-active) {
      position: relative;
      z-index: 0;
      isolation: isolate;
      background-color: transparent !important;
      background-image: linear-gradient(var(--pt-primary-bg), var(--pt-primary-bg)) !important;
      /* 与激活态一致：中间填充锚底、与两翼同高 35px */
      background-position: center bottom !important;
      background-size: 60% 35px !important;
      background-repeat: no-repeat !important;

      &::before,
      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        z-index: 0;
        width: 50%;
        height: 35px;
        pointer-events: none;
        background-color: var(--pt-primary-bg);
        -webkit-mask-image: var(--pt-mask-hover-image);
        -webkit-mask-repeat: no-repeat;
        -webkit-mask-position: left;
        -webkit-mask-size: auto;
        mask-image: var(--pt-mask-hover-image);
        mask-repeat: no-repeat;
        mask-position: left;
        mask-size: auto;
      }

      &::before {
        left: -4px;
      }

      &::after {
        right: -4px;
        transform: scaleX(-1);
      }

      .page-tag-content {
        z-index: 1;
        color: var(--pt-primary);

        &::after {
          display: none;
        }
      }
    }

    .page-tag-close,
    .page-tag-icon {
      color: var(--pt-primary);
    }
  }
}
</style>
