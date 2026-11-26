/** 把 hex 与白色按 ratio（0~1）混合，得到浅底色。 */
export function mixWithWhite(hex: string, ratio = 0.88) {
  const raw = hex.replace('#', '').trim()
  if (raw.length !== 6) return '#e6f4ff'
  const r = Number.parseInt(raw.slice(0, 2), 16)
  const g = Number.parseInt(raw.slice(2, 4), 16)
  const b = Number.parseInt(raw.slice(4, 6), 16)
  const blend = (c: number) => Math.round(c + (255 - c) * ratio)
  const toHex = (c: number) => blend(c).toString(16).padStart(2, '0')
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}
