<script setup lang="ts">
import type { Dayjs } from 'dayjs'
import { reactive, ref } from 'vue'

import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'

defineOptions({ name: 'SystemAudit' })

/** RangePicker 的值类型：与组件声明的二元组保持一致 */
type RangeValue = [start: string | Dayjs | null | undefined, end: string | Dayjs | null | undefined]

// 1. 查询表单
const defaultQuery = {
  keywords: '',
  actionType: '',
  dateRange: undefined as RangeValue | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)

function onSearch() {
  // 1. 查询回第一页，避免停留在空页
  pageNum.value = 1
  // 2. TODO: 对接接口后按 queryForm 拉取列表
}

function onReset() {
  Object.assign(queryForm, defaultQuery)
  onSearch()
}

// 2. 表格列与演示数据（接口接入后替换为真实分页数据）
const columns = [
  { title: '操作人', dataIndex: 'operator', width: 120 },
  { title: '模块', dataIndex: 'module', width: 120 },
  { title: '动作', dataIndex: 'actionType', width: 100 },
  { title: '结果', dataIndex: 'result', width: 90 },
  { title: 'IP', dataIndex: 'ip', width: 140 },
  { title: '操作时间', dataIndex: 'operateTime', width: 170 },
]

const tableData = ref([
  { id: 1, operator: 'superadmin', module: '用户管理', actionType: '新建', result: '1', ip: '127.0.0.1', operateTime: '2026-03-01 10:00:00' },
  { id: 2, operator: 'superadmin', module: '角色管理', actionType: '编辑', result: '1', ip: '127.0.0.1', operateTime: '2026-03-01 11:00:00' },
  { id: 3, operator: 'zhangsan', module: '登录', actionType: '登录', result: '0', ip: '192.168.1.10', operateTime: '2026-03-02 09:00:00' },
])
const tableLoading = ref(false)
const total = ref(3)
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="关键字">
        <AInput
          v-model:value="queryForm.keywords"
          placeholder="操作人/模块"
          allow-clear
          style="width: 220px"
        />
      </AFormItem>
      <AFormItem label="动作">
        <AInput v-model:value="queryForm.actionType" placeholder="请输入动作" allow-clear style="width: 140px" />
      </AFormItem>
      <AFormItem label="操作时间">
        <ARangePicker v-model:value="queryForm.dateRange" />
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
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'result'">
          <ATag :color="record.result === '1' ? 'green' : 'red'">
            {{ record.result === '1' ? '成功' : '失败' }}
          </ATag>
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
