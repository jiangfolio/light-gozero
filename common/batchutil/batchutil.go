package batchutil

import "github.com/jiangbyte/light-gozero/common/errorx"

// IN 查询单批上限。
const InSize = 500

// 批量写入单批上限。
const WriteSize = 1000

// 单次请求 ids 类入参上限，防止超大批次 DoS。
const MaxRequestIDs = 100

// CapRequestIDs 校验去重后的 ID 列表长度。
func CapRequestIDs(ids []string) ([]string, error) {
	ids = UniqueStrings(ids)
	if len(ids) > MaxRequestIDs {
		return nil, errorx.New(400, "单次最多处理 100 条")
	}
	return ids, nil
}

// Chunk 按 size 切分切片；空输入返回 nil。
func Chunk[T any](items []T, size int) [][]T {
	if len(items) == 0 {
		return nil
	}
	if size <= 0 {
		size = InSize
	}
	out := make([][]T, 0, (len(items)+size-1)/size)
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		out = append(out, items[i:end])
	}
	return out
}

// UniqueStrings 去重并保持首次出现顺序；空串跳过。
func UniqueStrings(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
