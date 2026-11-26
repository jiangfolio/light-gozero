<script setup lang="ts">
/**
 * 分配角色穿梭选择器：左待选（可搜可分页）/ 右已选，点加减迁移。
 */
import { computed, ref, watch } from 'vue'

import EllipsisText from '@/components/ellipsis-text/index.vue'

export type RoleAssignOption = {
  id: string
  name: string
  code?: string
  data_scope?: string
}

const props = withDefaults(
  defineProps<{
    value?: string[]
    options?: RoleAssignOption[]
  }>(),
  {
    value: () => [],
    options: () => [],
  },
)

const emit = defineEmits<{
  'update:value': [string[]]
}>()

const dataScopeLabel: Record<string, string> = {
  '1': '仅本人',
  '2': '本部门',
  '3': '本部门及以下',
  '4': '自定义部门',
  '5': '全部数据',
}

const searchDraft = ref('')
const searchName = ref('')
const leftPage = ref(1)
const leftPageSize = ref(10)
const rightPage = ref(1)
const rightPageSize = ref(10)

const selectedSet = computed(() => new Set(props.value.map(String)))

const optionMap = computed(() => {
  const map = new Map<string, RoleAssignOption>()
  for (const item of props.options) map.set(String(item.id), item)
  return map
})

/** 左侧：未选中且命中名称筛选 */
const leftAll = computed(() => {
  const q = searchName.value.trim().toLowerCase()
  return props.options.filter((item) => {
    if (selectedSet.value.has(String(item.id))) return false
    if (!q) return true
    return String(item.name || '')
      .toLowerCase()
      .includes(q)
  })
})

const leftTotal = computed(() => leftAll.value.length)

const leftRows = computed(() => {
  const start = (leftPage.value - 1) * leftPageSize.value
  return leftAll.value.slice(start, start + leftPageSize.value)
})

/** 右侧：已选角色（保持勾选顺序） */
const rightAll = computed(() => {
  const rows: RoleAssignOption[] = []
  for (const id of props.value) {
    const hit = optionMap.value.get(String(id))
    if (hit) rows.push(hit)
    else rows.push({ id: String(id), name: String(id) })
  }
  return rows
})

const rightTotal = computed(() => rightAll.value.length)

const rightRows = computed(() => {
  const start = (rightPage.value - 1) * rightPageSize.value
  return rightAll.value.slice(start, start + rightPageSize.value)
})

watch(
  () => [leftTotal.value, leftPageSize.value] as const,
  () => {
    const max = Math.max(1, Math.ceil(leftTotal.value / leftPageSize.value) || 1)
    if (leftPage.value > max) leftPage.value = max
  },
)

watch(
  () => [rightTotal.value, rightPageSize.value] as const,
  () => {
    const max = Math.max(1, Math.ceil(rightTotal.value / rightPageSize.value) || 1)
    if (rightPage.value > max) rightPage.value = max
  },
)

const leftColumns = [
  { title: '操作', dataIndex: 'action', width: 64 },
  { title: '名称', dataIndex: 'name' },
  { title: '分类', dataIndex: 'data_scope', width: 110 },
]

const rightColumns = [
  { title: '操作', dataIndex: 'action', width: 64 },
  { title: '名称', dataIndex: 'name' },
]

const leftPagination = computed(() => ({
  current: leftPage.value,
  pageSize: leftPageSize.value,
  total: leftTotal.value,
  size: 'small' as const,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50'],
  onChange: (page: number, size: number) => {
    leftPage.value = page
    leftPageSize.value = size
  },
}))

const rightPagination = computed(() => ({
  current: rightPage.value,
  pageSize: rightPageSize.value,
  total: rightTotal.value,
  size: 'small' as const,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50'],
  onChange: (page: number, size: number) => {
    rightPage.value = page
    rightPageSize.value = size
  },
}))

function emitIds(ids: string[]) {
  emit('update:value', ids)
}

function onSearch() {
  searchName.value = searchDraft.value
  leftPage.value = 1
}

function onResetSearch() {
  searchDraft.value = ''
  searchName.value = ''
  leftPage.value = 1
}

/** 单条加入已选 */
function addOne(id: string) {
  const sid = String(id)
  if (selectedSet.value.has(sid)) return
  emitIds([...props.value.map(String), sid])
}

