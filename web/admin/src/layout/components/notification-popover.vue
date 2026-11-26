<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{ tone?: 'light' | 'dark' }>()

const open = ref(false)
const isDark = computed(() => props.tone === 'dark')

const messages = [
  { id: 1, title: '系统提示', content: '欢迎使用 light-gozero 管理端', time: '刚刚' },
  { id: 2, title: '安全提醒', content: '建议定期修改登录密码', time: '今天' },
]

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
        <div class="notify-title">消息通知</div>
        <div v-for="m in messages" :key="m.id" class="notify-item">
          <div class="notify-item-title">{{ m.title }}</div>
          <div class="notify-item-content">{{ m.content }}</div>
          <div class="notify-item-time">{{ m.time }}</div>
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
      <ABadge :count="messages.length" :overflow-count="99" class="notify-badge">
        <span class="notify-icon-wrap" :style="iconStyle">
          <IconParkRemind :size="18" :style="iconStyle" />
        </span>
      </ABadge>
    </button>
  </APopover>
</template>

<style scoped lang="less">
.notify-panel {
  width: 280px;
}

.notify-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.notify-item {
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-of-type {
    border-bottom: 0;
  }
}

.notify-item-title {
  font-size: 13px;
  font-weight: 600;
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

.notify-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #aaa;
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
