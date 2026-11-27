/**
 * 工作台类型与空趋势占位（真实数据由 index 拉取）。
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

export type LoginTrend = {
  dates: string[]
  success: number[]
  fail: number[]
  locked: number[]
}

/** 近 7 日零值趋势，接口失败或无权限时使用 */
export function emptyLoginTrend(): LoginTrend {
  const dates: string[] = []
  for (let i = 6; i >= 0; i -= 1) {
    const d = new Date()
    d.setHours(0, 0, 0, 0)
    d.setDate(d.getDate() - i)
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    dates.push(`${mm}-${dd}`)
  }
  return {
    dates,
    success: dates.map(() => 0),
    fail: dates.map(() => 0),
    locked: dates.map(() => 0),
  }
}