/** 将当前页待选全部加入 */
function addCurrentPage() {
  if (!leftRows.value.length) return
  const next = [...props.value.map(String)]
  const set = new Set(next)
  for (const row of leftRows.value) {
    const id = String(row.id)
    if (!set.has(id)) {
      next.push(id)
      set.add(id)
    }
  }
  emitIds(next)
}

/** 单条移出已选 */
function removeOne(id: string) {
  const sid = String(id)
  emitIds(props.value.map(String).filter((x) => x !== sid))
}

/** 清空已选 */
function removeAll() {
  emitIds([])
}
</script>

<template>
  <div class="role-assign">
    <div class="role-assign-pane">
      <div class="role-assign-search">
        <AInput
          v-model:value="searchDraft"
          allow-clear
          placeholder="请输入角色名"
          @press-enter="onSearch"
        />
        <AButton type="primary" @click="onSearch">查询</AButton>
        <AButton @click="onResetSearch">重置</AButton>
      </div>

      <div class="role-assign-toolbar">
        <span class="role-assign-count">待选择列表 {{ leftTotal }} 条</span>
        <AButton type="link" size="small" :disabled="!leftRows.length" @click="addCurrentPage">
          添加当前数据
        </AButton>
      </div>

      <ATable
        size="small"
        bordered
        row-key="id"
        :columns="leftColumns"
        :data-source="leftRows"
        :pagination="leftPagination"
        :scroll="{ y: 280 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'action'">
            <button type="button" class="op-btn op-add" title="添加" @click="addOne(record.id)">
              <IconParkPlus :size="14" />
            </button>
          </template>
          <template v-else-if="column.dataIndex === 'name'">
            <EllipsisText :text="record.name" />
          </template>
          <template v-else-if="column.dataIndex === 'data_scope'">
            <ATag>{{ dataScopeLabel[record.data_scope] || record.data_scope || '-' }}</ATag>
          </template>
        </template>
      </ATable>
    </div>

    <div class="role-assign-pane">
      <div class="role-assign-toolbar role-assign-toolbar-solo">
        <span class="role-assign-count">已选择: {{ rightTotal }}</span>
        <AButton danger size="small" :disabled="!rightTotal" @click="removeAll">全部移除</AButton>
      </div>

      <ATable
        size="small"
        bordered
        row-key="id"
        :columns="rightColumns"
        :data-source="rightRows"
        :pagination="rightPagination"
        :scroll="{ y: 320 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'action'">
            <button
              type="button"
              class="op-btn op-remove"
              title="移除"
              @click="removeOne(record.id)"
            >
              <IconParkMinus :size="14" />
            </button>
          </template>
          <template v-else-if="column.dataIndex === 'name'">
            <EllipsisText :text="record.name" />
          </template>
        </template>
      </ATable>
    </div>
  </div>
</template>

<style scoped>
.role-assign {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  min-width: 0;
}

.role-assign-pane {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.role-assign-search {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-assign-search :deep(.ant-input-affix-wrapper),
.role-assign-search :deep(.ant-input) {
  flex: 1;
  min-width: 0;
}

.role-assign-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 32px;
}

.role-assign-toolbar-solo {
  margin-top: 0;
}

.role-assign-count {
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 13px;
}

.op-btn {
  display: inline-flex;
  width: 24px;
  height: 24px;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--ant-color-border, #d9d9d9);
  border-radius: var(--app-border-radius, 4px);
  background: transparent;
  cursor: pointer;
  padding: 0;
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  transition: color 0.15s, border-color 0.15s, background 0.15s;
}

.op-btn.op-add:hover {
  color: var(--ant-color-primary, #1677ff);
  border-color: var(--ant-color-primary, #1677ff);
  background: var(--ant-color-primary-bg, #e6f4ff);
}

.op-btn.op-remove {
  color: var(--ant-color-error, #ff4d4f);
  border-color: color-mix(in srgb, var(--ant-color-error, #ff4d4f) 45%, #fff);
}

.op-btn.op-remove:hover {
  border-color: var(--ant-color-error, #ff4d4f);
  background: var(--ant-color-error-bg, #fff2f0);
}
</style>
