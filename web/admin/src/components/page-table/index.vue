<script setup lang="ts">
import { computed, ref } from 'vue'

/**
 * 内容卡片：工具栏（左操作/右工具）+ 表体滚动区 + 独立分页脚部。
 * 卡片高度由页面 flex 布局撑满剩余空间，不随内容变化；只有表体滚动。
 * 表格属性（columns/dataSource/loading/rowKey/scroll 等）经 $attrs 透传给内部 ATable。
 * 分页器与表格完全解耦：由本组件以独立脚部渲染，不使用 ATable 自带分页。
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

const bodyRef = ref<HTMLElement>()

/* 表头吸顶到卡片自己的滚动容器，而非 window */
const stickyConfig = computed(() => ({
  getContainer: () => bodyRef.value as HTMLElement,
}))

/* 页码/每页条数双向绑定的中转：未传时使用默认值，保证分页器可控 */
const currentPage = computed({
  get: () => props.current ?? 1,
  set: (value: number) => emit('update:current', value),
})
const currentPageSize = computed({
  get: () => props.pageSize ?? 10,
  set: (value: number) => emit('update:pageSize', value),
})

/** 页码变化与每页条数变化都统一抛 change，页面只需监听一个事件 */
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
      <ATable v-bind="$attrs" :sticky="stickyConfig" :pagination="false">
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

/* 唯一滚动区：表格内容；表头经 sticky 吸附在此容器顶部 */
.page-table-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

/* 分页器独立于表格：固定在卡片底部，以分隔线与表体划开 */
.page-table-footer {
  display: flex;
  flex-shrink: 0;
  justify-content: flex-end;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--ant-color-split, #f0f0f0);
}
</style>
