<script setup lang="ts">
/**
 * 头部通知弹层：通知 Tab 拉已发布公告；已读状态与个人中心共享。
 * 消息 / 待办尚未接入后端，保持空列表。
 */
import { Empty, message } from 'antdv-next'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import {
  getPublishedNoticeDetail,
  pagePublishedNotice,
} from '@/api/system/notice'
import { useNoticeReadStore } from '@/stores/notice-read'

const props = defineProps<{ tone?: 'light' | 'dark' }>()

const router = useRouter()
const noticeReadStore = useNoticeReadStore()

const open = ref(false)
const loading = ref(false)
const isDark = computed(() => props.tone === 'dark')

type NotifyType = 'notice' | 'message' | 'todo'
type NotifyItem = {
  id: string
  title: string
  content: string
  time: string
  read: boolean
}

const tabs: { key: NotifyType; label: string }[] = [
  { key: 'notice', label: '通知' },
  { key: 'message', label: '消息' },
  { key: 'todo', label: '待办' },
]

const lists = ref<Record<NotifyType, NotifyItem[]>>({
  notice: [],
  message: [],
  todo: [],
})

const activeKey = ref<NotifyType>('notice')
const activeList = computed(() => lists.value[activeKey.value])
const activeLabel = computed(() => tabs.find((t) => t.key === activeKey.value)?.label ?? '')
const activeUnread = computed(() => unreadOf(activeKey.value))

function unreadOf(key: NotifyType) {
  return lists.value[key].filter((i) => !i.read).length
}

const totalUnread = computed(() => tabs.reduce((sum, t) => sum + unreadOf(t.key), 0))

/** 拉取最近已发布公告，并用共享已读集合打标 */
async function loadNotices() {
  loading.value = true
  try {
    const resp = await pagePublishedNotice({ page: '1', page_size: '20' })
    lists.value.notice = (resp.items ?? []).map((row: any) => ({
      id: String(row.id),
      title: row.title || '',
      content: row.content || '',
      time: row.publish_at || row.created_at || '',
      read: noticeReadStore.isRead(String(row.id)),
    }))
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载通知失败')
  } finally {
    loading.value = false
  }
}

/** 已读状态在别处变更时，同步列表打标（不强制重拉接口） */
watch(
  () => noticeReadStore.readVersion,
  () => {
    lists.value.notice = lists.value.notice.map((item) => ({
      ...item,
      read: noticeReadStore.isRead(item.id),
    }))
  },
)

/** 管理端发布/删除公告后重拉，刷新角标 */
watch(
  () => noticeReadStore.listTick,
  () => {
    void loadNotices()
  },
)

watch(open, (v) => {
  if (v) loadNotices()
})

onMounted(() => {
  // 进页先拉一次，铃铛角标才有未读数
  loadNotices()
})

/** 单条已读并写入共享 store */
function markRead(item: NotifyItem) {
  if (item.read) return
  item.read = true
  noticeReadStore.markRead(item.id)
}

/** 当前 Tab 全部已读（仅通知有真实数据） */
function markAllRead() {
  const ids = lists.value[activeKey.value].map((i) => i.id)
  lists.value[activeKey.value].forEach((i) => {
    i.read = true
  })
  if (activeKey.value === 'notice') noticeReadStore.markMany(ids)
}

/** 清空当前 Tab 展示列表（不删服务端数据） */
function clearCurrent() {
  lists.value[activeKey.value] = []
}

/** 跳转个人中心通知公告面板 */
function viewAll() {
  open.value = false
  router.push({ path: '/profile', query: { tab: 'notice' } })
}

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<any>(null)

const typeLabel: Record<string, string> = { '1': '通知', '2': '公告' }

