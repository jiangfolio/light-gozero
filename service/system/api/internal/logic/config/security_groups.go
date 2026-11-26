package config

// SecurityConfigGroups 网络安全配置专用分组（登录/密码/验证码等）。
var SecurityConfigGroups = []string{"auth", "captcha"}

func isSecurityConfigGroup(group string) bool {
	for _, g := range SecurityConfigGroups {
		if g == group {
			return true
		}
	}
	return false
}
