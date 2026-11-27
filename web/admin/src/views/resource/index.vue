<script setup lang="ts">
import { message, Modal } from 'antdv-next'
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import {
  createResource,
  deleteResource,
  listResource,
  treeResource,
  updateResource,
} from '@/api/system/resource'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import IconPicker from '@/components/icon-picker/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { toTreeSelectData } from '@/lib/format'
import { refreshMenus } from '@/lib/refresh-menus'
import { tableScrollX, withEllipsisColumns } from '@/lib/table'

defineOptions({ name: 'SystemResource' })

const router = useRouter()

const defaultQuery = {
  name: '',
  type: undefined as string | undefined,
  status: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

/** 后端树仅到菜单；按钮按菜单 id 另请求 */
const tableData = ref<any[]>([])
const parentTreeData = ref<any[]>([])
const tableLoading = ref(false)
const selectedKeys = ref<string[]>([])
const expandedKeys = ref<string[]>([])

const typeColor: Record<string, string> = {
  directory: 'blue',
  menu: 'orange',
  button: 'green',
}
const typeLabel: Record<string, string> = {
  directory: '目录',
  menu: '菜单',
  button: '按钮',
}

const columns = withEllipsisColumns([
  { title: '显示名称', dataIndex: 'name', width: 220 },
  { title: '类型', dataIndex: 'type', width: 90 },
  { title: '路由地址', dataIndex: 'path', width: 160 },
  { title: '组件', dataIndex: 'component_path', width: 180 },
  { title: '是否可见', dataIndex: 'hidden', width: 100 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '操作', dataIndex: 'action', width: 220, fixed: 'right' as const },
])

const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedKeys.value = keys.map(String)
  },
  checkStrictly: true,
}))

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

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await treeResource({
      name: queryForm.name || undefined,
      type: queryForm.type,
      status: queryForm.status,
    })
    const tree = resp.items ?? []
    tableData.value = tree
    parentTreeData.value = toTreeSelectData(tree)
    expandedKeys.value = collectIds(tree)
    selectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载资源失败')
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

// —— 目录 / 菜单 弹窗 ——
const formOpen = ref(false)
const formSubmitting = ref(false)
const editingId = ref<string | null>(null)
const formTitle = computed(() => (editingId.value ? '编辑资源' : '新建资源'))
const formRef = ref()
const formState = reactive({
  parent_id: undefined as string | undefined,
  name: '',
  type: 'menu',
  perm_code: '',
  path: '',
  component_path: '',
  redirect: '',
  route_name: '',
  icon: '',
  hidden: '0',
  keep_alive: '0',
  affix: '0',
  sort: '0',
  status: '1',
})

/** 排序数字与接口字符串字段互转 */
const sortNumber = computed({
  get: () => {
    const n = Number(formState.sort)
    return Number.isFinite(n) ? n : 0
  },
  set: (v: number | null) => {
    formState.sort = String(v ?? 0)
  },
})

function resetForm(parentId?: string, type = 'menu') {
  Object.assign(formState, {
    parent_id: parentId || undefined,
    name: '',
    type,
    perm_code: '',
    path: '',
    component_path: '',
    redirect: '',
    route_name: '',
    icon: '',
    hidden: '0',
    keep_alive: '0',
    affix: '0',
    sort: '0',
    status: '1',
  })
}

function openCreate(parentId = '', type = 'menu') {
  editingId.value = null
  resetForm(parentId || undefined, type)
  formOpen.value = true
}

function openEdit(record: any) {
  editingId.value = record.id
  Object.assign(formState, {
    parent_id: record.parent_id || undefined,
    name: record.name,
    type: record.type,
    perm_code: record.perm_code || '',
    path: record.path || '',
    component_path: record.component_path || '',
    redirect: record.redirect || '',
    route_name: record.route_name || '',
    icon: record.icon || '',
    hidden: record.hidden || '0',
    keep_alive: record.keep_alive || '0',
    affix: record.affix || '0',
    sort: record.sort || '0',
    status: record.status || '1',
  })
  formOpen.value = true
}

