<script setup lang="ts">
import { message, Modal } from 'antdv-next'
import { computed, reactive, ref } from 'vue'

import {
  createDictItem,
  createDictType,
  deleteDictItem,
  deleteDictType,
  pageDictItem,
  pageDictType,
  updateDictItem,
  updateDictType,
} from '@/api/system/dict'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { tableScrollX, withEllipsisColumns } from '@/lib/table'

defineOptions({ name: 'SystemDict' })

const defaultQuery = {
  keywords: '',
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
  { title: '编码', dataIndex: 'dict_code', width: 180 },
  { title: '名称', dataIndex: 'label', width: 160 },
  { title: '备注', dataIndex: 'remark', width: 200 },
  { title: '状态', dataIndex: 'status', width: 110 },
  { title: '更新时间', dataIndex: 'updated_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 160, fixed: 'right' as const },
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
    // 1. 关键字优先当 dict_code，否则当 label（后端无统一 keywords）
    const kw = queryForm.keywords.trim()
    const resp = await pageDictType({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      dict_code: kw.includes('_') || /^[A-Za-z0-9_]+$/.test(kw) ? kw || undefined : undefined,
      label: kw && !/^[A-Za-z0-9_]+$/.test(kw) ? kw : undefined,
      status: queryForm.status,
    })
    // 若关键字像编码但无结果，再按 label 兜底一次（同页）
    if (kw && (resp.items?.length ?? 0) === 0 && /^[A-Za-z0-9_]+$/.test(kw)) {
      const byLabel = await pageDictType({
        page: String(pageNum.value),
        page_size: String(pageSize.value),
        label: kw,
        status: queryForm.status,
      })
      tableData.value = byLabel.items ?? []
      total.value = Number(byLabel.total || 0)
    } else {
      tableData.value = resp.items ?? []
      total.value = Number(resp.total || 0)
    }
    selectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载字典失败')
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

// —— 类型新建/编辑 ——
const typeModalOpen = ref(false)
const typeSubmitting = ref(false)
const editingTypeId = ref<string | null>(null)
const typeModalTitle = computed(() => (editingTypeId.value ? '编辑字典' : '新建字典'))
const typeFormRef = ref()
const typeForm = reactive({
  dict_code: '',
  label: '',
  remark: '',
  status: '1',
})

function openCreateType() {
  editingTypeId.value = null
  Object.assign(typeForm, { dict_code: '', label: '', remark: '', status: '1' })
  typeModalOpen.value = true
}

function openEditType(record: any) {
  editingTypeId.value = record.id
  Object.assign(typeForm, {
    dict_code: record.dict_code,
    label: record.label,
    remark: record.remark || '',
    status: record.status || '1',
  })
  typeModalOpen.value = true
}

async function submitType() {
  try {
    await typeFormRef.value?.validate()
  } catch {
    return
  }
  typeSubmitting.value = true
  try {
    if (editingTypeId.value) {
      await updateDictType({
        id: editingTypeId.value,
        label: typeForm.label,
        remark: typeForm.remark,
        status: typeForm.status,
      })
      message.success('更新成功')
    } else {
      await createDictType({
        dict_code: typeForm.dict_code,
        label: typeForm.label,
        remark: typeForm.remark || undefined,
        status: typeForm.status,
      })
      message.success('创建成功')
    }
    typeModalOpen.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    typeSubmitting.value = false
  }
}

async function onTypeStatusChange(record: any, checked: boolean) {
  const next = checked ? '1' : '0'
  try {
    await updateDictType({ id: record.id, status: next })
    record.status = next
    message.success('状态已更新')
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

function confirmDeleteTypes(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择字典')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个字典类型？有字典值时将失败。`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteDictType(ids)
        message.success('删除成功')
        await loadData()
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '删除失败')
        throw e
      }
    },
  })
}

// —— 字典值抽屉 ——
const itemDrawerOpen = ref(false)
const currentType = ref<any>(null)
const itemQuery = reactive({ keywords: '', status: undefined as string | undefined })
const itemPageNum = ref(1)
const itemPageSize = ref(10)
const itemTotal = ref(0)
const itemData = ref<any[]>([])
const itemLoading = ref(false)
const itemSelectedKeys = ref<string[]>([])

