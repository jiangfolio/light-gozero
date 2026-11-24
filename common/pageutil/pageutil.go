package pageutil

import (
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MaxPage         = 10000
)

// Parse 解析并校验分页参数（字符串入参）。
func Parse(pageStr, pageSizeStr string) (page, pageSize int, err error) {
	// 1. 页码、页大小必须显式传入，禁止缺省当全量
	if pageStr == "" || pageSizeStr == "" {
		return 0, 0, errorx.New(400, "页码和每页条数不能为空")
	}
	page, err = convert.IntFromString(pageStr)
	if err != nil || page < 1 {
		return 0, 0, errorx.New(400, "页码无效")
	}
	pageSize, err = convert.IntFromString(pageSizeStr)
	if err != nil || pageSize < 1 {
		return 0, 0, errorx.New(400, "每页条数无效")
	}
	// 2. 强制上限，禁止大 page_size 当 list，并限制深分页
	if pageSize > MaxPageSize {
		return 0, 0, errorx.New(400, "每页条数超出上限")
	}
	if page > MaxPage {
		return 0, 0, errorx.New(400, "页码超出上限")
	}
	return page, pageSize, nil
}

// Offset 计算 SQL OFFSET。
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}
