<script setup lang="ts">
import { reactive, ref } from 'vue'

import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'

defineOptions({ name: 'SystemConfig' })

// 1. 查询表单
const defaultQuery = {
  keywords: '',
  groupName: '',
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
  { title: '配置名称', dataIndex: 'name', width: 160 },
  { title: '配置键', dataIndex: 'key', width: 200 },
  { title: '配置值', dataIndex: 'value', width: 140 },
  { title: '备注', dataIndex: 'remark', width: 200 },
  { title: '更新时间', dataIndex: 'updateTime', width: 170 },
  { title: '操作', dataIndex: 'action', width: 130, fixed: 'right' as const },
]

const tableData = ref([
  { id: 1, name: '密码最小长度', key: 'auth.password_min_len', value: '8', remark: '登录密码最小长度', updateTime: '2026-01-01 10:00:00' },
  { id: 2, name: '密码复杂度', key: 'auth.password_complexity_enabled', value: 'true', remark: '是否开启密码复杂度', updateTime: '2026-01-01 10:00:00' },
  { id: 3, name: '密码过期天数', key: 'auth.password_expire_days', value: '90', remark: '0 表示永不过期', updateTime: '2026-01-01 10:00:00' },
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
          placeholder="配置名称/键"
          allow-clear
          style="width: 220px"
        />
      </AFormItem>
      <AFormItem label="分组">
        <AInput v-model:value="queryForm.groupName" placeholder="请输入分组" allow-clear style="width: 160px" />
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

      <template #bodyCell="{ column }">
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
