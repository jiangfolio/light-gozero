package appcfg

// 配置键与默认值（库无行或停用时回退）。
const (
	KeyTokenTTLSeconds      = "auth.token_ttl_seconds"
	KeyMaxFailAttempts      = "auth.max_fail_attempts"
	KeyLockMinutes          = "auth.lock_minutes"
	KeyPasswordMinLen             = "auth.password_min_len"
	KeyPasswordComplexityEnabled  = "auth.password_complexity_enabled"
	KeyPasswordExpireDays         = "auth.password_expire_days"
	KeyPasswordHistoryCount       = "auth.password_history_count"
	KeyCaptchaTTLSeconds          = "captcha.ttl_seconds"
	KeyCaptchaLength              = "captcha.length"
)

const (
	DefaultTokenTTLSeconds           = 86400
	DefaultMaxFailAttempts           = 5
	DefaultLockMinutes               = 30
	DefaultPasswordMinLen            = 8
	DefaultPasswordComplexityEnabled = true
	DefaultPasswordExpireDays        = 90
	DefaultPasswordHistoryCount      = 5
	DefaultCaptchaTTLSeconds         = 120
	DefaultCaptchaLength             = 4
)

// PublicKeys 前端公开接口允许返回的配置键（硬编码白名单）。
var PublicKeys = []string{
	KeyPasswordMinLen,
	KeyPasswordComplexityEnabled,
	KeyPasswordExpireDays,
	KeyCaptchaLength,
}
