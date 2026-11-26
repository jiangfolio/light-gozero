<script setup lang="ts">
import { reactive, ref } from 'vue'

import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'

defineOptions({ name: 'SystemResource' })

// 1. 查询表单
const defaultQuery = {
  keywords: '',
  type: undefined as string | undefined,
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
  { title: '资源名称', dataIndex: 'name', width: 160 },
  { title: '类型', dataIndex: 'type', width: 90 },
  { title: '路由/路径', dataIndex: 'path', width: 180 },
  { title: '权限码', dataIndex: 'perm', width: 180 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '操作', dataIndex: 'action', width: 130, fixed: 'right' as const },
]

const tableData = ref([
  { id: 1, name: '系统管理', type: 'directory', path: '/system', perm: '', status: '1' },
  { id: 2, name: '用户管理', type: 'menu', path: '/system/user', perm: 'system:user:query', status: '1' },
  { id: 3, name: '新建用户', type: 'button', path: '', perm: 'system:user:create', status: '1' },
])
const tableLoading = ref(false)
const total = ref(3)

const typeText: Record<string, string> = {
  directory: '目录',
  menu: '菜单',
  button: '按钮',
}
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="关键字">
        <AInput
          v-model:value="queryForm.keywords"
          placeholder="资源名称/路径/权限码"
          allow-clear
          style="width: 220px"
        />
      </AFormItem>
      <AFormItem label="类型">
        <ASelect
          v-model:value="queryForm.type"
          placeholder="请选择类型"
          allow-clear
          style="width: 140px"
          :options="[
            { label: '目录', value: 'directory' },
            { label: '菜单', value: 'menu' },
            { label: '按钮', value: 'button' },
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
        <template v-if="column.dataIndex === 'type'">
          <ATag :color="record.type === 'menu' ? 'blue' : record.type === 'directory' ? 'purple' : 'default'">
            {{ typeText[record.type] || record.type }}
          </ATag>
        </template>
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
