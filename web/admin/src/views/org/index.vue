<script setup lang="ts">
import { reactive, ref } from 'vue'

import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'

defineOptions({ name: 'SystemOrg' })

// 1. 查询表单
const defaultQuery = {
  keywords: '',
  status: undefined as string | undefined,
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
  { title: '组织名称', dataIndex: 'name', width: 180 },
  { title: '组织编码', dataIndex: 'code', width: 140 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '创建时间', dataIndex: 'createTime', width: 170 },
  { title: '操作', dataIndex: 'action', width: 130, fixed: 'right' as const },
]

const tableData = ref([
  { id: 1, name: '总部', code: 'HQ', sort: 1, status: '1', createTime: '2026-01-01 10:00:00' },
  { id: 2, name: '研发部', code: 'RD', sort: 2, status: '1', createTime: '2026-01-02 11:00:00' },
  { id: 3, name: '市场部', code: 'MKT', sort: 3, status: '1', createTime: '2026-01-03 15:00:00' },
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
          placeholder="组织名称/编码"
          allow-clear
          style="width: 220px"
        />
      </AFormItem>
      <AFormItem label="状态">
        <ASelect
          v-model:value="queryForm.status"
          placeholder="请选择状态"
          allow-clear
          style="width: 140px"
          :options="[
            { label: '启用', value: '1' },
            { label: '禁用', value: '0' },
          ]"
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
    >
      <template #toolbar>
        <AButton type="primary">
          <template #icon>
            <IconParkPlus />
          </template>
          新建
        </AButton>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'status'">
          <ATag :color="record.status === '1' ? 'green' : 'red'">
            {{ record.status === '1' ? '启用' : '禁用' }}
          </ATag>
        </template>
        <template v-if="column.dataIndex === 'action'">
          <AButton type="link" size="small">编辑</AButton>
          <AButton type="link" size="small" danger>删除</AButton>
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