const itemColumns = withEllipsisColumns([
  { title: '值', dataIndex: 'value', width: 140 },
  { title: '名称', dataIndex: 'label', width: 140 },
  { title: '状态', dataIndex: 'status', width: 110 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '备注', dataIndex: 'remark', width: 160 },
  { title: '更新时间', dataIndex: 'updated_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 100, fixed: 'right' as const },
])

const itemRowSelection = computed(() => ({
  selectedRowKeys: itemSelectedKeys.value,
  onChange: (keys: (string | number)[]) => {
    itemSelectedKeys.value = keys.map(String)
  },
}))

async function loadItems() {
  if (!currentType.value) return
  itemLoading.value = true
  try {
    const kw = itemQuery.keywords.trim()
    const resp = await pageDictItem({
      page: String(itemPageNum.value),
      page_size: String(itemPageSize.value),
      parent_id: currentType.value.id,
      label: kw || undefined,
      status: itemQuery.status,
    })
    itemData.value = resp.items ?? []
    itemTotal.value = Number(resp.total || 0)
    itemSelectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载字典值失败')
  } finally {
    itemLoading.value = false
  }
}

function openItems(record: any) {
  currentType.value = record
  Object.assign(itemQuery, { keywords: '', status: undefined })
  itemPageNum.value = 1
  itemDrawerOpen.value = true
  loadItems()
}

function onItemSearch() {
  itemPageNum.value = 1
  loadItems()
}

function onItemReset() {
  Object.assign(itemQuery, { keywords: '', status: undefined })
  onItemSearch()
}

function onItemPageChange(page: number, size: number) {
  itemPageNum.value = page
  itemPageSize.value = size
  loadItems()
}

const itemFormOpen = ref(false)
const itemFormSubmitting = ref(false)
const editingItemId = ref<string | null>(null)
const itemFormRef = ref()
const itemForm = reactive({
  label: '',
  value: '',
  sort: '0',
  remark: '',
  status: '1',
})

function openCreateItem() {
  editingItemId.value = null
  Object.assign(itemForm, { label: '', value: '', sort: '0', remark: '', status: '1' })
  itemFormOpen.value = true
}

function openEditItem(record: any) {
  editingItemId.value = record.id
  Object.assign(itemForm, {
    label: record.label,
    value: record.value,
    sort: record.sort || '0',
    remark: record.remark || '',
    status: record.status || '1',
  })
  itemFormOpen.value = true
}

async function submitItem() {
  try {
    await itemFormRef.value?.validate()
  } catch {
    return
  }
  if (!currentType.value) return
  itemFormSubmitting.value = true
  try {
    if (editingItemId.value) {
      await updateDictItem({
        id: editingItemId.value,
        label: itemForm.label,
        value: itemForm.value,
        sort: itemForm.sort,
        remark: itemForm.remark,
        status: itemForm.status,
      })
      message.success('更新成功')
    } else {
      await createDictItem({
        parent_id: currentType.value.id,
        label: itemForm.label,
        value: itemForm.value,
        sort: itemForm.sort,
        remark: itemForm.remark || undefined,
        status: itemForm.status,
      })
      message.success('创建成功')
    }
    itemFormOpen.value = false
    await loadItems()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    itemFormSubmitting.value = false
  }
}

async function onItemStatusChange(record: any, checked: boolean) {
  const next = checked ? '1' : '0'
  try {
    await updateDictItem({ id: record.id, status: next })
    record.status = next
    message.success('状态已更新')
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

function confirmDeleteItems(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择字典值')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个字典值？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteDictItem(ids)
        message.success('删除成功')
        await loadItems()
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
      <AFormItem label="关键字">
        <AInput
          v-model:value="queryForm.keywords"
          placeholder="编码/名称"
          allow-clear
          style="width: 220px"
        />
      </AFormItem>
      <AFormItem label="状态">
        <ASelect
          v-model:value="queryForm.status"
          placeholder="请选择"
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
      :row-selection="rowSelection"
      row-key="id"
      size="small"
      bordered
      :scroll="{ x: 'max-content' }"
      @change="onPageChange"
    >
      <template #toolbar>
        <AButton v-privilege="'system:dict:create'" type="primary" @click="openCreateType">
          <template #icon>
            <IconParkPlus />
          </template>
          新建
        </AButton>
        <AButton
          v-privilege="'system:dict:delete'"
          danger
          :disabled="!selectedKeys.length"
          @click="confirmDeleteTypes(selectedKeys)"
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
        <template v-if="column.dataIndex === 'dict_code'">
          <AButton type="link" size="small" @click="openItems(record)">
            <EllipsisText :text="record.dict_code" />
          </AButton>
        </template>
        <template v-else-if="column.dataIndex === 'label'">
          <EllipsisText :text="record.label" />
        </template>
        <template v-else-if="column.dataIndex === 'remark'">
          <EllipsisText :text="record.remark" />
        </template>
        <template v-else-if="column.dataIndex === 'status'">
          <ASwitch
            :checked="record.status === '1'"
            checked-children="启用中"
            un-checked-children="已禁用"
            @change="(checked: boolean) => onTypeStatusChange(record, checked)"
          />
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <AButton
            v-privilege="'system:dict:update'"
            type="link"
            size="small"
            @click="openEditType(record)"
          >
            编辑
          </AButton>
          <AButton type="link" size="small" @click="openItems(record)">字典值</AButton>
        </template>
      </template>
    </PageTable>

    <AModal
      v-model:open="typeModalOpen"
      :title="typeModalTitle"
      :confirm-loading="typeSubmitting"
      destroy-on-close
      @ok="submitType"
    >
      <AForm ref="typeFormRef" :model="typeForm" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem
          label="字典编码"
          name="dict_code"
          :rules="editingTypeId ? [] : [{ required: true, message: '请输入字典编码' }]"
        >
          <AInput v-model:value="typeForm.dict_code" :disabled="!!editingTypeId" />
        </AFormItem>
        <AFormItem label="字典名称" name="label" :rules="[{ required: true, message: '请输入字典名称' }]">
          <AInput v-model:value="typeForm.label" />
        </AFormItem>
        <AFormItem label="备注" name="remark">
          <ATextarea v-model:value="typeForm.remark" :rows="3" />
        </AFormItem>
      </AForm>
    </AModal>

    <ADrawer
      v-model:open="itemDrawerOpen"
      :title="`字典值 - ${currentType?.label || ''}（${currentType?.dict_code || ''}）`"
      :size="880"
      destroy-on-close
    >
      <div class="item-panel">
        <div class="item-search">
          <AInput
            v-model:value="itemQuery.keywords"
            placeholder="关键字"
            allow-clear
            style="width: 180px"
          />
          <ASelect
            v-model:value="itemQuery.status"
            placeholder="请选择"
            allow-clear
            style="width: 120px"
            :options="[
              { label: '启用', value: '1' },
              { label: '禁用', value: '0' },
            ]"
          />
          <AButton type="primary" @click="onItemSearch">查询</AButton>
          <AButton @click="onItemReset">重置</AButton>
        </div>
        <div class="item-toolbar">
          <AButton v-privilege="'system:dict:create'" type="primary" @click="openCreateItem">
            新建
          </AButton>
          <AButton
            v-privilege="'system:dict:delete'"
            danger
            :disabled="!itemSelectedKeys.length"
            @click="confirmDeleteItems(itemSelectedKeys)"
          >
            批量删除
          </AButton>
        </div>
        <ATable
          :columns="itemColumns"
          :data-source="itemData"
          :loading="itemLoading"
          :row-selection="itemRowSelection"
          :pagination="false"
          :scroll="{ x: tableScrollX(itemColumns) }"
          table-layout="fixed"
          row-key="id"
          size="small"
          bordered
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'value'">
              <EllipsisText :text="record.value" />
            </template>
            <template v-else-if="column.dataIndex === 'label'">
              <EllipsisText :text="record.label" />
            </template>
            <template v-else-if="column.dataIndex === 'remark'">
              <EllipsisText :text="record.remark" />
            </template>
            <template v-else-if="column.dataIndex === 'status'">
              <ASwitch
                :checked="record.status === '1'"
                checked-children="启用中"
                un-checked-children="已禁用"
                @change="(checked: boolean) => onItemStatusChange(record, checked)"
              />
            </template>
            <template v-else-if="column.dataIndex === 'action'">
              <AButton
                v-privilege="'system:dict:update'"
                type="link"
                size="small"
                @click="openEditItem(record)"
              >
                编辑
              </AButton>
            </template>
          </template>
        </ATable>
        <div class="item-footer">
          <APagination
            v-model:current="itemPageNum"
            v-model:page-size="itemPageSize"
            :total="itemTotal"
            size="small"
            show-size-changer
            :show-total="(t: number) => `共 ${t} 条`"
            @change="onItemPageChange"
          />
        </div>
      </div>
    </ADrawer>

    <AModal
      v-model:open="itemFormOpen"
      :title="editingItemId ? '编辑字典值' : '新建字典值'"
      :confirm-loading="itemFormSubmitting"
      destroy-on-close
      @ok="submitItem"
    >
      <AForm ref="itemFormRef" :model="itemForm" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="值" name="value" :rules="[{ required: true, message: '请输入值' }]">
          <AInput v-model:value="itemForm.value" />
        </AFormItem>
        <AFormItem label="名称" name="label" :rules="[{ required: true, message: '请输入名称' }]">
          <AInput v-model:value="itemForm.label" />
        </AFormItem>
        <AFormItem label="排序" name="sort">
          <AInput v-model:value="itemForm.sort" />
        </AFormItem>
        <AFormItem label="备注" name="remark">
          <ATextarea v-model:value="itemForm.remark" :rows="2" />
        </AFormItem>
        <AFormItem label="状态" name="status">
          <ASelect
            v-model:value="itemForm.status"
            :options="[
              { label: '启用', value: '1' },
              { label: '禁用', value: '0' },
            ]"
          />
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

.item-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.item-search,
.item-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
