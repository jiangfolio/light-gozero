/**
 * IconPark Outline 图标集工具：名称规范化、解析与列表，供 MenuIcon / IconPicker 复用。
 */
import parkOutline from '@iconify-json/icon-park-outline/icons.json'

export const ICON_PARK_FALLBACK = 'application-one'

type IconifyIcon = { body: string; width?: number; height?: number }
type IconifyJSON = {
  width?: number
  height?: number
  icons: Record<string, IconifyIcon>
}

const collection = parkOutline as IconifyJSON

let nameCache: string[] | null = null

/** 规范化为 kebab 图标名 */
export function normalizeIconName(raw?: string) {
  const s = (raw || '').trim()
  if (!s) return ICON_PARK_FALLBACK
  return s
    .replace(/([a-z0-9])([A-Z])/g, '$1-$2')
    .replace(/[_\s]+/g, '-')
    .replace(/-+/g, '-')
    .toLowerCase()
}

/** 解析图标 SVG 数据；未知名称回退默认图标 */
export function resolveIconPark(name?: string) {
  const key = normalizeIconName(name)
  const icon = collection.icons[key] || collection.icons[ICON_PARK_FALLBACK]
  return {
    name: collection.icons[key] ? key : ICON_PARK_FALLBACK,
    body: icon.body,
    w: icon.width || collection.width || 48,
    h: icon.height || collection.height || 48,
  }
}

/** 全量图标名（排序缓存） */
export function listIconParkNames() {
  if (!nameCache) {
    nameCache = Object.keys(collection.icons).sort()
  }
  return nameCache
}

/**
 * 按关键词筛选图标名；空关键词返回前缀片段，避免一次渲染全集。
 */
export function filterIconParkNames(keyword: string, limit = 120) {
  // 1. 取全量列表
  const all = listIconParkNames()
  // 2. 空关键词只截取前缀，控制弹层初次渲染量
  const q = keyword.trim().toLowerCase()
  if (!q) return all.slice(0, limit)
  // 3. 包含匹配并截断，避免超长结果卡顿
  const out: string[] = []
  for (const name of all) {
    if (name.includes(q)) {
      out.push(name)
      if (out.length >= limit) break
    }
  }
  return out
}
