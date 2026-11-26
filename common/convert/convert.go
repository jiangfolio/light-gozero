package convert

import (
	"strconv"
	"strings"
	"time"

	"github.com/jiangbyte/light-gozero/common/errorx"
)

// Int64FromString 将入站字符串显式转为 int64。
func Int64FromString(s string) (int64, error) {
	// 1. 去掉首尾空白，空串直接判非法，避免 strconv 产生难读错误
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errorx.New(400, "数字不能为空")
	}
	// 2. 按十进制解析；失败返回可对外展示的业务错误
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errorx.New(400, "数字格式不正确")
	}
	return v, nil
}

// Int64ToString 将出站 int64 显式转为十进制字符串。
func Int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

// IntFromString 将入站字符串显式转为 int。
func IntFromString(s string) (int, error) {
	// 1. 先走 int64 解析，统一空串与非法格式处理
	v, err := Int64FromString(s)
	if err != nil {
		return 0, err
	}
	// 2. 校验是否落在平台 int 范围内，防止静默截断
	if int64(int(v)) != v {
		return 0, errorx.New(400, "数字超出范围")
	}
	return int(v), nil
}

// IntToString 将出站 int 显式转为十进制字符串。
func IntToString(v int) string {
	return strconv.Itoa(v)
}

// ParseStatus 解析可选 status；空串用默认值。
func ParseStatus(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := Int64FromString(s)
	if err != nil {
		return 0, errorx.New(400, "状态无效")
	}
	return v, nil
}

// ParseSort 解析可选 sort；空串用默认值。
func ParseSort(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := Int64FromString(s)
	if err != nil {
		return 0, errorx.New(400, "排序值无效")
	}
	return v, nil
}

// ParseFlag 解析可选 0/1 标记；空串用默认值。
func ParseFlag(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := Int64FromString(s)
	if err != nil || (v != 0 && v != 1) {
		return 0, errorx.New(400, "标志值无效")
	}
	return v, nil
}

// TimeToString 出站时间统一为「2006-01-02 15:04:05」；零值返回空串。
func TimeToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// DerefString 解引用字符串指针；nil 返回空串。
func DerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// StringPtr 空串返回 nil，便于写可空列。
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
