package passwd

import (
	"regexp"
	"unicode"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"golang.org/x/crypto/bcrypt"
)

var specialRe = regexp.MustCompile(`[^a-zA-Z0-9]`)

// Policy 密码策略参数。
type Policy struct {
	MinLen       int
	HistoryCount int
	ExpireDays   int
}

// ValidateComplexity 校验密码：始终校验最小长度；complexityEnabled 时再要求大小写+数字+特殊字符。
func ValidateComplexity(password string, minLen int, complexityEnabled bool) error {
	if minLen <= 0 {
		minLen = 8
	}
	if len(password) < minLen {
		return errorx.New(400, "密码长度不足")
	}
	if !complexityEnabled {
		return nil
	}
	var upper, lower, digit, special bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case specialRe.MatchString(string(r)):
			special = true
		}
	}
	if !upper || !lower || !digit || !special {
		return errorx.New(400, "密码须包含大小写字母、数字和特殊字符")
	}
	return nil
}

// Hash 生成 bcrypt 哈希。
func Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Match 比对明文与哈希。
func Match(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// InHistory 新密码是否命中历史哈希。
func InHistory(password string, historyHashes []string) bool {
	for _, h := range historyHashes {
		if Match(h, password) {
			return true
		}
	}
	return false
}
