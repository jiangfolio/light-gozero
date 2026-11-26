import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { mixWithWhite } from '@/lib/color'
import { useAppConfigStore } from '@/stores/app-config'
import { useTabsStore } from '@/stores/tabs'

/** 页签公共逻辑：同步路由、切换/关闭、主题色变量。 */
export function usePageTag() {
  const route = useRoute()
  const router = useRouter()
  const tabsStore = useTabsStore()
  const appConfig = useAppConfigStore()

  const pageTagFlag = computed(() => appConfig.state.pageTagFlag)
  const primary = computed(() => appConfig.state.themeColor)
  const primaryBg = computed(() => mixWithWhite(appConfig.state.themeColor, 0.88))
  const borderRadius = computed(() => `${appConfig.state.borderRadius}px`)

  watch(
    () => route.fullPath,
    () => {
      if (!route.name || route.meta.public) return
      tabsStore.openTab({
        path: route.path,
        fullPath: route.fullPath,
        name: String(route.name),
        title: String(route.meta.title || route.name),
        icon: route.meta.icon ? String(route.meta.icon) : undefined,
        affix: Boolean(route.meta.affix),
      })
    },
    { immediate: true },
  )

  const activeKey = computed({
    get: () => tabsStore.activePath || route.path,
    set: (path: string) => selectTab(path),
  })

  function selectTab(path: string | number) {
    const key = String(path)
    if (tabsStore.activePath === key) return
    const tab = tabsStore.tabs.find((t) => t.path === key)
    router.push(tab?.fullPath || key)
  }

  function closeTab(path: string, e?: Event) {
    e?.stopPropagation()
    e?.preventDefault()
    const closingActive = tabsStore.activePath === path
    tabsStore.closeTab(path)
    if (closingActive && tabsStore.activePath) {
      selectTab(tabsStore.activePath)
    }
  }

  function closeByMenu(closeAll: boolean) {
    const current = tabsStore.activePath
    if (closeAll || !current) {
      tabsStore.closeAll()
      if (tabsStore.activePath) selectTab(tabsStore.activePath)
      else router.push('/home')
      return
    }
    tabsStore.closeOthers(current)
    selectTab(current)
  }

  function onContext(key: string, path?: string) {
    const target = path || tabsStore.activePath
    if (!target) return
    if (key === 'close') closeTab(target)
    if (key === 'others') {
      tabsStore.closeOthers(target)
      selectTab(target)
    }
    if (key === 'all') closeByMenu(true)
  }

  const operateMenu = {
    items: [
      { key: 'others', label: '关闭其他' },
      { key: 'all', label: '关闭全部' },
    ],
    onClick: (info: { key: string | number }) => onContext(String(info.key)),
  }

  return {
    pageTagFlag,
    tabsStore,
    activeKey,
    primary,
    primaryBg,
    borderRadius,
    selectTab,
    closeTab,
    closeByMenu,
    onContext,
    operateMenu,
  }
}
