<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const emit = defineEmits<{ search: []; reset: [] }>()

/**
 * 检索卡片：一行放不下全部条件时自动出现"展开/收起"。
 * 条件区超过一行即视为可折叠，收起态只保留首行。
 */
const bodyRef = ref<HTMLElement>()
const collapsible = ref(false)
const expanded = ref(false)

/* 收起态保留的高度：表单项 32px + 少量余量 */
const COLLAPSED_HEIGHT = 40

let observer: ResizeObserver | null = null

function measure() {
  const el = bodyRef.value
  if (!el) return
  // scrollHeight 不受 max-height 裁剪影响，量到的始终是内容全高
  collapsible.value = el.scrollHeight > COLLAPSED_HEIGHT
  if (!collapsible.value) expanded.value = false
}

onMounted(() => {
  measure()
  if (bodyRef.value) {
    observer = new ResizeObserver(measure)
    observer.observe(bodyRef.value)
  }
})

onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <ACard size="small" :bordered="false" class="page-search">
    <div class="page-search-main">
      <div
        ref="bodyRef"
        class="page-search-body"
        :class="{ collapsed: collapsible && !expanded }"
      >
        <AForm layout="inline">
          <slot />
        </AForm>
      </div>
      <div class="page-search-actions">
        <AButton type="primary" @click="emit('search')">
          <template #icon>
            <IconParkSearch />
          </template>
          查询
        </AButton>
        <AButton @click="emit('reset')">
          <template #icon>
            <IconParkRefresh />
          </template>
          重置
        </AButton>
        <a v-if="collapsible" class="page-search-toggle" @click="expanded = !expanded">
          {{ expanded ? '收起' : '展开' }}
          <IconParkUp v-if="expanded" />
          <IconParkDown v-else />
        </a>
      </div>
    </div>
  </ACard>
</template>

<style scoped lang="less">
.page-search {
  flex-shrink: 0;

  :deep(.ant-card-body) {
    padding: 12px 12px 4px;
  }
}

.page-search-main {
  display: flex;
  align-items: flex-start;
}

.page-search-body {
  flex: 1;
  min-width: 0;
  overflow: hidden;

  &.collapsed {
    max-height: 40px;
  }

  :deep(.ant-form-item) {
    margin-right: 16px;
    margin-bottom: 8px;
  }
}

.page-search-actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 8px;
  margin-left: 12px;

  .page-search-toggle {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    white-space: nowrap;
    cursor: pointer;
    user-select: none;
  }
}
</style>
