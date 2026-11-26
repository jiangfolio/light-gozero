import { http } from '@/lib/http'

export function createDictType(data: any) {
  return http.post('/api/system/v1/dict/type/create', data) as Promise<any>
}

export function updateDictType(data: any) {
  return http.post('/api/system/v1/dict/type/update', data) as Promise<any>
}

export function deleteDictType(ids: any) {
  return http.post('/api/system/v1/dict/type/delete', { ids }) as Promise<any>
}

export function getDictTypeDetail(id: any) {
  return http.get('/api/system/v1/dict/type/detail', { params: { id } }) as Promise<any>
}

export function pageDictType(params: any) {
  return http.get('/api/system/v1/dict/type/page', { params }) as Promise<any>
}

export function createDictItem(data: any) {
  return http.post('/api/system/v1/dict/item/create', data) as Promise<any>
}

export function updateDictItem(data: any) {
  return http.post('/api/system/v1/dict/item/update', data) as Promise<any>
}

export function deleteDictItem(ids: any) {
  return http.post('/api/system/v1/dict/item/delete', { ids }) as Promise<any>
}

export function getDictItemDetail(id: any) {
  return http.get('/api/system/v1/dict/item/detail', { params: { id } }) as Promise<any>
}

export function pageDictItem(params: any) {
  return http.get('/api/system/v1/dict/item/page', { params }) as Promise<any>
}

export function getDictByCode(dictCode: any) {
  return http.get('/api/system/v1/dict/by_code', {
    params: { dict_code: dictCode },
  }) as Promise<any>
}
