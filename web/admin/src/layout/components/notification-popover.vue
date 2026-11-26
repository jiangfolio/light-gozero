<script setup lang="ts">
/**
 * 头部消息通知弹层：按类型分 Tab（通知/消息/待办），支持单条已读、全部已读、清空。
 * 当前为演示数据，未对接真实消息接口，接入时替换 demoData 来源并把已读/清空改为调接口即可。
 */
import { Empty, message } from 'antdv-next'
import { computed, ref } from 'vue'

const props = defineProps<{ tone?: 'light' | 'dark' }>()

const open = ref(false)
const isDark = computed(() => props.tone === 'dark')

type NotifyType = 'notice' | 'message' | 'todo'
type NotifyItem = {
  id: number
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

/** 演示数据：按类型分组，待办留空以展示空占位效果 */
const demoData = ref<Record<NotifyType, NotifyItem[]>>({
  notice: [
    { id: 1, title: '系统提示', content: '欢迎使用 light-gozero 管理端', time: '刚刚', read: false },
    { id: 2, title: '安全提醒', content: '建议定期修改登录密码', time: '今天', read: false },
  ],
  message: [
    { id: 3, title: '站内信', content: '演示环境消息通道未开启', time: '昨天', read: false },
  ],
  todo: [],
})

const activeKey = ref<NotifyType>('notice')

const activeList = computed(() => demoData.value[activeKey.value])
const activeLabel = computed(() => tabs.find((t) => t.key === activeKey.value)?.label ?? '')
const activeUnread = computed(() => unreadOf(activeKey.value))

/** 未读数：Tab 标签与铃铛角标统一按未读统计 */
const unreadOf = (key: NotifyType) => demoData.value[key].filter((i) => !i.read).length
const totalUnread = computed(() => tabs.reduce((sum, t) => sum + unreadOf(t.key), 0))

/** 单条标记已读 */
function markRead(item: NotifyItem) {
  item.read = true
}

/** 当前 Tab 全部标记已读 */
function markAllRead() {
  demoData.value[activeKey.value].forEach((i) => {
    i.read = true
  })
}

/** 清空当前 Tab 列表 */
function clearCurrent() {
  demoData.value[activeKey.value] = []
}

/** 演示环境无消息中心页面，先给提示占位 */
function viewAll() {
  message.info('演示环境暂无消息中心页面')
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
          <template v-if="activeList.length">
            <div
              v-for="m in activeList"
              :key="m.id"
              class="notify-item"
              :class="{ 'is-read': m.read }"
              @click="markRead(m)"
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
            :description="`暂无${activeLabel}`"
            class="notify-empty"
          />
        </div>

        <div class="notify-tip">演示数据，未对接真实消息接口</div>
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
</template>

<style scoped lang="less">
.notify-panel {
  width: 320px;
}

.notify-tabs {
  :deep(.ant-tabs-nav) {
    margin-bottom: 0;
  }

  /* Tab 左对齐，右侧留出操作按钮空间 */
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
  font-size: 12px;
  color: #666;
}

.notify-item-time {
  margin-top: 4px;
  font-size: 12px;
  color: #999;
}

.notify-empty {
  margin: 24px 0;
}

.notify-tip {
  padding: 4px 0 0;
  font-size: 12px;
  color: #ccc;
  text-align: center;
}

.notify-trigger {
  display: inline-flex;
  width: 36px;
  height: 36px;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border: 0;
  /* 跟随主题圆角，与头部其他 hover 底统一 */
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
  /* Badge 默认会把子节点文字色打成正文黑，强制跟随 wrap 颜色 */
  :deep(.ant-badge) {
    color: inherit;
  }

  :deep(svg) {
    color: inherit;
    stroke: currentColor !important;
    fill: none !important;
  }
}
</style>
