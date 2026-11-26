<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue'

import { tableScrollX, withEllipsisColumns } from '@/lib/table'

/**
 * 内容卡片：工具栏（左操作/右工具）+ 表体滚动区 + 独立分页脚部。
 * 卡片高度由页面 flex 布局撑满剩余空间，不随内容变化；只有表体滚动。
 * 表格属性经 attrs 透传；列默认开启 ellipsis，scroll.x 按列宽求和。
 */
defineOptions({ inheritAttrs: false })

const props = defineProps<{
  /** 总条数；传入即显示分页脚部，不传则不渲染分页器 */
  total?: number
  /** 当前页码，配合 v-model:current 使用 */
  current?: number
  /** 每页条数，配合 v-model:page-size 使用 */
  pageSize?: number
}>()

const emit = defineEmits<{
  'update:current': [value: number]
  'update:pageSize': [value: number]
  /** 页码或每页条数变化（含切换每页条数），页面在此重新拉取列表 */
  change: [current: number, pageSize: number]
}>()

const attrs = useAttrs()
const bodyRef = ref<HTMLElement>()

/* 表头吸顶到卡片自己的滚动容器，而非 window */
const stickyConfig = computed(() => ({
  getContainer: () => bodyRef.value as HTMLElement,
}))

/** 文本列默认 ellipsis；操作列 / 显式配置保持原样 */
const tableColumns = computed(() => withEllipsisColumns(((attrs.columns as any[]) ?? []) as any[]))

/**
 * max-content 会按内容撑开列宽导致省略失效；
 * 改为列宽求和，横向滚动与单行省略同时生效。
 */
const tableScroll = computed(() => {
  const scroll = (attrs.scroll as Record<string, unknown> | undefined) ?? {}
  if (scroll.x != null && scroll.x !== 'max-content') return scroll
  return { ...scroll, x: tableScrollX(tableColumns.value) }
})

/** 去掉已单独处理的 columns/scroll，避免覆盖 */
const tableBind = computed(() => {
  const rest: Record<string, unknown> = { ...attrs }
  delete rest.columns
  delete rest.scroll
  return rest
})

const currentPage = computed({
  get: () => props.current ?? 1,
  set: (value: number) => emit('update:current', value),
})
const currentPageSize = computed({
  get: () => props.pageSize ?? 10,
  set: (value: number) => emit('update:pageSize', value),
})

function onPageChange(page: number, size: number) {
  emit('change', page, size)
}

const showTotal = (t: number) => `共 ${t} 条`
</script>

<template>
  <ACard size="small" :bordered="false" class="page-table">
    <div v-if="$slots.toolbar || $slots.toolbarRight" class="page-table-toolbar">
      <div class="page-table-toolbar-left">
        <slot name="toolbar" />
      </div>
      <div class="page-table-toolbar-right">
        <slot name="toolbarRight" />
      </div>
    </div>

    <div ref="bodyRef" class="page-table-body">
      <ATable
        v-bind="tableBind"
        :columns="tableColumns"
        :scroll="tableScroll"
        :sticky="stickyConfig"
        :pagination="false"
        table-layout="fixed"
      >
        <template v-for="(_, name) in $slots" :key="name" #[name]="slotProps">
          <slot :name="name" v-bind="slotProps ?? {}" />
        </template>
      </ATable>
    </div>

    <div v-if="total !== undefined" class="page-table-footer">
      <APagination
        v-model:current="currentPage"
        v-model:page-size="currentPageSize"
        :total="total"
        size="small"
        show-size-changer
        :show-total="showTotal"
        @change="onPageChange"
        @show-size-change="onPageChange"
      />
    </div>
  </ACard>
</template>

<style scoped lang="less">
.page-table {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;

  :deep(.ant-card-body) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 12px;
  }

  /* 兜底省略；自定义 bodyCell 请包 EllipsisText */
  :deep(.ant-table-cell-ellipsis),
  :deep(.ant-table-cell-ellipsis .ant-table-cell-content) {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    word-break: keep-all;
  }
}

.page-table-toolbar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  .page-table-toolbar-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-table-toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.page-table-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.page-table-footer {
  display: flex;
  flex-shrink: 0;
  justify-content: flex-end;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--ant-color-split, #f0f0f0);
}
</style>
