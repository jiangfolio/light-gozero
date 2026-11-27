import { onActivated, onMounted, ref } from 'vue'

/**
 * 列表/看板在 keep-alive 下：首次挂载拉数，回切页签再拉，避免缓存脏数据。
 */
export function usePageReload(reload: () => void | Promise<void>) {
  const primed = ref(false)

  onMounted(async () => {
    await reload()
    primed.value = true
  })

  onActivated(() => {
    // 首次挂载时 onActivated 也会触发，跳过以免连打两次
    if (!primed.value) return
    void reload()
  })
}
