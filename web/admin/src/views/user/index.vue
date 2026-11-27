<script setup lang="ts">
import { message, Modal } from 'antdv-next'
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { fetchPublicKey } from '@/api/system/auth'
import { treeOrg } from '@/api/system/org'
import { pageRole } from '@/api/system/role'
import {
  bindUserRoles,
  createUser,
  deleteUser,
  getUserDetail,
  pageUser,
  unbindUserRoles,
  unlockUser,
  updateUser,
} from '@/api/system/user'
import EllipsisText from '@/components/ellipsis-text/index.vue'
import PageSearch from '@/components/page-search/index.vue'
import PageTable from '@/components/page-table/index.vue'
import RoleAssignSelector from '@/components/role-assign-selector/index.vue'
import { usePageReload } from '@/composables/use-page-reload'
import { encryptPassword } from '@/lib/crypto/sm2'
import { toTreeSelectData } from '@/lib/format'
import { refreshMenus } from '@/lib/refresh-menus'
import { withEllipsisColumns } from '@/lib/table'
import { fetchAllPages } from '@/lib/tree'
import { useUserStore } from '@/stores/user'

defineOptions({ name: 'SystemUser' })

const router = useRouter()
const userStore = useUserStore()

const orgTreeData = ref<any[]>([])
const orgTreeSelectData = ref<any[]>([])
const selectedOrgId = ref<string | undefined>()
const orgNameMap = ref<Record<string, string>>({})

const defaultQuery = {
  username: '',
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
  { title: '账号', dataIndex: 'username', width: 140 },
  { title: '姓名', dataIndex: 'nickname', width: 120 },
  { title: '所属组织', dataIndex: 'org_id', width: 160 },
  { title: '状态', dataIndex: 'status', width: 90 },
  { title: '锁定', dataIndex: 'locked_until', width: 180 },
  { title: '创建时间', dataIndex: 'created_at', width: 180 },
  { title: '操作', dataIndex: 'action', width: 260, fixed: 'right' as const },
])

/** 锁定截止时间未过期则视为仍锁定 */
function isLocked(record: any) {
  if (!record?.locked_until) return false
  const t = Date.parse(String(record.locked_until).replace(/-/g, '/'))
  return Number.isFinite(t) && t > Date.now()
}

function confirmUnlock(record: any) {
  Modal.confirm({
    title: '确认解锁',
    content: `确定解除账号「${record.username}」的登录锁定？`,
    async onOk() {
      try {
        await unlockUser(record.id)
        message.success('已解锁')
        await loadData()
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '解锁失败')
        throw e
      }
    },
  })
}

const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedKeys.value = keys.map(String)
  },
}))

function collectOrgNames(nodes: any[], map: Record<string, string>) {
  for (const n of nodes) {
    map[n.id] = n.name
    if (n.children?.length) collectOrgNames(n.children, map)
  }
}

async function loadOrgTree() {
  const resp = await treeOrg()
  const tree = resp.items ?? []
  orgTreeData.value = toTreeSelectData(tree)
  orgTreeSelectData.value = orgTreeData.value
  const map: Record<string, string> = {}
  collectOrgNames(tree, map)
  orgNameMap.value = map
}

