<script setup lang="ts">
import { message } from 'antdv-next'
import { reactive, ref } from 'vue'

import { pageAuditOper } from '@/api/system/audit'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { withEllipsisColumns } from '@/lib/table'

defineOptions({ name: 'SystemAuditOper' })

const defaultQuery = {
  username: '',
  module: '',
  result: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)

const columns = withEllipsisColumns([
  { title: '用户名', dataIndex: 'username', width: 120 },
  { title: '模块', dataIndex: 'module', width: 110 },
  { title: '动作', dataIndex: 'action', width: 120 },
  { title: '说明', dataIndex: 'message', width: 200 },
  { title: '方法', dataIndex: 'method', width: 80 },
  { title: '路径', dataIndex: 'path', width: 220 },
  { title: 'IP', dataIndex: 'ip', width: 130 },
  { title: '耗时(ms)', dataIndex: 'latency_ms', width: 100 },
  { title: '结果', dataIndex: 'result', width: 90 },
  { title: '时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', dataIndex: 'action_btn', width: 90, fixed: 'right' as const },
])

const resultOptions = [
  { label: '成功', value: 'success' },
  { label: '失败', value: 'fail' },
]

const detailOpen = ref(false)
const detailRecord = ref<any>(null)

function openDetail(record: any) {
  detailRecord.value = record
  detailOpen.value = true
}

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pageAuditOper({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      username: queryForm.username || undefined,
      module: queryForm.module || undefined,
      result: queryForm.result,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载操作日志失败')
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

usePageReload(loadData)
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="用户名">
        <AInput
          v-model:value="queryForm.username"
          placeholder="请输入用户名"
          allow-clear
          style="width: 160px"
        />
      </AFormItem>
      <AFormItem label="模块">
        <AInput
          v-model:value="queryForm.module"
          placeholder="请输入模块"
          allow-clear
          style="width: 160px"
        />
      </AFormItem>
      <AFormItem label="结果">
        <ASelect
          v-model:value="queryForm.result"
          placeholder="请选择"
          allow-clear
          style="width: 140px"
          :options="resultOptions"
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
        <template v-if="column.dataIndex === 'result'">
          <ATag :color="record.result === 'success' ? 'green' : 'red'">
            {{ record.result_name || record.result || '-' }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'message'">
          <EllipsisText :text="record.message" />
        </template>
        <template v-else-if="column.dataIndex === 'path'">
          <EllipsisText :text="record.path" />
        </template>
        <template v-else-if="column.dataIndex === 'action_btn'">
          <AButton type="link" size="small" @click="openDetail(record)">详情</AButton>
        </template>
      </template>
    </PageTable>

    <ADrawer
      v-model:open="detailOpen"
      title="操作日志详情"
      :size="640"
      destroy-on-close
    >
      <ADescriptions v-if="detailRecord" :column="1" bordered size="small">
        <ADescriptionsItem label="用户名">{{ detailRecord.username }}</ADescriptionsItem>
        <ADescriptionsItem label="模块">{{ detailRecord.module }}</ADescriptionsItem>
        <ADescriptionsItem label="动作">{{ detailRecord.action }}</ADescriptionsItem>
        <ADescriptionsItem label="说明">{{ detailRecord.message || '-' }}</ADescriptionsItem>
        <ADescriptionsItem label="方法">{{ detailRecord.method }}</ADescriptionsItem>
        <ADescriptionsItem label="路径">
          <EllipsisText :text="detailRecord.path" />
        </ADescriptionsItem>
        <ADescriptionsItem label="IP">{{ detailRecord.ip }}</ADescriptionsItem>
        <ADescriptionsItem label="耗时">{{ detailRecord.latency_ms }} ms</ADescriptionsItem>
        <ADescriptionsItem label="结果">
          <ATag :color="detailRecord.result === 'success' ? 'green' : 'red'">
            {{ detailRecord.result_name || detailRecord.result || '-' }}
          </ATag>
        </ADescriptionsItem>
        <ADescriptionsItem label="时间">{{ detailRecord.created_at || '-' }}</ADescriptionsItem>
        <ADescriptionsItem label="请求体">
          <pre class="req-body">{{ detailRecord.req_body || '-' }}</pre>
        </ADescriptionsItem>
      </ADescriptions>
    </ADrawer>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 10px;
  padding-bottom: 10px;
}

.req-body {
  margin: 0;
  max-height: 280px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
}
</style>
