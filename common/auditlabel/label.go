package auditlabel

// ResultName 登录/操作结果中文（列表筛选常用，单独映射）。
func ResultName(result string) string {
	switch result {
	case "success":
		return "成功"
	case "fail":
		return "失败"
	case "locked":
		return "锁定"
	case "logout":
		return "登出"
	default:
		return result
	}
}