async function loadData() {
  tableLoading.value = true
  try {
    const resp = await pageUser({
      page: String(pageNum.value),
      page_size: String(pageSize.value),
      username: queryForm.username || undefined,
      org_id: selectedOrgId.value || undefined,
      status: queryForm.status,
    })
    tableData.value = resp.items ?? []
    total.value = Number(resp.total || 0)
    selectedKeys.value = []
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载用户失败')
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

function onSelectOrg(keys: (string | number)[]) {
  selectedOrgId.value = keys.length ? String(keys[0]) : undefined
  pageNum.value = 1
  loadData()
}

const modalOpen = ref(false)
const modalSubmitting = ref(false)
const editingId = ref<string | null>(null)
const modalTitle = computed(() => (editingId.value ? '编辑用户' : '新建用户'))
const formRef = ref()
const formState = reactive({
  org_id: '',
  username: '',
  nickname: '',
  password: '',
  status: '1',
})

function openCreate() {
  editingId.value = null
  Object.assign(formState, {
    org_id: selectedOrgId.value || orgTreeSelectData.value[0]?.value || '',
    username: '',
    nickname: '',
    password: '',
    status: '1',
  })
  modalOpen.value = true
}

function openEdit(record: any) {
  editingId.value = record.id
  Object.assign(formState, {
    org_id: record.org_id,
    username: record.username,
    nickname: record.nickname,
    password: '',
    status: record.status || '1',
  })
  modalOpen.value = true
}

async function encryptIfNeeded(plain: string) {
  if (!plain) return ''
  const { public_key } = await fetchPublicKey()
  return encryptPassword(plain, public_key)
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
      const password = formState.password
        ? await encryptIfNeeded(formState.password)
        : undefined
      await updateUser({
        id: editingId.value,
        org_id: formState.org_id,
        nickname: formState.nickname,
        password: password || undefined,
        status: formState.status,
      })
      message.success('更新成功')
    } else {
      const password = await encryptIfNeeded(formState.password)
      await createUser({
        org_id: formState.org_id,
        username: formState.username,
        password,
        nickname: formState.nickname || undefined,
        status: formState.status,
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
    message.warning('请先选择用户')
    return
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定删除选中的 ${ids.length} 个用户？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteUser(ids)
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

const roleModalOpen = ref(false)
const roleSubmitting = ref(false)
const roleUserId = ref('')
const roleOptions = ref<any[]>([])
const checkedRoleIds = ref<string[]>([])
const originRoleIds = ref<string[]>([])

async function openBindRoles(record: any) {
  roleUserId.value = record.id
  try {
    const [roles, detail] = await Promise.all([
      fetchAllPages((page, pageSize) =>
        pageRole({ page: String(page), page_size: String(pageSize) }),
      ),
      getUserDetail(record.id),
    ])
    roleOptions.value = roles
    originRoleIds.value = detail.role_ids ?? []
    checkedRoleIds.value = [...originRoleIds.value]
    roleModalOpen.value = true
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载角色失败')
  }
}

async function submitBindRoles() {
  roleSubmitting.value = true
  try {
    const selected = checkedRoleIds.value
    const origin = originRoleIds.value
    const toAdd = selected.filter((id) => !origin.includes(id))
    const toRemove = origin.filter((id) => !selected.includes(id))
    if (toAdd.length) await bindUserRoles(roleUserId.value, toAdd)
    if (toRemove.length) await unbindUserRoles(roleUserId.value, toRemove)
    message.success('角色分配成功')
    roleModalOpen.value = false
    // 改的是当前登录用户时，立刻刷新菜单/权限
    if (String(roleUserId.value) === String(userStore.userInfo?.id || '')) {
      await refreshMenus(router)
    }
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '分配失败')
  } finally {
    roleSubmitting.value = false
  }
}

async function bootstrap() {
  try {
    await loadOrgTree()
  } catch {
    // 组织树失败不阻断列表
  }
  await loadData()
}

usePageReload(bootstrap)
</script>

<template>
  <div class="page-split">
    <ACard size="small" variant="borderless" class="org-pane">
      <div class="org-pane-title">组织</div>
      <ATree
        :tree-data="orgTreeData"
        :selected-keys="selectedOrgId ? [selectedOrgId] : []"
        default-expand-all
        :field-names="{ title: 'title', key: 'key', children: 'children' }"
        @select="onSelectOrg"
      />
    </ACard>

    <div class="list-pane">
      <PageSearch @search="onSearch" @reset="onReset">
        <AFormItem label="账号">
          <AInput
            v-model:value="queryForm.username"
            placeholder="请输入账号"
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
          <AButton v-privilege="'system:user:create'" type="primary" @click="openCreate">
            <template #icon>
              <IconParkPlus />
            </template>
            新建
          </AButton>
          <AButton
            v-privilege="'system:user:delete'"
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
          <template v-if="column.dataIndex === 'org_id'">
            <EllipsisText :text="orgNameMap[record.org_id] || record.org_id" />
          </template>
          <template v-else-if="column.dataIndex === 'status'">
            <ATag :color="record.status === '1' ? 'green' : 'red'">
              {{ record.status === '1' ? '启用' : '禁用' }}
            </ATag>
          </template>
          <template v-else-if="column.dataIndex === 'locked_until'">
            <template v-if="isLocked(record)">
              <ATooltip :title="`失败 ${record.fail_count || 0} 次，锁定至 ${record.locked_until}`">
                <ATag color="orange">锁定中</ATag>
              </ATooltip>
            </template>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.dataIndex === 'action'">
            <AButton v-privilege="'system:user:update'" type="link" size="small" @click="openEdit(record)">
              编辑
            </AButton>
            <AButton
              v-if="isLocked(record)"
              v-privilege="'system:user:update'"
              type="link"
              size="small"
              @click="confirmUnlock(record)"
            >
              解锁
            </AButton>
            <AButton
              v-privilege="'system:user:bind_roles'"
              type="link"
              size="small"
              @click="openBindRoles(record)"
            >
              分配角色
            </AButton>
            <AButton
              v-privilege="'system:user:delete'"
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
    </div>

    <AModal
      v-model:open="modalOpen"
      :title="modalTitle"
      :confirm-loading="modalSubmitting"
      destroy-on-close
      @ok="submitForm"
    >
      <AForm ref="formRef" :model="formState" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <AFormItem label="所属组织" name="org_id" :rules="[{ required: true, message: '请选择组织' }]">
          <ATreeSelect
            v-model:value="formState.org_id"
            :tree-data="orgTreeSelectData"
            tree-default-expand-all
            allow-clear
            placeholder="请选择组织"
            :field-names="{ label: 'title', value: 'value', children: 'children' }"
          />
        </AFormItem>
        <AFormItem
          label="账号"
          name="username"
          :rules="editingId ? [] : [{ required: true, message: '请输入账号' }]"
        >
          <AInput v-model:value="formState.username" :disabled="!!editingId" />
        </AFormItem>
        <AFormItem label="姓名" name="nickname">
          <AInput v-model:value="formState.nickname" />
        </AFormItem>
        <AFormItem
          label="密码"
          name="password"
          :rules="editingId ? [] : [{ required: true, message: '请输入密码' }]"
        >
          <AInputPassword
            v-model:value="formState.password"
            :placeholder="editingId ? '不填则不修改' : '请输入密码'"
          />
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
      v-model:open="roleModalOpen"
      title="分配角色"
      :width="960"
      :confirm-loading="roleSubmitting"
      destroy-on-close
      ok-text="确定"
      cancel-text="取消"
      @ok="submitBindRoles"
    >
      <RoleAssignSelector v-model:value="checkedRoleIds" :options="roleOptions" />
    </AModal>
  </div>
</template>

<style scoped>
.page-split {
  display: flex;
  gap: 10px;
  height: 100%;
  min-height: 0;
  padding-bottom: 10px;
}

.org-pane {
  flex: 0 0 240px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: auto;
}

.org-pane-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.list-pane {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  gap: 10px;
}
</style>
