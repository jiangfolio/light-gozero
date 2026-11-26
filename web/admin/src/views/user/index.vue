<script setup lang="ts">
import type { Dayjs } from 'dayjs'
import { reactive, ref } from 'vue'

import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'

defineOptions({ name: 'SystemUser' })

/** RangePicker 的值类型：与组件声明的二元组保持一致 */
type RangeValue = [start: string | Dayjs | null | undefined, end: string | Dayjs | null | undefined]

const permCreate = 'system:user:create'
const permUpdate = 'system:user:update'

// 1. 查询表单：字段多于一行时检索卡片自动出现展开/收起
const defaultQuery = {
  keywords: '',
  status: undefined as string | undefined,
  orgName: '',
  roleName: '',
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
  { title: '账号', dataIndex: 'username', width: 120 },
  { title: '姓名', dataIndex: 'nickname', width: 120 },
  { title: '手机号', dataIndex: 'phone', width: 140 },
  { title: '所属组织', dataIndex: 'orgName', width: 160 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '创建时间', dataIndex: 'createTime', width: 170 },
  { title: '操作', dataIndex: 'action', width: 130, fixed: 'right' as const },
]

const tableData = ref([
  { id: 1, username: 'superadmin', nickname: '超级管理员', phone: '13800000001', orgName: '总部', status: '1', createTime: '2026-01-01 10:00:00' },
  { id: 2, username: 'zhangsan', nickname: '张三', phone: '13800000002', orgName: '研发部', status: '1', createTime: '2026-02-11 09:30:00' },
  { id: 3, username: 'lisi', nickname: '李四', phone: '13800000003', orgName: '市场部', status: '0', createTime: '2026-03-05 14:20:00' },
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
          placeholder="账号/姓名/手机号"
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
      <AFormItem label="所属组织">
        <AInput v-model:value="queryForm.orgName" placeholder="请输入组织名称" allow-clear style="width: 180px" />
      </AFormItem>
      <AFormItem label="角色">
        <AInput v-model:value="queryForm.roleName" placeholder="请输入角色名称" allow-clear style="width: 180px" />
      </AFormItem>
      <AFormItem label="创建时间">
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
      <template #toolbar>
        <AButton v-privilege="permCreate" type="primary">
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
          <AButton v-privilege="permUpdate" type="link" size="small">编辑</AButton>
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
