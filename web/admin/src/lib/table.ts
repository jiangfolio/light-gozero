/** 纯操作列 dataIndex（不含业务字段 action，如操作日志的「动作」） */
const OP_COLUMN_KEYS = new Set(['action_btn', 'operations'])

export type TableColumnLike = {
  dataIndex?: string | string[]
  key?: string | number
  title?: unknown
  fixed?: string | boolean
  ellipsis?: boolean | { showTitle?: boolean }
  width?: number | string
  [k: string]: unknown
}

function isOpColumn(col: TableColumnLike): boolean {
  if (col.fixed === 'right') return true
  const key = String(col.dataIndex ?? col.key ?? '')
  if (OP_COLUMN_KEYS.has(key)) return true
  // 约定：dataIndex=action 且标题为「操作」才是操作列
  return key === 'action' && String(col.title ?? '') === '操作'
}

/**
 * 管理端表格列规范化：文本列默认 ellipsis（带悬停 title），操作列保持不省略。
 * PageTable 与抽屉内裸 ATable 都应走此函数，避免超长撑破布局。
 */
export function withEllipsisColumns<T extends TableColumnLike>(columns: T[]): T[] {
  return columns.map((col) => {
    if (col == null || typeof col !== 'object') return col
    if (col.ellipsis !== undefined) return col
    if (isOpColumn(col)) {
      return { ...col, ellipsis: false }
    }
    return { ...col, ellipsis: { showTitle: true } }
  })
}

/** 按列 width 求和得到 scroll.x；避免 max-content 撑开导致 ellipsis 失效 */
export function tableScrollX(columns: TableColumnLike[], min = 960): number {
  const sum = columns.reduce((s, c) => {
    const w = Number(c?.width)
    return s + (Number.isFinite(w) && w > 0 ? w : 120)
  }, 0)
  return Math.max(sum, min)
}