async function submitForm() {
  // 校验失败时 reject，避免 Modal 自动关闭
  await formRef.value?.validate()
  formSubmitting.value = true
  try {
    const payload = {
      parent_id: formState.parent_id || undefined,
      name: formState.name,
      type: formState.type,
      path: formState.path || undefined,
      component_path: formState.component_path || undefined,
      redirect: formState.redirect || undefined,
      route_name: formState.route_name || undefined,
      icon: formState.icon || undefined,
      hidden: formState.hidden,
      keep_alive: formState.keep_alive,
      affix: formState.affix,
      sort: String(formState.sort ?? '0'),
      status: formState.status,
    }
    if (editingId.value) {
      await updateResource({ id: editingId.value, ...payload })
      message.success('更新成功')
    } else {
      await createResource(payload)
      message.success('创建成功')
    }
    formOpen.value = false
    await loadData()
    await refreshMenus(router)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
    throw e
  } finally {
    formSubmitting.value = false
  }
}

function confirmDelete(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择资源')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个资源？有子节点时将失败。`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteResource(ids)
        message.success('删除成功')
        await loadData()
        await refreshMenus(router)
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '删除失败')
        throw e
      }
    },
  })
}

// —— 菜单「更多 → 按钮权限」：按菜单 id 一次性 list 拉取 ——
const btnDrawerOpen = ref(false)
const currentMenu = ref<any>(null)
const btnLoading = ref(false)
const btnList = ref<any[]>([])

const btnColumns = withEllipsisColumns([
  { title: '名称', dataIndex: 'name', width: 160 },
  { title: '编码', dataIndex: 'perm_code', width: 220 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '操作', dataIndex: 'action', width: 140, fixed: 'right' as const },
])

async function loadButtons() {
  if (!currentMenu.value) return
  btnLoading.value = true
  try {
    const resp = await listResource({
      parent_id: currentMenu.value.id,
      type: 'button',
    })
    btnList.value = resp.items ?? []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载按钮失败')
  } finally {
    btnLoading.value = false
  }
}

function openButtons(record: any) {
  currentMenu.value = record
  btnDrawerOpen.value = true
  loadButtons()
}

const btnFormOpen = ref(false)
const btnFormSubmitting = ref(false)
const editingBtnId = ref<string | null>(null)
const btnFormRef = ref()
const btnForm = reactive({
  name: '',
  perm_code: '',
  sort: '0',
  status: '1',
})

function openCreateBtn() {
  editingBtnId.value = null
  Object.assign(btnForm, { name: '', perm_code: '', sort: '0', status: '1' })
  btnFormOpen.value = true
}

function openEditBtn(record: any) {
  editingBtnId.value = record.id
  Object.assign(btnForm, {
    name: record.name,
    perm_code: record.perm_code || '',
    sort: record.sort || '0',
    status: record.status || '1',
  })
  btnFormOpen.value = true
}

async function submitBtn() {
  try {
    await btnFormRef.value?.validate()
  } catch {
    return
  }
  if (!currentMenu.value) return
  btnFormSubmitting.value = true
  try {
    const payload = {
      parent_id: currentMenu.value.id,
      name: btnForm.name,
      type: 'button',
      perm_code: btnForm.perm_code,
      sort: btnForm.sort,
      status: btnForm.status,
      hidden: '0',
      keep_alive: '0',
      affix: '0',
    }
    if (editingBtnId.value) {
      await updateResource({ id: editingBtnId.value, ...payload })
      message.success('更新成功')
    } else {
      await createResource(payload)
      message.success('创建成功')
    }
    btnFormOpen.value = false
    await loadButtons()
    await refreshMenus(router)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    btnFormSubmitting.value = false
  }
}

function confirmDeleteBtn(ids: string[]) {
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个按钮？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteResource(ids)
        message.success('删除成功')
        await loadButtons()
        await refreshMenus(router)
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '删除失败')
        throw e
      }
    },
  })
}

const menuMoreItems = (record: any) => ({
  items: [{ key: 'buttons', label: '按钮权限' }],
  onClick: ({ key }: { key: string | number }) => {
    if (String(key) === 'buttons') openButtons(record)
  },
})

