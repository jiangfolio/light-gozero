<script setup lang="ts">
/**
 * 工作台：KPI + 登录趋势与审计动态（栅格等高）+ 个人快捷入口（只存资源 ID）。
 */
import { message, Modal } from 'antdv-next'
import dayjs from 'dayjs'
import { LineChart } from 'echarts/charts'
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import VChart from 'vue-echarts'

import { pageAuditLogin, pageAuditOper } from '@/api/system/audit'
import { treeOrg } from '@/api/system/org'
import { pageRole } from '@/api/system/role'
import { treeResource } from '@/api/system/resource'
import {
  createShortcut,
  deleteShortcut,
  listShortcut,
  updateShortcut,
} from '@/api/system/shortcut'
import { pageUser } from '@/api/system/user'
import { usePageReload } from '@/composables/use-page-reload'
import MenuIcon from '@/layout/components/menu-icon.vue'
import { usePermissionStore } from '@/stores/permission'
import { useUserStore } from '@/stores/user'

import { buildLoginTrendOption } from './charts'
import {
  emptyLoginTrend,
  type FeedItem,
  type KpiItem,
  type LoginTrend,
} from './mock'

defineOptions({ name: 'Home' })

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

type ShortcutRow = {
  id: string
  resource_id: string
  title: string
  path: string
  icon: string
  sort: string
}

const router = useRouter()
const userStore = useUserStore()
const permissionStore = usePermissionStore()
const now = ref(dayjs())
let timer: ReturnType<typeof setInterval> | undefined

const displayName = computed(
  () => userStore.userInfo?.nickname || userStore.userInfo?.username || '管理员',
)

