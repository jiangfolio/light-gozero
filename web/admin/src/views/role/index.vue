<script setup lang="ts">
import { message, Modal } from 'antdv-next'
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { treeOrg } from '@/api/system/org'
import { listResource, treeResource } from '@/api/system/resource'
import {
  bindRoleResources,
  createRole,
  deleteRole,
  getRoleDetail,
  pageRole,
  unbindRoleResources,
  updateRole,
} from '@/api/system/role'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { toTreeSelectData } from '@/lib/format'
import { refreshMenus } from '@/lib/refresh-menus'
import { tableScrollX, withEllipsisColumns } from '@/lib/table'

defineOptions({ name: 'SystemRole' })

const router = useRouter()

const dataScopeOptions = [
  { label: '仅本人', value: '1' },
  { label: '本部门', value: '2' },
  { label: '本部门及以下', value: '3' },
  { label: '自定义部门', value: '4' },
  { label: '全部数据', value: '5' },
]
const dataScopeLabel: Record<string, string> = Object.fromEntries(
  dataScopeOptions.map((o) => [o.value, o.label]),
)

const defaultQuery = {
  name: '',
  code: '',
  status: undefined as string | undefined,
}
const queryForm = reactive({ ...defaultQuery })

const pageNum = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref<any[]>([])
const tableLoading = ref(false)
const selectedKeys = ref<string[]>([])
const orgTreeSelectData = ref<any[]>([])

