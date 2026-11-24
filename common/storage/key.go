package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

const MaxKeyLen = 512

// 逻辑 key：字母数字与 . _ / -，分段非空，禁止穿越与绝对路径形态。
var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

// NormalizeKey 校验并规范化逻辑对象键。
func NormalizeKey(key string) (string, error) {
	// 1. 去空白并拒绝空键 / 超长
	key = strings.TrimSpace(key)
	if key == "" || len(key) > MaxKeyLen {
		return "", ErrInvalidKey
	}
	// 2. 拒绝绝对路径、UNC、盘符、反斜杠与控制字符
	if strings.HasPrefix(key, "/") || strings.HasPrefix(key, "\\") || strings.Contains(key, "\\") {
		return "", ErrInvalidKey
	}
	if len(key) >= 2 && unicode.IsLetter(rune(key[0])) && key[1] == ':' {
		return "", ErrInvalidKey
	}
	for _, r := range key {
		if unicode.IsControl(r) {
			return "", ErrInvalidKey
		}
	}
	// 3. 白名单与分段规则：禁 ..、空段、连续 //
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidKey
	}
	if strings.Contains(key, "//") || strings.Contains(key, "..") {
		return "", ErrInvalidKey
	}
	parts := strings.Split(key, "/")
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return "", ErrInvalidKey
		}
	}
	return key, nil
}

// ObjectHash 对逻辑 key 做 SHA-256，供 local 内容寻址布局使用。
func ObjectHash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
