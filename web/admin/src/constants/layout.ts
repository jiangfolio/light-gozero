export const LAYOUT_ENUM = {
  SIDE: 'side',
  SIDE_EXPAND: 'sideExpand',
  TOP: 'top',
  TOP_EXPAND: 'topExpand',
} as const

export type LayoutMode = (typeof LAYOUT_ENUM)[keyof typeof LAYOUT_ENUM]

export const PAGE_TAG_STYLE = {
  ANTDV: 'antdv',
  CHROME: 'chrome',
} as const

export type PageTagStyle = (typeof PAGE_TAG_STYLE)[keyof typeof PAGE_TAG_STYLE]

export const THEME_COLORS = [
  '#1677ff',
  '#2F54EB',
  '#00b96b',
  '#F5222D',
  '#13c2c2',
  '#EB2F96',
  '#722ED1',
  '#52c41a',
  '#009688',
  '#fa541c',
]

export const DEFAULT_APP_CONFIG = {
  layout: LAYOUT_ENUM.SIDE as LayoutMode,
  themeColor: '#1677ff',
  /* 设计约定默认圆角为 0（antdv 派生的 LG/SM/XS 随之全为 0）；Chrome 页签弧形除外 */
  borderRadius: 0,
  menuWidth: 210,
  menuTheme: 'dark' as 'dark' | 'light',
  compact: false,
  pageTagLocation: 'center' as 'top' | 'center',
  pageTagStyle: PAGE_TAG_STYLE.CHROME as PageTagStyle,
  menuAccordion: true,
  pageTagFlag: true,
  breadcrumbFlag: true,
  footerFlag: true,
  watermarkFlag: true,
  helpDocFlag: false,
  darkMode: false,
}
