<script setup lang="ts">
import { message } from 'antdv-next'
import { onMounted, reactive, ref } from 'vue'

import { pageConfig, updateConfig } from '@/api/system/config'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { withEllipsisColumns } from '@/lib/table'
defineOptions({ name: 'SystemConfig' })

const defaultQuery = {
  group_name: '',
  config_key: '',
  status: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)

const columns = withEllipsisColumns([
  { title: '分组', dataIndex: 'group_name', width: 120 },
  { title: '配置键', dataIndex: 'config_key', width: 240 },
  { title: '配置值', dataIndex: 'config_value', width: 180 },
  { title: '类型', dataIndex: 'value_type', width: 90 },
  { title: '备注', dataIndex: 'remark', width: 220 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '更新时间', dataIndex: 'updated_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 100, fixed: 'right' as const },
])

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pageConfig({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      group_name: queryForm.group_name || undefined,
      config_key: queryForm.config_key || undefined,
      status: queryForm.status,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载配置失败')
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

const editOpen = ref(false)
const editSubmitting = ref(false)
const editFormRef = ref()
const editForm = reactive({
  config_key: '',
  config_value: '',
  remark: '',
  value_type: '',
})

function openEdit(record: any) {
  Object.assign(editForm, {
    config_key: record.config_key,
    config_value: record.config_value,
    remark: record.remark || '',
    value_type: record.value_type || '',
  })
  editOpen.value = true
}

async function submitEdit() {
  try {
    await editFormRef.value?.validate()
  } catch {
    return
  }
  editSubmitting.value = true
  try {
    await updateConfig([{ config_key: editForm.config_key, config_value: editForm.config_value }])
    message.success('更新成功')
    editOpen.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '更新失败')
  } finally {
    editSubmitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="分组">
        <AInput
          v-model:value="queryForm.group_name"
          placeholder="配置分组"
          allow-clear
          style="width: 160px"
        />
      </AFormItem>
      <AFormItem label="配置键">
        <AInput
          v-model:value="queryForm.config_key"
          placeholder="请输入配置键"
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
        <template v-if="column.dataIndex === 'config_key'">
          <EllipsisText :text="record.config_key" />
        </template>
        <template v-else-if="column.dataIndex === 'config_value'">
          <EllipsisText :text="record.config_value" />
        </template>
        <template v-else-if="column.dataIndex === 'remark'">
          <EllipsisText :text="record.remark" />
        </template>
        <template v-else-if="column.dataIndex === 'status'">
          <ATag :color="record.status === '1' ? 'green' : 'red'">
            {{ record.status === '1' ? '启用' : '禁用' }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <AButton
            v-privilege="'system:config:update'"
            type="link"
            size="small"
            @click="openEdit(record)"
          >
            编辑
          </AButton>
        </template>
      </template>
    </PageTable>

    <AModal
      v-model:open="editOpen"
      title="编辑配置"
      :confirm-loading="editSubmitting"
      destroy-on-close
      @ok="submitEdit"
    >
      <AForm ref="editFormRef" :model="editForm" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="配置键">
          <AInput v-model:value="editForm.config_key" disabled />
        </AFormItem>
        <AFormItem label="类型">
          <AInput v-model:value="editForm.value_type" disabled />
        </AFormItem>
        <AFormItem
          label="配置值"
          name="config_value"
          :rules="[{ required: true, message: '请输入配置值' }]"
        >
          <AInput v-model:value="editForm.config_value" />
        </AFormItem>
        <AFormItem label="备注">
          <ATextarea v-model:value="editForm.remark" :rows="2" disabled />
        </AFormItem>
      </AForm>
    </AModal>
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
