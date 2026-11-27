<script setup lang="ts">
/**
 * 个人中心-通知公告：仅展示已发布列表，点开查看详情。
 */
import { message } from 'antdv-next'
import { reactive, ref } from 'vue'

import {
  getPublishedNoticeDetail,
  pagePublishedNotice,
} from '@/api/system/notice'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { withEllipsisColumns } from '@/lib/table'
import { useNoticeReadStore } from '@/stores/notice-read'

const typeOptions = [
  { label: '通知', value: '1' },
  { label: '公告', value: '2' },
]
const typeLabel: Record<string, string> = Object.fromEntries(
  typeOptions.map((o) => [o.value, o.label]),
)

const defaultQuery = {
  title: '',
  type: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)

const columns = withEllipsisColumns([
  { title: '标题', dataIndex: 'title', width: 260 },
  { title: '类型', dataIndex: 'type', width: 100 },
  { title: '发布时间', dataIndex: 'publish_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 100, fixed: 'right' as const },
])

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pagePublishedNotice({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      title: queryForm.title || undefined,
      type: queryForm.type,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载通知公告失败')
  } finally {
    tableLoading.value = false
  }
}

function onSearch() {
  pageNum.value = 1
  loadData()
}

function onReset() {
  Object.assign(queryForm, defaultQuery)
  onSearch()
}

function onPageChange(page: number, size: number) {
  pageNum.value = page
  pageSize.value = size
  loadData()
}

const noticeReadStore = useNoticeReadStore()

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<any>(null)

async function openDetail(record: any) {
  // 1. 先记已读，头部角标与弹层同步
  noticeReadStore.markRead(String(record.id))
  // 2. 再拉详情
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await getPublishedNoticeDetail(record.id)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载详情失败')
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

usePageReload(loadData)
</script>

<template>
  <div class="notice-page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="标题">
        <AInput
          v-model:value="queryForm.title"
          placeholder="请输入标题"
          allow-clear
          style="width: 200px"
        />
      </AFormItem>
      <AFormItem label="类型">
        <ASelect
          v-model:value="queryForm.type"
          placeholder="请选择"
          allow-clear
          style="width: 140px"
          :options="typeOptions"
        />
      </AFormItem>
    </PageSearch>

    <PageTable
      v-model:current="pageNum"
      v-model:page-size="pageSize"
      :total="total"
      :columns="columns"
      :data-source="tableData"
      :loading="tableLoading"
      row-key="id"
      size="small"
      bordered
      :scroll="{ x: 'max-content' }"
      @change="onPageChange"
    >
      <template #toolbarRight>
        <AButton type="text" @click="loadData">
          <template #icon>
            <IconParkRefresh />
          </template>
        </AButton>
      </template>
      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'title'">
          <EllipsisText :text="record.title" />
        </template>
        <template v-else-if="column.dataIndex === 'type'">
          {{ typeLabel[record.type] || record.type }}
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <AButton type="link" size="small" @click="openDetail(record)">查看</AButton>
        </template>
      </template>
    </PageTable>

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
  </div>
</template>

<style scoped>
.notice-page {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  gap: 10px;
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
