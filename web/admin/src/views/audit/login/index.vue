<script setup lang="ts">
import { message } from 'antdv-next'
import { reactive, ref } from 'vue'

import { pageAuditLogin } from '@/api/system/audit'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { withEllipsisColumns } from '@/lib/table'

defineOptions({ name: 'SystemAuditLogin' })

const defaultQuery = {
  username: '',
  result: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)

const columns = withEllipsisColumns([
  { title: '用户名', dataIndex: 'username', width: 140 },
  { title: 'IP', dataIndex: 'ip', width: 140 },
  { title: '结果', dataIndex: 'result', width: 100 },
  { title: '说明', dataIndex: 'message', width: 220 },
  { title: 'User-Agent', dataIndex: 'user_agent', width: 280 },
  { title: '时间', dataIndex: 'created_at', width: 180 },
])

const resultOptions = [
  { label: '成功', value: 'success' },
  { label: '失败', value: 'fail' },
  { label: '锁定', value: 'locked' },
  { label: '登出', value: 'logout' },
]

const resultColor: Record<string, string> = {
  success: 'green',
  fail: 'red',
  locked: 'orange',
  logout: 'blue',
}

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pageAuditLogin({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      username: queryForm.username || undefined,
      result: queryForm.result,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载登录日志失败')
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
          style="width: 180px"
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
          <ATag :color="resultColor[record.result] || 'default'">
            {{ record.result_name || record.result || '-' }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'message'">
          <EllipsisText :text="record.message" />
        </template>
        <template v-else-if="column.dataIndex === 'user_agent'">
          <EllipsisText :text="record.user_agent" />
        </template>
      </template>
    </PageTable>
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
</style>