/** 点开条目：标记已读并拉详情 */
async function onItemClick(item: NotifyItem) {
  markRead(item)
  if (activeKey.value !== 'notice') return
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await getPublishedNoticeDetail(item.id)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载详情失败')
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

/** 暗色顶栏强制白色，绕过 Badge 默认文字色 */
const iconStyle = computed(() =>
  isDark.value
    ? { color: 'rgba(255, 255, 255, 0.85)', stroke: 'rgba(255, 255, 255, 0.85)' }
    : { color: 'rgba(0, 0, 0, 0.65)', stroke: 'rgba(0, 0, 0, 0.65)' },
)
</script>

<template>
  <APopover v-model:open="open" placement="bottomRight" trigger="click">
    <template #content>
      <div class="notify-panel">
        <ATabs v-model:active-key="activeKey" size="small" class="notify-tabs">
          <template #rightExtra>
            <div class="notify-actions">
              <ATooltip title="全部已读" :mouse-enter-delay="0.3">
                <button
                  type="button"
                  class="notify-action"
                  :disabled="!activeUnread"
                  @click="markAllRead"
                >
                  <IconParkDoneAll :size="15" />
                </button>
              </ATooltip>
              <ATooltip title="清空" :mouse-enter-delay="0.3">
                <button
                  type="button"
                  class="notify-action"
                  :disabled="!activeList.length"
                  @click="clearCurrent"
                >
                  <IconParkClear :size="15" />
                </button>
              </ATooltip>
              <ATooltip title="查看全部" :mouse-enter-delay="0.3">
                <button type="button" class="notify-action" @click="viewAll">
                  <IconParkExpandRight :size="15" />
                </button>
              </ATooltip>
            </div>
          </template>
          <ATabPane v-for="t in tabs" :key="t.key">
            <template #tab>
              <span class="notify-tab-label">
                {{ t.label }}
                <span v-if="unreadOf(t.key)" class="notify-tab-count">({{ unreadOf(t.key) }})</span>
              </span>
            </template>
          </ATabPane>
        </ATabs>

        <div class="notify-list">
          <div v-if="loading && activeKey === 'notice'" class="notify-loading">
            <ASpin size="small" />
          </div>
          <template v-else-if="activeList.length">
            <div
              v-for="m in activeList"
              :key="m.id"
              class="notify-item"
              :class="{ 'is-read': m.read }"
              @click="onItemClick(m)"
            >
              <div class="notify-item-title">
                <span v-if="!m.read" class="notify-dot" />
                {{ m.title }}
              </div>
              <div class="notify-item-content">{{ m.content }}</div>
              <div class="notify-item-time">{{ m.time }}</div>
            </div>
          </template>
          <AEmpty
            v-else
            :image="Empty.PRESENTED_IMAGE_SIMPLE"
            :description="
              activeKey === 'notice' ? `暂无${activeLabel}` : `${activeLabel}暂未接入`
            "
            class="notify-empty"
          />
        </div>
      </div>
    </template>
    <button
      type="button"
      title="通知"
      class="notify-trigger"
      :class="isDark ? 'is-dark' : 'is-light'"
    >
      <ABadge :count="totalUnread" :overflow-count="99" class="notify-badge">
        <span class="notify-icon-wrap" :style="iconStyle">
          <IconParkRemind :size="18" :style="iconStyle" />
        </span>
      </ABadge>
    </button>
  </APopover>

  <AModal
    v-model:open="detailOpen"
    title="通知公告详情"
    :footer="null"
    destroy-on-close
    width="640px"
  >
    <ASpin :spinning="detailLoading">
      <template v-if="detail">
        <div class="detail-title">{{ detail.title }}</div>
        <div class="detail-meta">
          <span>{{ typeLabel[detail.type] || detail.type }}</span>
          <span v-if="detail.publish_at">{{ detail.publish_at }}</span>
        </div>
        <div class="detail-content">{{ detail.content }}</div>
      </template>
    </ASpin>
  </AModal>
</template>

<style scoped lang="less">
.notify-panel {
  width: 320px;
}

.notify-tabs {
  :deep(.ant-tabs-nav) {
    margin-bottom: 0;
  }

  :deep(.ant-tabs-extra-content) {
    line-height: 1;
  }
}

.notify-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.notify-action {
  display: inline-flex;
  width: 24px;
  height: 24px;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #666;
  padding: 0;

  &:hover:not(:disabled) {
    background: rgba(0, 0, 0, 0.05);
    color: var(--ant-color-primary, #1677ff);
  }

  &:disabled {
    cursor: not-allowed;
    color: #ccc;
  }
}

.notify-tab-label {
  font-size: 13px;
}

.notify-tab-count {
  color: #999;
}

.notify-list {
  max-height: 260px;
  overflow-x: hidden;
  overflow-y: auto;
}

.notify-loading {
  display: flex;
  justify-content: center;
  padding: 32px 0;
}

.notify-item {
  padding: 8px 4px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;

  &:last-child {
    border-bottom: 0;
  }

  &.is-read {
    .notify-item-title,
    .notify-item-content {
      color: #999;
    }
  }
}

.notify-item-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
}

.notify-dot {
  width: 6px;
  height: 6px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--ant-color-primary, #1677ff);
}

.notify-item-content {
  margin-top: 2px;
  overflow: hidden;
  font-size: 12px;
  color: #666;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notify-item-time {
  margin-top: 4px;
  font-size: 12px;
  color: #999;
}

.notify-empty {
  margin: 24px 0;
}

.notify-trigger {
  display: inline-flex;
  width: 36px;
  height: 36px;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: var(--app-border-radius, 0px);
  background: transparent;
  padding: 0;

  &.is-dark:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  &.is-light:hover {
    background: rgba(0, 0, 0, 0.05);
  }
}

.notify-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.notify-badge {
  :deep(.ant-badge) {
    color: inherit;
  }

  :deep(svg) {
    color: inherit;
    stroke: currentColor !important;
    fill: none !important;
  }
}

.detail-title {
  margin-bottom: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1e1e1e;
}

.detail-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #64748b;
}

.detail-content {
  white-space: pre-wrap;
  line-height: 1.7;
  color: #334155;
}
</style>
