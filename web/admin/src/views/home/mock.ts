/**
 * 工作台占位数据：围绕本系统能力（用户/组织/权限/审计），不接真实接口。
 */

export type KpiItem = {
  key: string
  label: string
  value: string
  hint: string
  icon: string
  tone: string
}

export type FeedItem = {
  id: string
  title: string
  meta: string
  time: string
  ok?: boolean
}

export const kpis: KpiItem[] = [
  {
    key: 'user',
    label: '用户',
    value: '128',
    hint: '启用 121',
    icon: 'user',
    tone: '#1677ff',
  },
  {
    key: 'org',
    label: '组织',
    value: '16',
    hint: '含下级节点',
    icon: 'peoples',
    tone: '#13c2c2',
  },
  {
    key: 'role',
    label: '角色',
    value: '8',
    hint: '含数据范围',
    icon: 'protect',
    tone: '#52c41a',
  },
  {
    key: 'resource',
    label: '菜单资源',
    value: '42',
    hint: '目录/菜单/按钮',
    icon: 'all-application',
    tone: '#fa8c16',
  },
]

/** 近 7 日登录结果占位 */
export const loginTrend = {
  dates: ['11-20', '11-21', '11-22', '11-23', '11-24', '11-25', '11-26'],
  success: [36, 48, 52, 41, 63, 71, 58],
  fail: [4, 6, 3, 9, 5, 8, 4],
  locked: [0, 1, 0, 2, 1, 0, 1],
}

export const recentLogins: FeedItem[] = [
  {
    id: '1',
    title: '超级管理员 登录成功',
    meta: '127.0.0.1 · Chrome',
    time: '22:18',
    ok: true,
  },
  {
    id: '2',
    title: 'ops 登录失败',
    meta: '密码错误 · 已累计 2 次',
    time: '21:04',
    ok: false,
  },
  {
    id: '3',
    title: 'auditor 登录成功',
    meta: '10.0.0.18 · Edge',
    time: '20:41',
    ok: true,
  },
  {
    id: '4',
    title: 'guest 触发锁定',
    meta: '连续失败超过阈值',
    time: '19:12',
    ok: false,
  },
  {
    id: '5',
    title: '超级管理员 退出登录',
    meta: '主动登出',
    time: '18:55',
    ok: true,
  },
]

export const recentOpers: FeedItem[] = [
  {
    id: '1',
    title: '更新角色权限',
    meta: 'system:role:bind_resources',
    time: '22:02',
    ok: true,
  },
  {
    id: '2',
    title: '创建菜单资源',
    meta: 'system:resource:create',
    time: '21:36',
    ok: true,
  },
  {
    id: '3',
    title: '保存安全配置',
    meta: 'system:security_config:update',
    time: '20:15',
    ok: true,
  },
  {
    id: '4',
    title: '删除用户',
    meta: '校验未通过',
    time: '19:48',
    ok: false,
  },
  {
    id: '5',
    title: '分配用户角色',
    meta: 'system:user:bind_roles',
    time: '18:20',
    ok: true,
  },
]
