import { defineStore } from 'pinia'
import { ref } from 'vue'

export type TabItem = {
  path: string
  fullPath: string
  name: string
  title: string
  /** 菜单图标名，与 route.meta.icon / 侧栏 MenuIcon 一致 */
  icon?: string
  affix?: boolean
}

/**
 * 多页签状态：打开、关闭、关闭其他/全部。
 */
export const useTabsStore = defineStore('tabs', () => {
  const tabs = ref<TabItem[]>([])
  const activePath = ref('')

  function openTab(tab: TabItem) {
    // 1. 已存在则激活，并回写 title/icon；否则追加
    const exist = tabs.value.find((t) => t.path === tab.path)
    if (!exist) {
      tabs.value.push(tab)
    } else {
      if (tab.title) exist.title = tab.title
      if (tab.fullPath) exist.fullPath = tab.fullPath
      if (tab.icon) exist.icon = tab.icon
    }
    activePath.value = tab.path
  }

  function closeTab(path: string) {
    const idx = tabs.value.findIndex((t) => t.path === path)
    if (idx < 0) return
    const item = tabs.value[idx]
    if (item?.affix) return
    tabs.value.splice(idx, 1)
    if (activePath.value === path) {
      const next = tabs.value[idx] || tabs.value[idx - 1]
      activePath.value = next?.path || ''
    }
  }

  function closeOthers(path: string) {
    tabs.value = tabs.value.filter((t) => t.affix || t.path === path)
    activePath.value = path
  }

  function closeAll() {
    tabs.value = tabs.value.filter((t) => t.affix)
    activePath.value = tabs.value[0]?.path || ''
  }

  function reset() {
    tabs.value = []
    activePath.value = ''
  }

  return { tabs, activePath, openTab, closeTab, closeOthers, closeAll, reset }
})