const columns = withEllipsisColumns([
  { title: '角色名称', dataIndex: 'name', width: 160 },
  { title: '角色编码', dataIndex: 'code', width: 140 },
  { title: '数据范围', dataIndex: 'data_scope', width: 140 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '备注', dataIndex: 'remark', width: 180 },
  { title: '创建时间', dataIndex: 'created_at', width: 200 },
  { title: '操作', dataIndex: 'action', width: 220, fixed: 'right' as const },
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
    const resp = await pageRole({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      name: queryForm.name || undefined,
      code: queryForm.code || undefined,
      status: queryForm.status,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
    selectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载角色失败')
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

// —— 新建 / 编辑 ——
const modalOpen = ref(false)
const modalSubmitting = ref(false)
const editingId = ref<string | null>(null)
const modalTitle = computed(() => (editingId.value ? '编辑角色' : '新建角色'))
const formRef = ref()
const formState = reactive({
  name: '',
  code: '',
  remark: '',
  status: '1',
})

function openCreate() {
  editingId.value = null
  Object.assign(formState, { name: '', code: '', remark: '', status: '1' })
  modalOpen.value = true
}

function openEdit(record: any) {
  editingId.value = record.id
  Object.assign(formState, {
    name: record.name,
    code: record.code,
    remark: record.remark || '',
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
    if (editingId.value) {
      const detail = await getRoleDetail(editingId.value)
      await updateRole({
        id: editingId.value,
        name: formState.name,
        code: formState.code,
        remark: formState.remark || undefined,
        status: formState.status,
        data_scope: detail.data_scope || '5',
        org_ids: detail.data_scope === '4' ? detail.org_ids ?? [] : [],
      })
      message.success('更新成功')
    } else {
      await createRole({
        name: formState.name,
        code: formState.code,
        remark: formState.remark || undefined,
        status: formState.status,
        data_scope: '5',
        org_ids: [],
      })
      message.success('创建成功')
    }
    modalOpen.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    modalSubmitting.value = false
  }
}

function confirmDelete(ids: string[]) {
  if (!ids.length) {
    message.warning('请先选择角色')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个角色？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteRole(ids)
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

// —— 授权下拉 ——
function authMenu(record: any) {
  return {
    items: [
      { key: 'resource', label: '授权资源' },
      { key: 'scope', label: '数据范围' },
    ],
    onClick: ({ key }: { key: string | number }) => {
      const k = String(key)
      if (k === 'resource') openGrantResource(record)
      else if (k === 'scope') openGrantScope(record)
    },
  }
}

// —— 授权资源：目录 | 菜单 | 按钮 矩阵 ——
type GrantRow = {
  key: string
  dirId: string
  dirName: string
  menuId: string
  menuName: string
  buttons: Array<{ id: string; name: string }>
  dirRowSpan: number
}

const grantOpen = ref(false)
const grantSubmitting = ref(false)
const grantRole = ref<any>(null)
const grantRows = ref<GrantRow[]>([])
const checkedIds = ref<Set<string>>(new Set())
const originIds = ref<string[]>([])
const grantLoading = ref(false)

/** 将菜单树 + 按钮列表压成「一级目录 + 菜单 + 按钮」矩阵行 */
function buildGrantRows(nodes: any[], buttons: any[]): GrantRow[] {
  const buttonsByMenu: Record<string, Array<{ id: string; name: string }>> = {}
  for (const b of buttons) {
    const pid = b.parent_id || ''
    if (!pid) continue
    if (!buttonsByMenu[pid]) buttonsByMenu[pid] = []
    buttonsByMenu[pid].push({ id: b.id, name: b.name })
  }

  const rows: GrantRow[] = []
  const dirs = nodes.filter((n) => n.type === 'directory')
  const orphanMenus = nodes.filter((n) => n.type === 'menu')

  const pushDir = (dir: any) => {
    const menus = collectMenus(dir)
    const start = rows.length
    for (const menu of menus) {
      rows.push({
        key: menu.id,
        dirId: dir.id,
        dirName: dir.name,
        menuId: menu.id,
        menuName: menu.name,
        buttons: buttonsByMenu[menu.id] ?? [],
        dirRowSpan: 0,
      })
    }
    const count = rows.length - start
    if (count > 0) {
      rows[start].dirRowSpan = count
      for (let i = start + 1; i < rows.length; i++) rows[i].dirRowSpan = 0
    }
  }

  for (const d of dirs) pushDir(d)
  if (orphanMenus.length) {
    pushDir({ id: '__root__', name: '未分组', children: orphanMenus, type: 'directory' })
  }
  return rows
}

/** 收集目录下全部菜单（含子目录中的菜单） */
function collectMenus(node: any): any[] {
  const out: any[] = []
  const walk = (n: any) => {
    for (const c of n.children ?? []) {
      if (c.type === 'menu') out.push(c)
      else if (c.type === 'directory') walk(c)
    }
  }
  walk(node)
  return out
}

/** 从整棵资源树收集全部菜单节点 */
function collectAllMenus(nodes: any[]): any[] {
  const out: any[] = []
  const walk = (list: any[]) => {
    for (const n of list) {
      if (n.type === 'menu') out.push(n)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(nodes)
  return out
}

/** Checkbox change 兼容 boolean / event */
function toChecked(v: any) {
  if (typeof v === 'boolean') return v
  return !!v?.target?.checked
}

function isChecked(id: string) {
  return checkedIds.value.has(id)
}

function menuIdsOfDir(dirId: string) {
  return grantRows.value.filter((r) => r.dirId === dirId).map((r) => r.menuId)
}

function buttonIdsOfMenu(menuId: string) {
  const row = grantRows.value.find((r) => r.menuId === menuId)
  return row ? row.buttons.map((b) => b.id) : []
}

/** 菜单节点资源：菜单本身 + 其下全部按钮 */
function menuResourceIds(menuId: string) {
  return [menuId, ...buttonIdsOfMenu(menuId)]
}

/** 目录节点资源：目录本身（虚拟根除外）+ 下属全部菜单资源 */
function dirResourceIds(dirId: string) {
  const menus = menuIdsOfDir(dirId).flatMap((mid) => menuResourceIds(mid))
  if (!dirId || dirId === '__root__') return menus
  return [dirId, ...menus]
}

function countChecked(ids: string[]) {
  let n = 0
  for (const id of ids) {
    if (checkedIds.value.has(id)) n += 1
  }
  return n
}

/** 菜单全选：菜单 + 全部按钮都在选中集 */
function menuChecked(menuId: string) {
  const ids = menuResourceIds(menuId)
  return ids.length > 0 && countChecked(ids) === ids.length
}

/** 菜单半选：有选中但未全选 */
function menuIndeterminate(menuId: string) {
  const ids = menuResourceIds(menuId)
  const n = countChecked(ids)
  return n > 0 && n < ids.length
}

/** 目录全选：目录及下属菜单/按钮全部选中 */
function dirChecked(dirId: string) {
  const ids = dirResourceIds(dirId)
  return ids.length > 0 && countChecked(ids) === ids.length
}

/** 目录半选：下属有选中但未全选 */
function dirIndeterminate(dirId: string) {
  const ids = dirResourceIds(dirId)
  const n = countChecked(ids)
  return n > 0 && n < ids.length
}

/** 目录下是否仍有任意选中资源（不含目录 id 自身） */
function dirHasSelection(next: Set<string>, dirId: string) {
  return menuIdsOfDir(dirId).some((mid) => menuResourceIds(mid).some((id) => next.has(id)))
}

/** 按目录下选中情况写入/清除目录 id（保证资源树父节点完整） */
function syncDirId(next: Set<string>, dirId: string) {
  if (!dirId || dirId === '__root__') return
  if (dirHasSelection(next, dirId)) next.add(dirId)
  else next.delete(dirId)
}

/** 勾选/取消目录：整枝联动 */
function onDirCheck(dirId: string, checked: boolean) {
  const next = new Set(checkedIds.value)
  for (const id of dirResourceIds(dirId)) {
    if (checked) next.add(id)
    else next.delete(id)
  }
  checkedIds.value = next
}

/** 勾选/取消菜单：菜单 + 全部按钮联动，再校正目录 */
function onMenuCheck(menuId: string, dirId: string, checked: boolean) {
  const next = new Set(checkedIds.value)
  for (const id of menuResourceIds(menuId)) {
    if (checked) next.add(id)
    else next.delete(id)
  }
  syncDirId(next, dirId)
  checkedIds.value = next
}

/** 勾选/取消单个按钮：有按钮则保留菜单；全无则去掉菜单；再校正目录 */
function onButtonCheck(btnId: string, menuId: string, dirId: string, checked: boolean) {
  const next = new Set(checkedIds.value)
  if (checked) next.add(btnId)
  else next.delete(btnId)

  const btns = buttonIdsOfMenu(menuId)
  const anyBtn = btns.some((id) => next.has(id))
  if (anyBtn) next.add(menuId)
  else next.delete(menuId)

  syncDirId(next, dirId)
  checkedIds.value = next
}

async function openGrantResource(record: any) {
  grantRole.value = record
  grantLoading.value = true
  grantOpen.value = true
  try {
    // 树只到菜单；每个菜单用 list 按 parent_id 一次性拉按钮
    const [treeResp, detail] = await Promise.all([treeResource(), getRoleDetail(record.id)])
    const menus = collectAllMenus(treeResp.items ?? [])
    const buttonChunks = await Promise.all(
      menus.map((m) => listResource({ parent_id: m.id, type: 'button' })),
    )
    const buttons = buttonChunks.flatMap((r) => r.items ?? [])
    grantRows.value = buildGrantRows(treeResp.items ?? [], buttons)
    originIds.value = detail.resource_ids ?? []
    // 回填时补齐父级：有按钮则带上菜单，有下属则带上目录，保证勾选态与保存一致
    const next = new Set(originIds.value as string[])
    for (const row of grantRows.value) {
      const anyBtn = row.buttons.some((b) => next.has(b.id))
      if (anyBtn) next.add(row.menuId)
      syncDirId(next, row.dirId)
    }
    checkedIds.value = next
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载资源失败')
    grantOpen.value = false
  } finally {
    grantLoading.value = false
  }
}

async function submitGrantResource() {
  if (!grantRole.value) return
  grantSubmitting.value = true
  try {
    const selected = [...checkedIds.value]
    const origin = originIds.value
    const toAdd = selected.filter((id) => !origin.includes(id))
    const toRemove = origin.filter((id) => !selected.includes(id))
    if (toAdd.length) await bindRoleResources(grantRole.value.id, toAdd)
    if (toRemove.length) await unbindRoleResources(grantRole.value.id, toRemove)
    message.success('资源授权成功')
    grantOpen.value = false
    // 角色资源变更会影响当前用户菜单，立即重拉
    await refreshMenus(router)
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '授权失败')
  } finally {
    grantSubmitting.value = false
  }
}

const grantColumns = withEllipsisColumns([
  {
    title: '一级目录',
    dataIndex: 'dir',
    width: 160,
    customCell: (row: GrantRow) => ({ rowSpan: row.dirRowSpan }),
  },
  { title: '菜单', dataIndex: 'menu', width: 180 },
  { title: '按钮授权', dataIndex: 'buttons', ellipsis: false },
])

// —— 数据范围 ——
const scopeOpen = ref(false)
const scopeSubmitting = ref(false)
const scopeRole = ref<any>(null)
const scopeForm = reactive({
  data_scope: '5',
  org_ids: [] as string[],
})

async function openGrantScope(record: any) {
  scopeRole.value = record
  try {
    const detail = await getRoleDetail(record.id)
    Object.assign(scopeForm, {
      data_scope: detail.data_scope || '5',
      org_ids: detail.org_ids ?? [],
    })
    scopeOpen.value = true
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载数据范围失败')
  }
}

async function submitGrantScope() {
  if (!scopeRole.value) return
  scopeSubmitting.value = true
  try {
    await updateRole({
      id: scopeRole.value.id,
      name: scopeRole.value.name,
      code: scopeRole.value.code,
      remark: scopeRole.value.remark || undefined,
      status: scopeRole.value.status || '1',
      data_scope: scopeForm.data_scope,
      org_ids: scopeForm.data_scope === '4' ? scopeForm.org_ids : [],
    })
    message.success('数据范围已保存')
    scopeOpen.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    scopeSubmitting.value = false
  }
}

async function bootstrap() {
  try {
    const resp = await treeOrg()
    orgTreeSelectData.value = toTreeSelectData(resp.items ?? [])
  } catch {
    // ignore
  }
  await loadData()
}

usePageReload(bootstrap)
</script>

<template>
  <div class="page">
    <PageSearch @search="onSearch" @reset="onReset">
      <AFormItem label="角色名称">
        <AInput
          v-model:value="queryForm.name"
          placeholder="请输入角色名称"
          allow-clear
          style="width: 180px"
        />
      </AFormItem>
      <AFormItem label="角色编码">
        <AInput
          v-model:value="queryForm.code"
          placeholder="请输入角色编码"
          allow-clear
          style="width: 180px"
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
      :row-selection="rowSelection"
      row-key="id"
      size="small"
      bordered
      :scroll="{ x: 'max-content' }"
      @change="onPageChange"
    >
      <template #toolbar>
        <AButton v-privilege="'system:role:create'" type="primary" @click="openCreate">
          <template #icon>
            <IconParkPlus />
          </template>
          新建
        </AButton>
        <AButton
          v-privilege="'system:role:delete'"
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
        <template v-if="column.dataIndex === 'data_scope'">
          <ATag>{{ dataScopeLabel[record.data_scope] || record.data_scope }}</ATag>
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
            v-privilege="'system:role:update'"
            type="link"
            size="small"
            @click="openEdit(record)"
          >
            编辑
          </AButton>
          <AButton
            v-privilege="'system:role:delete'"
            type="link"
            size="small"
            danger
            @click="confirmDelete([record.id])"
          >
            删除
          </AButton>
          <ADropdown
            v-privilege="'system:role:bind_resources'"
            :trigger="['click']"
            :menu="authMenu(record)"
          >
            <AButton type="link" size="small">授权</AButton>
          </ADropdown>
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
        <AFormItem label="角色名称" name="name" :rules="[{ required: true, message: '请输入角色名称' }]">
          <AInput v-model:value="formState.name" />
        </AFormItem>
        <AFormItem label="角色编码" name="code" :rules="[{ required: true, message: '请输入角色编码' }]">
          <AInput v-model:value="formState.code" :disabled="!!editingId" />
        </AFormItem>
        <AFormItem label="备注" name="remark">
          <ATextarea v-model:value="formState.remark" :rows="3" />
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

    <AModal
      v-model:open="grantOpen"
      :title="`授权资源 - ${grantRole?.name || ''}`"
      :confirm-loading="grantSubmitting"
      width="960px"
      destroy-on-close
      :styles="{ body: { maxHeight: 'none', overflow: 'hidden' } }"
      @ok="submitGrantResource"
    >
      <ASpin :spinning="grantLoading">
        <ATable
          :columns="grantColumns"
          :data-source="grantRows"
          :pagination="false"
          row-key="key"
          size="small"
          bordered
          table-layout="fixed"
          :scroll="{ x: tableScrollX(grantColumns, 720), y: 'calc(100vh - 280px)' }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'dir'">
              <ACheckbox
                :checked="dirChecked(record.dirId)"
                :indeterminate="dirIndeterminate(record.dirId)"
                @change="(v: any) => onDirCheck(record.dirId, toChecked(v))"
              >
                {{ record.dirName }}
              </ACheckbox>
            </template>
            <template v-if="column.dataIndex === 'menu'">
              <ACheckbox
                :checked="menuChecked(record.menuId)"
                :indeterminate="menuIndeterminate(record.menuId)"
                @change="(v: any) => onMenuCheck(record.menuId, record.dirId, toChecked(v))"
              >
                {{ record.menuName }}
              </ACheckbox>
            </template>
            <template v-if="column.dataIndex === 'buttons'">
              <div class="btn-checks">
                <ACheckbox
                  v-for="btn in record.buttons"
                  :key="btn.id"
                  :checked="isChecked(btn.id)"
                  @change="(v: any) => onButtonCheck(btn.id, record.menuId, record.dirId, toChecked(v))"
                >
                  {{ btn.name }}
                </ACheckbox>
                <span v-if="!record.buttons.length" class="muted">无按钮</span>
              </div>
            </template>
          </template>
        </ATable>
      </ASpin>
    </AModal>

    <AModal
      v-model:open="scopeOpen"
      :title="`数据范围 - ${scopeRole?.name || ''}`"
      :confirm-loading="scopeSubmitting"
      destroy-on-close
      @ok="submitGrantScope"
    >
      <AForm :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="数据范围">
          <ASelect v-model:value="scopeForm.data_scope" :options="dataScopeOptions" />
        </AFormItem>
        <AFormItem v-if="scopeForm.data_scope === '4'" label="指定部门">
          <ATreeSelect
            v-model:value="scopeForm.org_ids"
            :tree-data="orgTreeSelectData"
            tree-checkable
            tree-default-expand-all
            allow-clear
            placeholder="请选择部门"
            :field-names="{ label: 'title', value: 'value', children: 'children' }"
            style="width: 100%"
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

.btn-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
}

.muted {
  color: var(--ant-color-text-secondary, #999);
}
</style>
