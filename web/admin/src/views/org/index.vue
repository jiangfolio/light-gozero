<script setup lang="ts">
import { message, Modal } from 'antdv-next'
import { computed, onMounted, reactive, ref } from 'vue'

import {
  createOrg,
  deleteOrg,
  treeOrg,
  updateOrg,
} from '@/api/system/org'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { withEllipsisColumns } from '@/lib/table'
import { toTreeSelectData } from '@/lib/format'

defineOptions({ name: 'SystemOrg' })

const defaultQuery = {
  name: '',
  status: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const tableData = ref<any[]>([])
const parentTreeData = ref<any[]>([])
const tableLoading = ref(false)
const selectedKeys = ref<string[]>([])
const expandedKeys = ref<string[]>([])

/** 收集树中全部节点 id，供展开全部 */
function collectIds(nodes: any[]): string[] {
  const ids: string[] = []
  const walk = (list: any[]) => {
    for (const n of list) {
      ids.push(n.id)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(nodes)
  return ids
}

const columns = withEllipsisColumns([
  { title: '组织名称', dataIndex: 'name', width: 220 },
  { title: '组织编码', dataIndex: 'code', width: 140 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 220, fixed: 'right' as const },
])

const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedKeys.value = keys.map(String)
  },
  checkStrictly: true,
}))

/** 直接拉后端组织树 */
async function loadData() {
  tableLoading.value = true
  try {
    const resp = await treeOrg({
      name: queryForm.name || undefined,
      status: queryForm.status,
    })
    const tree = resp.items ?? []
    tableData.value = tree
    parentTreeData.value = toTreeSelectData(tree)
    expandedKeys.value = collectIds(tree)
    selectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载组织失败')
  } finally {
    tableLoading.value = false
  }
}

function onSearch() {
  loadData()
}

function onReset() {
  Object.assign(queryForm, defaultQuery)
  loadData()
}

// —— 新建 / 编辑弹窗 ——
const modalOpen = ref(false)
const modalSubmitting = ref(false)
const editingId = ref<string | null>(null)
const modalTitle = computed(() => (editingId.value ? '编辑组织' : '新建组织'))
const formRef = ref()
const formState = reactive({
  parent_id: undefined as string | undefined,
  name: '',
  code: '',
  sort: '0',
  status: '1',
})

/** 打开新建（可选指定父级） */
function openCreate(parentId = '') {
  editingId.value = null
  Object.assign(formState, {
    parent_id: parentId || undefined,
    name: '',
    code: '',
    sort: '0',
    status: '1',
  })
  modalOpen.value = true
}

/** 打开编辑 */
function openEdit(record: any) {
  editingId.value = record.id
  Object.assign(formState, {
    parent_id: record.parent_id || undefined,
    name: record.name,
    code: record.code,
    sort: record.sort || '0',
    status: record.status || '1',
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
    // 1. 按新建/编辑分流提交
    if (editingId.value) {
      await updateOrg({
        id: editingId.value,
        parent_id: formState.parent_id || undefined,
        name: formState.name,
        code: formState.code,
        sort: formState.sort,
        status: formState.status,
      })
      message.success('更新成功')
    } else {
      await createOrg({
        parent_id: formState.parent_id || undefined,
        name: formState.name,
        code: formState.code,
        sort: formState.sort,
        status: formState.status,
      })
      message.success('创建成功')
    }
    // 2. 关闭弹窗并刷新树
    modalOpen.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    modalSubmitting.value = false
  }
}

/** 删除单个或批量 */
function confirmDelete(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择组织')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个组织？有子组织或用户时将失败。`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteOrg(ids)
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

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="组织名称">
        <AInput
          v-model:value="queryForm.name"
          placeholder="请输入组织名称"
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
      :columns="columns"
      :data-source="tableData"
      :loading="tableLoading"
      :row-selection="rowSelection"
      row-key="id"
      size="small"
      bordered
      :scroll="{ x: 'max-content' }"
      v-model:expanded-row-keys="expandedKeys"
    >
      <template #toolbar>
        <AButton v-privilege="'system:org:create'" type="primary" @click="openCreate()">
          <template #icon>
            <IconParkPlus />
          </template>
          新建
        </AButton>
        <AButton
          v-privilege="'system:org:delete'"
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
        <template v-if="column.dataIndex === 'status'">
          <ATag :color="record.status === '1' ? 'green' : 'red'">
            {{ record.status === '1' ? '启用' : '禁用' }}
          </ATag>
        </template>
        <template v-if="column.dataIndex === 'action'">
          <AButton
            v-privilege="'system:org:create'"
            type="link"
            size="small"
            @click="openCreate(record.id)"
          >
            添加下级
          </AButton>
          <AButton
            v-privilege="'system:org:update'"
            type="link"
            size="small"
            @click="openEdit(record)"
          >
            编辑
          </AButton>
          <AButton
            v-privilege="'system:org:delete'"
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
      @ok="submitForm"
    >
      <AForm ref="formRef" :model="formState" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="上级组织" name="parent_id">
          <ATreeSelect
            v-model:value="formState.parent_id"
            :tree-data="parentTreeData"
            tree-default-expand-all
            allow-clear
            placeholder="不选则为顶级组织"
            :field-names="{ label: 'title', value: 'value', children: 'children' }"
          />
        </AFormItem>
        <AFormItem label="组织名称" name="name" :rules="[{ required: true, message: '请输入组织名称' }]">
          <AInput v-model:value="formState.name" placeholder="请输入组织名称" />
        </AFormItem>
        <AFormItem label="组织编码" name="code" :rules="[{ required: true, message: '请输入组织编码' }]">
          <AInput v-model:value="formState.code" placeholder="请输入组织编码" :disabled="!!editingId" />
        </AFormItem>
        <AFormItem label="排序" name="sort">
          <AInput v-model:value="formState.sort" placeholder="数字越小越靠前" />
        </AFormItem>
        <AFormItem label="状态" name="status">
          <ASelect
            v-model:value="formState.status"
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
</style>
