import { defineStore } from 'pinia'
import { reactive, watch } from 'vue'

import {
  DEFAULT_APP_CONFIG,
  PAGE_TAG_STYLE,
  type LayoutMode,
  type PageTagStyle,
} from '@/constants/layout'

const STORAGE_KEY = 'app_config'

type AppConfigState = typeof DEFAULT_APP_CONFIG

function loadConfig(): AppConfigState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULT_APP_CONFIG }
    const saved = JSON.parse(raw) as Partial<AppConfigState>
    // 圆角暂无设置入口，历史存档里的旧值会盖过新默认值，直接忽略
    delete saved.borderRadius
    // 页签样式历史值迁移：default/antd 已合并为 antdv，无法识别的值回退出厂默认
    if (saved.pageTagStyle !== undefined) {
      const legacy: Record<string, PageTagStyle> = {
        default: PAGE_TAG_STYLE.ANTDV,
        antd: PAGE_TAG_STYLE.ANTDV,
      }
      saved.pageTagStyle =
        legacy[saved.pageTagStyle]
        ?? (Object.values(PAGE_TAG_STYLE).includes(saved.pageTagStyle)
          ? saved.pageTagStyle
          : DEFAULT_APP_CONFIG.pageTagStyle)
    }
    return { ...DEFAULT_APP_CONFIG, ...saved }
  } catch {
    return { ...DEFAULT_APP_CONFIG }
  }
}

/**
 * 布局与主题配置；变更写入 localStorage。
 */
export const useAppConfigStore = defineStore('app-config', () => {
  const state = reactive<AppConfigState>(loadConfig())

  watch(
    state,
    () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
    },
    { deep: true },
  )

  function setLayout(layout: LayoutMode) {
    state.layout = layout
  }

  function setThemeColor(color: string) {
    state.themeColor = color
  }

  function setPageTagStyle(style: PageTagStyle) {
    state.pageTagStyle = style
  }

  function reset() {
    Object.assign(state, { ...DEFAULT_APP_CONFIG })
  }

  return { state, setLayout, setThemeColor, setPageTagStyle, reset }
})