const greeting = computed(() => {
  const h = now.value.hour()
  if (h < 6) return '凌晨好'
  if (h < 12) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const dateText = computed(() => now.value.format('YYYY-MM-DD HH:mm'))

const kpis = ref<KpiItem[]>([
  { key: 'user', label: '用户', value: '-', hint: '加载中', icon: 'user', tone: '#1677ff' },
  { key: 'org', label: '组织', value: '-', hint: '加载中', icon: 'peoples', tone: '#13c2c2' },
  { key: 'role', label: '角色', value: '-', hint: '加载中', icon: 'protect', tone: '#52c41a' },
  {
    key: 'resource',
    label: '菜单资源',
    value: '-',
    hint: '加载中',
    icon: 'all-application',
    tone: '#fa8c16',
  },
])
const recentLogins = ref<FeedItem[]>([])
const recentOpers = ref<FeedItem[]>([])
const loginTrend = ref<LoginTrend>(emptyLoginTrend())
const loginTrendOption = computed(() => buildLoginTrendOption(loginTrend.value))
const overviewLoading = ref(false)

/** 菜单资源树：目录不可选，仅菜单可选 */
const menuSelectTree = computed(() => buildMenuSelectTree(permissionStore.menuTree || []))

const shortcuts = ref<ShortcutRow[]>([])
const shortcutLoading = ref(false)
/** 单用户快捷入口上限，与后端 MaxPerUser 保持一致 */
const SHORTCUT_MAX = 8
const shortcutFull = computed(() => shortcuts.value.length >= SHORTCUT_MAX)
const formOpen = ref(false)
const formSubmitting = ref(false)
const editingId = ref<string | null>(null)
const formRef = ref()
const formState = reactive({
  resource_id: undefined as string | undefined,
  sort: '0',
})

const formTitle = computed(() => (editingId.value ? '编辑快捷入口' : '添加快捷入口'))

function buildMenuSelectTree(nodes: any[]): any[] {
  const out: any[] = []
  for (const n of nodes) {
    const children = n.children?.length ? buildMenuSelectTree(n.children) : []
    if (n.type === 'directory') {
      out.push({
        title: n.name,
        value: `dir-${n.id}`,
        key: `dir-${n.id}`,
        disabled: true,
        selectable: false,
        children,
      })
      continue
    }
    if (n.type === 'menu') {
      out.push({
        title: n.name,
        value: String(n.id),
        key: String(n.id),
        icon: n.icon,
        children: children.length ? children : undefined,
      })
    }
  }
  return out
}

/** 递归统计树节点数 */
function countTreeNodes(nodes: any[] | undefined): number {
  if (!nodes?.length) return 0
  let n = 0
  for (const node of nodes) {
    n += 1
    n += countTreeNodes(node.children)
  }
  return n
}

/** 取组织树根节点名称，作 KPI 副文案 */
function rootOrgHint(nodes: any[] | undefined): string {
  const root = nodes?.[0]
  return root?.name ? String(root.name) : '含下级节点'
}

/** 登录日志聚合成近 7 日趋势 */
function aggregateLoginTrend(items: any[]): LoginTrend {
  const base = emptyLoginTrend()
  const index = new Map(base.dates.map((d, i) => [d, i]))
  const start = dayjs().subtract(6, 'day').startOf('day')
  for (const row of items) {
    const t = dayjs(row.created_at)
    if (!t.isValid() || t.isBefore(start)) continue
    const key = t.format('MM-DD')
    const i = index.get(key)
    if (i === undefined) continue
    if (row.result === 'success') base.success[i] += 1
    else if (row.result === 'fail') base.fail[i] += 1
    else if (row.result === 'locked') base.locked[i] += 1
  }
  return base
}

const loginResultLabel: Record<string, string> = {
  success: '登录成功',
  fail: '登录失败',
  locked: '触发锁定',
  logout: '退出登录',
}

/** 拉取 KPI / 趋势 / 最近动态（单项失败不阻断其余） */
async function loadOverview() {
  overviewLoading.value = true
  try {
    // 1. 并行请求用户/启用用户/组织树/角色/资源树/审计
    const settled = await Promise.allSettled([
      pageUser({ page: '1', page_size: '1' }),
      pageUser({ page: '1', page_size: '1', status: '1' }),
      treeOrg(),
      pageRole({ page: '1', page_size: '1' }),
      treeResource(),
      pageAuditLogin({ page: '1', page_size: '200' }),
      pageAuditLogin({ page: '1', page_size: '5' }),
      pageAuditOper({ page: '1', page_size: '5' }),
    ])

    const val = <T>(i: number): T | null =>
      settled[i]?.status === 'fulfilled' ? ((settled[i] as PromiseFulfilledResult<T>).value as T) : null

    const userPage = val<any>(0)
    const enabledPage = val<any>(1)
    const orgTree = val<any>(2)
    const rolePage = val<any>(3)
    const resTree = val<any>(4)
    const loginTrendPage = val<any>(5)
    const loginFeedPage = val<any>(6)
    const operFeedPage = val<any>(7)

    // 2. 回填 KPI
    const orgItems = orgTree?.items ?? []
    const resItems = resTree?.items ?? []
    kpis.value = [
      {
        key: 'user',
        label: '用户',
        value: String(userPage?.total ?? '-'),
        hint: enabledPage ? `启用 ${enabledPage.total}` : '启用 -',
        icon: 'user',
        tone: '#1677ff',
      },
      {
        key: 'org',
        label: '组织',
        value: String(orgTree ? countTreeNodes(orgItems) : '-'),
        hint: orgTree ? rootOrgHint(orgItems) : '加载失败',
        icon: 'peoples',
        tone: '#13c2c2',
      },
      {
        key: 'role',
        label: '角色',
        value: String(rolePage?.total ?? '-'),
        hint: '含数据范围',
        icon: 'protect',
        tone: '#52c41a',
      },
      {
        key: 'resource',
        label: '菜单资源',
        value: String(resTree ? countTreeNodes(resItems) : '-'),
        hint: '目录/菜单/按钮',
        icon: 'all-application',
        tone: '#fa8c16',
      },
    ]

    // 3. 趋势与动态（无审计权限时保持空/零）
    loginTrend.value = loginTrendPage
      ? aggregateLoginTrend(loginTrendPage.items ?? [])
      : emptyLoginTrend()

    recentLogins.value = (loginFeedPage?.items ?? []).map((row: any) => ({
      id: String(row.id),
      title: `${row.username || '-'} ${loginResultLabel[row.result] || row.result_name || row.result || ''}`,
      meta: [row.ip, row.message].filter(Boolean).join(' · ') || '-',
      time: row.created_at ? dayjs(row.created_at).format('HH:mm') : '',
      ok: row.result === 'success' || row.result === 'logout',
    }))

    recentOpers.value = (operFeedPage?.items ?? []).map((row: any) => ({
      id: String(row.id),
      title: [row.module, row.action].filter(Boolean).join(' / ') || row.path || '操作',
      meta: row.message || row.path || '-',
      time: row.created_at ? dayjs(row.created_at).format('HH:mm') : '',
      ok: row.result === 'success',
    }))
  } finally {
    overviewLoading.value = false
  }
}

async function loadShortcuts() {
  shortcutLoading.value = true
  try {
    const resp = await listShortcut()
    shortcuts.value = (resp.items ?? []).map((x: any) => ({
      id: String(x.id),
      resource_id: String(x.resource_id || ''),
      title: x.title || '',
      path: x.path || '',
      icon: x.icon || '',
      sort: String(x.sort ?? '0'),
    }))
  } catch (e: unknown) {
    const err = e as { message?: string }
    message.error(err.message || '加载快捷入口失败')
  } finally {
    shortcutLoading.value = false
  }
}

/** 工作台整页刷新：快捷入口 + 看板 */
async function reloadHome() {
  await Promise.all([loadShortcuts(), loadOverview()])
}

function resetForm() {
  Object.assign(formState, { resource_id: undefined, sort: '0' })
}

function openCreate() {
  // 前端先挡一层上限，后端仍会兜底校验
  if (shortcutFull.value) {
    message.warning(`最多添加 ${SHORTCUT_MAX} 个快捷入口`)
    return
  }
  editingId.value = null
  resetForm()
  formOpen.value = true
}

function openEdit(row: ShortcutRow) {
  editingId.value = row.id
  Object.assign(formState, {
    resource_id: row.resource_id || undefined,
    sort: row.sort || '0',
  })
  formOpen.value = true
}

async function submitForm() {
  await formRef.value?.validate()
  formSubmitting.value = true
  try {
    const payload = {
      resource_id: formState.resource_id,
      sort: formState.sort || '0',
    }
    if (editingId.value) {
      await updateShortcut({ id: editingId.value, ...payload })
      message.success('更新成功')
    } else {
      await createShortcut(payload)
      message.success('添加成功')
    }
    formOpen.value = false
    await loadShortcuts()
  } catch (e: unknown) {
    const err = e as { message?: string }
    if (err.message) message.error(err.message)
    throw e
  } finally {
    formSubmitting.value = false
  }
}

function confirmDelete(row: ShortcutRow) {
  Modal.confirm({
    title: '删除快捷入口',
    content: `确定删除「${row.title || row.path}」？`,
    okType: 'danger',
    async onOk() {
      try {
        await deleteShortcut([row.id])
        message.success('已删除')
        await loadShortcuts()
      } catch (e: unknown) {
        const err = e as { message?: string }
        message.error(err.message || '删除失败')
        throw e
      }
    },
  })
}

function go(path: string) {
  if (!path) {
    message.warning('资源路径无效')
    return
  }
  router.push(path).catch(() => {
    message.warning('路由不存在或暂无权限')
  })
}

usePageReload(async () => {
  await reloadHome()
  if (!timer) {
    timer = setInterval(() => {
      now.value = dayjs()
    }, 30_000)
  }
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="home-page">
    <section class="hero">
      <div class="hero-main">
        <div class="hero-title">{{ greeting }}，{{ displayName }}</div>
        <div class="hero-desc">
          快捷入口只保存菜单资源 ID，名称与路径始终跟资源同步，避免旧数据。
        </div>
      </div>
      <div class="hero-side">{{ dateText }}</div>
    </section>

    <!-- 快捷入口占左 2 格；4 个统计以 2×2 占右 2 格 -->
    <section class="quick-row">
      <ACard variant="borderless" class="panel shortcut-panel" :bordered="false">
        <template #title>快捷入口</template>
        <template #extra>
          <ATooltip title="刷新" :mouse-enter-delay="0.3">
            <AButton type="text" size="small" :loading="shortcutLoading" @click="loadShortcuts">
              <template #icon>
                <IconParkRefresh />
              </template>
            </AButton>
          </ATooltip>
          <ATooltip :title="shortcutFull ? `已达上限（${SHORTCUT_MAX}个）` : '添加'" :mouse-enter-delay="0.3">
            <AButton type="text" size="small" :disabled="shortcutFull" @click="openCreate">
              <template #icon>
                <IconParkPlus />
              </template>
            </AButton>
          </ATooltip>
        </template>

        <ASpin :spinning="shortcutLoading">
          <div v-if="shortcuts.length" class="shortcut-grid">
            <div
              v-for="item in shortcuts"
              :key="item.id"
              class="shortcut"
              @click="go(item.path)"
            >
              <MenuIcon :name="item.icon || 'application-one'" :size="22" class="shortcut-icon" />
              <span class="shortcut-title">{{ item.title || '资源已删除' }}</span>
              <div class="shortcut-ops" @click.stop>
                <AButton type="text" size="small" @click="openEdit(item)">
                  <template #icon>
                    <IconParkEdit />
                  </template>
                </AButton>
                <AButton type="text" size="small" danger @click="confirmDelete(item)">
                  <template #icon>
                    <IconParkDelete />
                  </template>
                </AButton>
              </div>
            </div>
            <button v-if="!shortcutFull" type="button" class="shortcut-add" @click="openCreate">
              <IconParkPlus :size="20" />
              <span>添加</span>
            </button>
          </div>
          <AEmpty
            v-else
            description="暂无快捷入口，点击右上角添加"
            :image-style="{ height: '64px' }"
          />
        </ASpin>
      </ACard>

      <section class="kpi-row" :class="{ 'is-loading': overviewLoading }">
        <div v-for="item in kpis" :key="item.key" class="kpi-card">
          <div class="kpi-icon" :style="{ color: item.tone, background: `${item.tone}14` }">
            <MenuIcon :name="item.icon" :size="22" />
          </div>
          <div class="kpi-body">
            <div class="kpi-label">{{ item.label }}</div>
            <div class="kpi-value">{{ item.value }}</div>
            <div class="kpi-hint">{{ item.hint }}</div>
          </div>
        </div>
      </section>
    </section>

    <!-- CSS Grid：minmax(0,fr) 避免撑出横向滚动；三模块一行、等高 -->
    <section class="main-grid">
      <ACard variant="borderless" class="panel panel-fill" :bordered="false" title="近 7 日登录趋势">
        <VChart class="chart" :option="loginTrendOption" autoresize />
      </ACard>
      <ACard variant="borderless" class="panel panel-fill" :bordered="false">
        <template #title>最近登录</template>
        <template #extra>
          <AButton type="link" size="small" @click="go('/security/login-log')">查看</AButton>
        </template>
        <ul v-if="recentLogins.length" class="feed">
          <li v-for="item in recentLogins" :key="item.id" class="feed-item">
            <span class="feed-dot" :class="item.ok ? 'ok' : 'bad'" />
            <div class="feed-body">
              <div class="feed-title">{{ item.title }}</div>
              <div class="feed-meta">{{ item.meta }}</div>
            </div>
            <span class="feed-time">{{ item.time }}</span>
          </li>
        </ul>
        <AEmpty v-else description="暂无登录记录" :image-style="{ height: '48px' }" />
      </ACard>
      <ACard variant="borderless" class="panel panel-fill" :bordered="false">
        <template #title>最近操作</template>
        <template #extra>
          <AButton type="link" size="small" @click="go('/security/oper-log')">查看</AButton>
        </template>
        <ul v-if="recentOpers.length" class="feed">
          <li v-for="item in recentOpers" :key="item.id" class="feed-item">
            <span class="feed-dot" :class="item.ok ? 'ok' : 'bad'" />
            <div class="feed-body">
              <div class="feed-title">{{ item.title }}</div>
              <div class="feed-meta">{{ item.meta }}</div>
            </div>
            <span class="feed-time">{{ item.time }}</span>
          </li>
        </ul>
        <AEmpty v-else description="暂无操作记录" :image-style="{ height: '48px' }" />
      </ACard>
    </section>

    <AModal
      v-model:open="formOpen"
      :title="formTitle"
      :width="480"
      :confirm-loading="formSubmitting"
      destroy-on-close
      ok-text="确定"
      cancel-text="取消"
      @ok="submitForm"
    >
      <AForm ref="formRef" :model="formState" layout="vertical">
        <AFormItem
          label="菜单资源"
          name="resource_id"
          :rules="[{ required: true, message: '请选择菜单' }]"
        >
          <ATreeSelect
            v-model:value="formState.resource_id"
            :tree-data="menuSelectTree"
            tree-default-expand-all
            show-search
            tree-node-filter-prop="title"
            placeholder="请选择菜单"
            :field-names="{ label: 'title', value: 'value', children: 'children' }"
          />
        </AFormItem>
        <AFormItem label="排序" name="sort">
          <AInput v-model:value="formState.sort" placeholder="数字越小越靠前" />
        </AFormItem>
      </AForm>
    </AModal>
  </div>
</template>

<style scoped>
.home-page {
  box-sizing: border-box;
  display: flex;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  height: 100%;
  flex-direction: column;
  gap: 12px;
  overflow-x: hidden;
  overflow-y: auto;
  padding-bottom: 12px;
}

.hero {
  display: flex;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-bg-container, #fff);
  box-shadow: var(--app-card-shadow);
}

.hero-main {
  min-width: 0;
  flex: 1;
}

.hero-title {
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 20px;
  font-weight: 600;
  line-height: 1.35;
}

.hero-desc {
  max-width: 640px;
  margin-top: 6px;
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  font-size: 13px;
  line-height: 1.6;
}

.hero-side {
  flex-shrink: 0;
  color: var(--ant-color-text-tertiary, rgba(0, 0, 0, 0.35));
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}

/* 快捷入口 + 统计一行：4 格网格，左 2 格给快捷入口，右 2 格给 2×2 统计 */
.quick-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  align-items: stretch;
}

.shortcut-panel {
  grid-column: span 2;
}

.kpi-row {
  display: grid;
  grid-column: span 2;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

@media (max-width: 1200px) {
  .quick-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .shortcut-panel,
  .kpi-row {
    grid-column: span 2;
  }
}

@media (max-width: 576px) {
  .kpi-row {
    grid-template-columns: 1fr;
  }

  .hero {
    flex-direction: column;
  }
}

.kpi-card {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-bg-container, #fff);
  box-shadow: var(--app-card-shadow);
}

.kpi-body {
  min-width: 0;
}

.kpi-icon {
  display: inline-flex;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
}

.kpi-label {
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  font-size: 13px;
}

.kpi-value {
  margin-top: 2px;
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 24px;
  font-weight: 650;
  line-height: 1.15;
}

.kpi-hint {
  margin-top: 2px;
  color: var(--ant-color-text-tertiary, rgba(0, 0, 0, 0.35));
  font-size: 12px;
}

.main-grid {
  display: grid;
  /* 图表 + 两个动态列表一行排列；minmax(0,fr) 防止撑出横向滚动 */
  grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  align-items: stretch;
}

@media (max-width: 1200px) {
  .main-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.panel {
  border-radius: var(--app-border-radius, 6px);
}

.panel-fill {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.panel-fill :deep(.ant-card-body) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.panel :deep(.ant-card-head) {
  min-height: 48px;
  padding: 0 16px;
}

.panel :deep(.ant-card-body) {
  padding: 12px 16px 16px;
}

.chart {
  width: 100%;
  flex: 1;
  min-height: 300px;
}

.feed {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  margin: 0;
  padding: 0;
  list-style: none;
  /* 与图表卡等高后，条目超出仅在列表内滚动 */
  overflow-x: hidden;
  overflow-y: auto;
}

.feed-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.feed-dot {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
  margin-top: 6px;
  border-radius: 50%;
}

.feed-dot.ok {
  background: #52c41a;
}

.feed-dot.bad {
  background: #ff4d4f;
}

.feed-body {
  flex: 1;
  min-width: 0;
}

.feed-title {
  overflow: hidden;
  color: var(--ant-color-text, rgba(0, 0, 0, 0.88));
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-meta {
  margin-top: 2px;
  overflow: hidden;
  color: var(--ant-color-text-secondary, rgba(0, 0, 0, 0.45));
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-time {
  flex-shrink: 0;
  color: var(--ant-color-text-tertiary, rgba(0, 0, 0, 0.35));
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.shortcut-grid {
  display: grid;
  /* 紧凑竖向小块，一行四个 */
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  width: 100%;
  min-width: 0;
}

@media (max-width: 576px) {
  .shortcut-grid {
    grid-template-columns: 1fr;
  }
}

/* 参考 AccountBook：图标在上标题在下的紧凑小块，悬浮高亮并浮现操作 */
.shortcut {
  position: relative;
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 4px;
  cursor: pointer;
  border: 1px solid var(--ant-color-split, #f0f0f0);
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-fill-alter, #fafafa);
  transition:
    background-color 0.2s,
    border-color 0.2s,
    color 0.2s;

  &:hover {
    border-color: var(--ant-color-primary-border, #91caff);
    background: var(--ant-color-primary-bg, #e6f4ff);
    color: var(--ant-color-primary, #1677ff);

    .shortcut-ops {
      opacity: 1;
    }
  }
}

.shortcut-icon {
  line-height: 1;
}

.shortcut-title {
  box-sizing: border-box;
  max-width: 100%;
  overflow: hidden;
  padding: 0 4px;
  font-size: 13px;
  line-height: 1.2;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shortcut-ops {
  position: absolute;
  top: 2px;
  right: 2px;
  display: flex;
  opacity: 0;
  transition: opacity 0.2s;

  :deep(.ant-btn) {
    width: 20px;
    height: 20px;
    padding: 0;
    font-size: 12px;
  }
}

.shortcut-add {
  display: flex;
  min-width: 0;
  min-height: 72px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
  border: 1px dashed var(--ant-color-border, #d9d9d9);
  border-radius: var(--app-border-radius, 6px);
  background: var(--ant-color-bg-container, #fff);
  color: var(--ant-color-text-tertiary, #8c8c8c);
  font-size: 13px;

  &:hover {
    border-color: var(--ant-color-primary, #1677ff);
    color: var(--ant-color-primary, #1677ff);
  }
}
</style>
