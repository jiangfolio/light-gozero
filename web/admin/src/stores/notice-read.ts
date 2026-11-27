import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { useUserStore } from './user'

const READ_KEY_PREFIX = 'notice_read_ids:'

/**
 * 通知已读状态：头部铃铛与个人中心共享，落本地按用户隔离。
 */
export const useNoticeReadStore = defineStore('notice-read', () => {
  /** 变更序号：已读写入后递增，驱动角标/列表重算 */
  const version = ref(0)
  /** 公告列表变更序号：发布/删除后递增，驱动头部铃铛重拉 */
  const listTick = ref(0)

  function bumpList() {
    listTick.value += 1
  }

  function storageKey() {
    const user = useUserStore().userInfo
    const uid = user?.id || user?.username || 'anon'
    return `${READ_KEY_PREFIX}${uid}`
  }

  function loadReadIds(): Set<string> {
    try {
      const raw = localStorage.getItem(storageKey())
      const arr = raw ? (JSON.parse(raw) as string[]) : []
      return new Set(Array.isArray(arr) ? arr.map(String) : [])
    } catch {
      return new Set()
    }
  }

  function persist(ids: Set<string>) {
    localStorage.setItem(storageKey(), JSON.stringify([...ids]))
    version.value += 1
  }

  function isRead(id: string) {
    void version.value
    return loadReadIds().has(String(id))
  }

  function markRead(id: string) {
    const key = String(id)
    if (!key) return
    const ids = loadReadIds()
    if (ids.has(key)) return
    ids.add(key)
    persist(ids)
  }

  function markMany(idList: string[]) {
    const ids = loadReadIds()
    let changed = false
    for (const id of idList) {
      const key = String(id)
      if (!key || ids.has(key)) continue
      ids.add(key)
      changed = true
    }
    if (changed) persist(ids)
  }

  const readVersion = computed(() => version.value)

  return {
    readVersion,
    listTick,
    bumpList,
    loadReadIds,
    isRead,
    markRead,
    markMany,
  }
})
