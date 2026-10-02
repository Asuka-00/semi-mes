package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

var zhCN = map[string]string{
	"error.auth.invalid_credentials":    "用户名或密码错误",
	"error.auth.user_disabled":          "用户已禁用",
	"error.auth.unauthorized":           "未登录或登录已失效",
	"error.auth.forbidden":              "没有权限执行该操作",
	"error.validation.invalid_params":   "请求参数不合法",
	"error.resource.not_found":          "资源不存在",
	"error.server.internal":             "服务器内部错误",
	"error.user.username_exists":        "用户名已存在",
	"error.role.code_exists":            "角色编码已存在",
	"error.menu.has_children":           "请先删除子菜单",
	"error.route.not_found":             "工艺路线版本不存在",
	"error.route.not_draft":             "已发布或作废的版本不能修改",
	"error.route.no_path":               "当前节点没有可用的下一步",
}

var enUS = map[string]string{
	"error.auth.invalid_credentials":    "Invalid username or password",
	"error.auth.user_disabled":          "User is disabled",
	"error.auth.unauthorized":           "Unauthorized",
	"error.auth.forbidden":              "Insufficient permission",
	"error.validation.invalid_params":   "Invalid request parameters",
	"error.resource.not_found":          "Resource not found",
	"error.server.internal":             "Internal server error",
	"error.user.username_exists":        "Username already exists",
	"error.role.code_exists":            "Role code already exists",
	"error.menu.has_children":           "Delete child menus first",
	"error.route.not_found":             "Route version not found",
	"error.route.not_draft":             "Released or obsolete versions cannot be edited",
	"error.route.no_path":               "Current node has no next step",
}

// T returns a message for the request language. Default is zh-CN.
func T(c *gin.Context, key string) string {
	lang := ""
	if c != nil {
		lang = c.GetHeader("Accept-Language")
	}
	table := zhCN
	if strings.Contains(strings.ToLower(lang), "en") {
		table = enUS
	}
	if msg, ok := table[key]; ok {
		return msg
	}
	return key
}
