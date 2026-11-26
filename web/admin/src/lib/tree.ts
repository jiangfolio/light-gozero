/**
 * 循环拉取分页直到取完或达上限（如下拉选项需要扁平全量列表时）。
 */
export async function fetchAllPages(
  fetcher: (page: number, pageSize: number) => Promise<any>,
  pageSize = 100,
  maxPages = 50,
): Promise<any[]> {
  const all: any[] = []
  let page = 1
  while (page <= maxPages) {
    const resp = await fetcher(page, pageSize)
    all.push(...(resp.items ?? []))
    const total = Number(resp.total || 0)
    if (all.length >= total || (resp.items?.length ?? 0) < pageSize) {
      break
    }
    page += 1
  }
  return all
}
