<script setup lang="ts">
/**
 * 通知公告管理：标准分页 CRUD。
 */
import { message, Modal } from 'antdv-next'
import { computed, reactive, ref } from 'vue'

import {
  createNotice,
  deleteNotice,
  pageNotice,
  updateNotice,
} from '@/api/system/notice'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { withEllipsisColumns } from '@/lib/table'
import { useNoticeReadStore } from '@/stores/notice-read'

defineOptions({ name: 'SystemNotice' })

const noticeReadStore = useNoticeReadStore()

const typeOptions = [
  { label: '通知', value: '1' },
  { label: '公告', value: '2' },
]
const typeLabel: Record<string, string> = Object.fromEntries(
  typeOptions.map((o) => [o.value, o.label]),
)
const statusOptions = [
  { label: '草稿', value: '0' },
  { label: '已发布', value: '1' },
]

const defaultQuery = {
  title: '',
  type: undefined as string | undefined,
  status: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)
const selectedKeys = ref<string[]>([])

const columns = withEllipsisColumns([
  { title: '标题', dataIndex: 'title', width: 220 },
  { title: '类型', dataIndex: 'type', width: 100 },
  { title: '状态', dataIndex: 'status', width: 100 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '发布时间', dataIndex: 'publish_at', width: 180 },
  { title: '更新时间', dataIndex: 'updated_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 140, fixed: 'right' as const },
])

const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedKeys.value = keys.map(String)
  },
}))

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pageNotice({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      title: queryForm.title || undefined,
      type: queryForm.type,
      status: queryForm.status,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
    selectedKeys.value = []
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

const modalOpen = ref(false)
const modalSubmitting = ref(false)
const editingId = ref<string | null>(null)
const modalTitle = computed(() => (editingId.value ? '编辑通知公告' : '新建通知公告'))
const formRef = ref()
const formState = reactive({
  title: '',
  content: '',
  type: '1',
  status: '0',
  sort: '0',
  remark: '',
})

function openCreate() {
  editingId.value = null
  Object.assign(formState, {
    title: '',
    content: '',
    type: '1',
    status: '0',
    sort: '0',
    remark: '',
  })
  modalOpen.value = true
}

function openEdit(record: any) {
  editingId.value = record.id
  Object.assign(formState, {
    title: record.title || '',
    content: record.content || '',
    type: record.type || '1',
    status: record.status || '0',
    sort: record.sort || '0',
    remark: record.remark || '',
  })
  modalOpen.value = true
}

async function submitForm() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  modalSubmitting.value = true
  try {
    const payload = {
      title: formState.title,
      content: formState.content,
      type: formState.type,
      status: formState.status,
      sort: formState.sort,
      remark: formState.remark || undefined,
    }
    if (editingId.value) {
      await updateNotice({ id: editingId.value, ...payload })
      message.success('更新成功')
    } else {
      await createNotice(payload)
      message.success('创建成功')
    }
    modalOpen.value = false
    await loadData()
    noticeReadStore.bumpList()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    modalSubmitting.value = false
  }
}

function confirmDelete(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择通知公告')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 条通知公告？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteNotice(ids)
        message.success('删除成功')
        await loadData()
        noticeReadStore.bumpList()
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '删除失败')
        throw e
      }
    },
  })
}

usePageReload(loadData)
</script>

<template>
  <div class="page">
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
      <AFormItem label="状态">
        <ASelect
          v-model:value="queryForm.status"
          placeholder="请选择"
          allow-clear
          style="width: 140px"
          :options="statusOptions"
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
      :row-selection="rowSelection"
      row-key="id"
      size="small"
      bordered
      :scroll="{ x: 'max-content' }"
      @change="onPageChange"
    >
      <template #toolbar>
        <AButton v-privilege="'system:notice:create'" type="primary" @click="openCreate">
          新建
        </AButton>
        <AButton
          v-privilege="'system:notice:delete'"
          danger
          :disabled="!selectedKeys.length"
          @click="confirmDelete(selectedKeys)"
        >
          批量删除
        </AButton>
      </template>
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
        <template v-else-if="column.dataIndex === 'status'">
          <ATag :color="record.status === '1' ? 'green' : 'default'">
            {{ record.status === '1' ? '已发布' : '草稿' }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <AButton
            v-privilege="'system:notice:update'"
            type="link"
            size="small"
            @click="openEdit(record)"
          >
            编辑
          </AButton>
          <AButton
            v-privilege="'system:notice:delete'"
            type="link"
            size="small"
            danger
            @click="confirmDelete([record.id])"
          >
            删除
          </AButton>
        </template>
      </template>
    </PageTable>

    <AModal
      v-model:open="modalOpen"
      :title="modalTitle"
      :confirm-loading="modalSubmitting"
      destroy-on-close
      width="640px"
      @ok="submitForm"
    >
      <AForm ref="formRef" :model="formState" layout="vertical">
        <AFormItem
          label="标题"
          name="title"
          :rules="[{ required: true, message: '请输入标题' }]"
        >
          <AInput
            v-model:value="formState.title"
            placeholder="请输入标题"
            :maxlength="200"
            show-count
            allow-clear
          />
        </AFormItem>
        <AFormItem
          label="正文"
          name="content"
          :rules="[{ required: true, message: '请输入正文' }]"
        >
          <ATextarea
            v-model:value="formState.content"
            :rows="6"
            placeholder="请输入正文"
            :maxlength="20000"
            show-count
            allow-clear
          />
        </AFormItem>
        <AFormItem
          label="类型"
          name="type"
          :rules="[{ required: true, message: '请选择类型' }]"
        >
          <ASelect v-model:value="formState.type" :options="typeOptions" />
        </AFormItem>
        <AFormItem label="状态" name="status">
          <ASelect v-model:value="formState.status" :options="statusOptions" />
        </AFormItem>
        <AFormItem label="排序" name="sort">
          <AInput v-model:value="formState.sort" placeholder="数字越小越靠前" />
        </AFormItem>
        <AFormItem label="备注" name="remark">
          <AInput
            v-model:value="formState.remark"
            placeholder="选填"
            :maxlength="255"
            show-count
            allow-clear
          />
        </AFormItem>
      </AForm>
    </AModal>
  </div>
</template>

<style scoped>
.page {
  display: flex;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  gap: 10px;
}
</style>
