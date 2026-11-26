<script setup lang="ts">
import { LAYOUT_ENUM, PAGE_TAG_STYLE, THEME_COLORS } from '@/constants/layout'
import { useAppConfigStore } from '@/stores/app-config'

const open = defineModel<boolean>('open', { default: false })
const appConfig = useAppConfigStore()
</script>

<template>
  <!-- 横向表单 + 四段按钮需要更宽容器，避免控件被裁切 -->
  <ADrawer v-model:open="open" title="布局设置" :size="480" placement="right">
    <AForm layout="horizontal" :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
      <AFormItem label="布局模式">
        <ARadioGroup
          v-model:value="appConfig.state.layout"
          button-style="solid"
          option-type="button"
          :options="[
            { label: '侧边', value: LAYOUT_ENUM.SIDE },
            { label: '侧边展开', value: LAYOUT_ENUM.SIDE_EXPAND },
            { label: '顶部', value: LAYOUT_ENUM.TOP },
            { label: '顶部展开', value: LAYOUT_ENUM.TOP_EXPAND },
          ]"
        />
      </AFormItem>

      <AFormItem label="主题色">
        <div class="theme-colors">
          <button
            v-for="c in THEME_COLORS"
            :key="c"
            type="button"
            class="theme-color-block"
            :style="{ background: c }"
            @click="appConfig.setThemeColor(c)"
          >
            <IconParkCheckSmall
              v-if="appConfig.state.themeColor === c"
              :size="14"
              class="theme-color-check"
            />
          </button>
        </div>
      </AFormItem>

      <AFormItem label="菜单主题">
        <ARadioGroup v-model:value="appConfig.state.menuTheme" button-style="solid">
          <ARadioButton value="dark">深色</ARadioButton>
          <ARadioButton value="light">浅色</ARadioButton>
        </ARadioGroup>
      </AFormItem>

      <AFormItem label="页签样式">
        <ARadioGroup
          v-model:value="appConfig.state.pageTagStyle"
          button-style="solid"
          option-type="button"
          :options="[
            { label: 'Antdv', value: PAGE_TAG_STYLE.ANTDV },
            { label: 'Chrome', value: PAGE_TAG_STYLE.CHROME },
          ]"
        />
      </AFormItem>

      <AFormItem label="页签位置">
        <ARadioGroup v-model:value="appConfig.state.pageTagLocation" button-style="solid">
          <ARadioButton value="top">顶栏内</ARadioButton>
          <ARadioButton value="center">内容上方</ARadioButton>
        </ARadioGroup>
      </AFormItem>

      <AFormItem label="显示页签">
        <ASwitch
          v-model:checked="appConfig.state.pageTagFlag"
          checked-children="显示"
          un-checked-children="隐藏"
        />
      </AFormItem>
      <AFormItem label="显示面包屑">
        <ASwitch
          v-model:checked="appConfig.state.breadcrumbFlag"
          checked-children="显示"
          un-checked-children="隐藏"
        />
      </AFormItem>
      <AFormItem label="显示页脚">
        <ASwitch
          v-model:checked="appConfig.state.footerFlag"
          checked-children="显示"
          un-checked-children="隐藏"
        />
      </AFormItem>
      <AFormItem label="水印">
        <ASwitch
          v-model:checked="appConfig.state.watermarkFlag"
          checked-children="显示"
          un-checked-children="隐藏"
        />
      </AFormItem>
      <AFormItem label="菜单手风琴">
        <ASwitch
          v-model:checked="appConfig.state.menuAccordion"
          checked-children="开"
          un-checked-children="关"
        />
      </AFormItem>
      <AFormItem
        v-if="
          appConfig.state.layout === LAYOUT_ENUM.SIDE ||
          appConfig.state.layout === LAYOUT_ENUM.SIDE_EXPAND ||
          appConfig.state.layout === LAYOUT_ENUM.TOP_EXPAND
        "
        :label="`菜单宽度 ${appConfig.state.menuWidth}`"
      >
        <ASlider v-model:value="appConfig.state.menuWidth" :min="160" :max="280" :step="10" />
      </AFormItem>
    </AForm>

    <AButton block danger type="primary" @click="appConfig.reset()">恢复默认配置</AButton>
  </ADrawer>
</template>

<style scoped lang="less">
.theme-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.theme-color-block {
  display: flex;
  width: 22px;
  height: 22px;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 2px;
  padding: 0;
}

.theme-color-check {
  color: #fff;
}
</style>