usePageReload(loadData)
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="名称">
        <AInput
          v-model:value="queryForm.name"
          placeholder="请输入资源名称"
          allow-clear
          style="width: 200px"
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
          ]"
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
        <AButton v-privilege="'system:resource:create'" type="primary" @click="openCreate('', 'directory')">
          <template #icon>
            <IconParkPlus />
          </template>
          新建目录
        </AButton>
        <AButton v-privilege="'system:resource:create'" @click="openCreate('', 'menu')">新建菜单</AButton>
        <AButton
          v-privilege="'system:resource:delete'"
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
        <template v-if="column.dataIndex === 'type'">
          <ATag :color="typeColor[record.type] || 'default'">
            {{ typeLabel[record.type] || record.type }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'path'">
          <EllipsisText :text="record.path" />
        </template>
        <template v-else-if="column.dataIndex === 'component_path'">
          <EllipsisText :text="record.component_path" />
        </template>
        <template v-else-if="column.dataIndex === 'name'">
          <EllipsisText :text="record.name" />
        </template>
        <template v-else-if="column.dataIndex === 'hidden'">
          <ATag>{{ record.hidden === '1' ? '否' : '是' }}</ATag>
        </template>
        <template v-else-if="column.dataIndex === 'status'">
          <ATag :color="record.status === '1' ? 'green' : 'red'">
            {{ record.status === '1' ? '启用' : '禁用' }}
          </ATag>
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <AButton
            v-if="record.type === 'directory'"
            v-privilege="'system:resource:create'"
            type="link"
            size="small"
            @click="openCreate(record.id, 'menu')"
          >
            添加下级
          </AButton>
          <AButton
            v-privilege="'system:resource:update'"
            type="link"
            size="small"
            @click="openEdit(record)"
          >
            编辑
          </AButton>
          <AButton
            v-privilege="'system:resource:delete'"
            type="link"
            size="small"
            danger
            @click="confirmDelete([record.id])"
          >
            删除
          </AButton>
          <ADropdown
            v-if="record.type === 'menu'"
            :trigger="['click']"
            :menu="menuMoreItems(record)"
          >
            <AButton type="link" size="small">更多</AButton>
          </ADropdown>
        </template>
      </template>
    </PageTable>

    <AModal
      v-model:open="formOpen"
      :title="formTitle"
      :width="880"
      :confirm-loading="formSubmitting"
      destroy-on-close
      ok-text="确定"
      cancel-text="取消"
      @ok="submitForm"
    >
      <AForm ref="formRef" :model="formState" layout="vertical" class="resource-form">
        <ARow :gutter="[16, 0]">
          <ACol :xs="24" :md="12">
            <AFormItem
              label="显示名称"
              name="name"
              :rules="[{ required: true, message: '请输入名称' }]"
            >
              <AInput v-model:value="formState.name" placeholder="请输入显示名称" />
            </AFormItem>
          </ACol>
          <ACol :xs="24" :md="12">
            <AFormItem
              label="资源类型"
              name="type"
              :rules="[{ required: true, message: '请选择类型' }]"
            >
              <ARadioGroup
                v-model:value="formState.type"
                button-style="solid"
                :disabled="!!editingId"
              >
                <ARadioButton value="directory">目录</ARadioButton>
                <ARadioButton value="menu">菜单</ARadioButton>
              </ARadioGroup>
            </AFormItem>
          </ACol>

          <ACol :xs="24" :md="12">
            <AFormItem label="上级资源" name="parent_id">
              <ATreeSelect
                v-model:value="formState.parent_id"
                :tree-data="parentTreeData"
                tree-default-expand-all
                allow-clear
                placeholder="不选则为顶级"
                :field-names="{ label: 'title', value: 'value', children: 'children' }"
              />
            </AFormItem>
          </ACol>
          <ACol :xs="24" :md="12">
            <AFormItem label="路由地址" name="path">
              <AInput v-model:value="formState.path" placeholder="如 /system/org" />
            </AFormItem>
          </ACol>

          <ACol :xs="24" :md="12">
            <AFormItem label="组件地址" name="component_path">
              <AInput v-model:value="formState.component_path" placeholder="如 org/index" />
            </AFormItem>
          </ACol>
          <ACol :xs="24" :md="12">
            <AFormItem label="路由名" name="route_name">
              <AInput v-model:value="formState.route_name" placeholder="可选，建议与页面 name 一致" />
            </AFormItem>
          </ACol>

          <ACol :xs="24" :md="12">
            <AFormItem label="重定向" name="redirect">
              <AInput v-model:value="formState.redirect" placeholder="目录跳转路径，可选" />
            </AFormItem>
          </ACol>

          <ACol :xs="24" :md="12">
            <AFormItem label="图标" name="icon">
              <IconPicker v-model:value="formState.icon" placeholder="请选择图标" />
            </AFormItem>
          </ACol>
          <ACol :xs="24" :md="12">
            <AFormItem label="排序" name="sort">
              <AInputNumber
                v-model:value="sortNumber"
                :min="0"
                :precision="0"
                class="field-full"
                placeholder="数字越小越靠前"
              />
            </AFormItem>
          </ACol>

        </ARow>

        <!-- 可见 / 缓存 / 状态：固定三列同行，不受断点换行 -->
        <ARow :gutter="16" class="flag-row" :wrap="false">
          <ACol :span="8">
            <AFormItem label="是否可见" name="hidden">
              <ARadioGroup v-model:value="formState.hidden" button-style="solid">
                <ARadioButton value="0">是</ARadioButton>
                <ARadioButton value="1">否</ARadioButton>
              </ARadioGroup>
            </AFormItem>
          </ACol>
          <ACol :span="8">
            <AFormItem label="是否缓存" name="keep_alive">
              <ARadioGroup v-model:value="formState.keep_alive" button-style="solid">
                <ARadioButton value="1">是</ARadioButton>
                <ARadioButton value="0">否</ARadioButton>
              </ARadioGroup>
            </AFormItem>
          </ACol>
          <ACol :span="8">
            <AFormItem label="状态" name="status">
              <ARadioGroup v-model:value="formState.status" button-style="solid">
                <ARadioButton value="1">启用</ARadioButton>
                <ARadioButton value="0">禁用</ARadioButton>
              </ARadioGroup>
            </AFormItem>
          </ACol>
        </ARow>
      </AForm>
    </AModal>

    <ADrawer
      v-model:open="btnDrawerOpen"
      :title="`按钮权限 - ${currentMenu?.name || ''}`"
      :size="780"
      destroy-on-close
    >
      <div class="btn-panel">
        <div class="btn-toolbar">
          <AButton v-privilege="'system:resource:create'" type="primary" @click="openCreateBtn">
            <template #icon>
              <IconParkPlus />
            </template>
            新增按钮
          </AButton>
          <AButton type="text" :loading="btnLoading" @click="loadButtons">
            <template #icon>
              <IconParkRefresh />
            </template>
          </AButton>
        </div>
        <ATable
          :columns="btnColumns"
          :data-source="btnList"
          :loading="btnLoading"
          :pagination="false"
          :children-column-name="'__no_children__'"
          :scroll="{ x: tableScrollX(btnColumns) }"
          table-layout="fixed"
          row-key="id"
          size="small"
          bordered
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'name'">
              <EllipsisText :text="record.name" />
            </template>
            <template v-else-if="column.dataIndex === 'perm_code'">
              <EllipsisText :text="record.perm_code" />
            </template>
            <template v-else-if="column.dataIndex === 'status'">
              <ATag :color="record.status === '1' ? 'green' : 'red'">
                {{ record.status === '1' ? '启用' : '禁用' }}
              </ATag>
            </template>
            <template v-else-if="column.dataIndex === 'action'">
              <AButton
                v-privilege="'system:resource:update'"
                type="link"
                size="small"
                @click="openEditBtn(record)"
              >
                编辑
              </AButton>
              <AButton
                v-privilege="'system:resource:delete'"
                type="link"
                size="small"
                danger
                @click="confirmDeleteBtn([record.id])"
              >
                删除
              </AButton>
            </template>
          </template>
        </ATable>
      </div>
    </ADrawer>

    <AModal
      v-model:open="btnFormOpen"
      :title="editingBtnId ? '编辑按钮' : '新增按钮'"
      :confirm-loading="btnFormSubmitting"
      destroy-on-close
      @ok="submitBtn"
    >
      <AForm ref="btnFormRef" :model="btnForm" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="名称" name="name" :rules="[{ required: true, message: '请输入名称' }]">
          <AInput v-model:value="btnForm.name" />
        </AFormItem>
        <AFormItem
          label="编码"
          name="perm_code"
          :rules="[{ required: true, message: '请输入权限编码' }]"
        >
          <AInput v-model:value="btnForm.perm_code" placeholder="module:resource:action" />
        </AFormItem>
        <AFormItem label="排序" name="sort">
          <AInput v-model:value="btnForm.sort" />
        </AFormItem>
        <AFormItem label="状态" name="status">
          <ASelect
            v-model:value="btnForm.status"
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

.resource-form {
  overflow-x: hidden;
}

.resource-form :deep(.ant-form-item) {
  margin-bottom: 12px;
}

.resource-form :deep(.ant-col) {
  min-width: 0;
}

.flag-row {
  flex-wrap: nowrap;
}

.flag-row :deep(.ant-radio-group) {
  display: flex;
  flex-wrap: nowrap;
}

.field-full {
  width: 100%;
}

.btn-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.btn-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
