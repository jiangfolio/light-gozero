/** 树节点转 TreeSelect / Tree 选项（title/value/key/children） */
export function toTreeSelectData(nodes: any[] = [], labelKey = 'name', valueKey = 'id'): any[] {
  return nodes.map((n) => ({
    ...n,
    title: n[labelKey],
    value: n[valueKey],
    key: n[valueKey],
    children: n.children?.length ? toTreeSelectData(n.children, labelKey, valueKey) : undefined,
  }))
}