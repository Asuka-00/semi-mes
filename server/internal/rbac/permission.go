package rbac

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/jwt"

	"semi-mes/server/internal/database"
	"semi-mes/server/internal/i18n"
)

type resource struct {
	prefix string
	perm   string
}

var resources = []resource{
	{"/api/v1/baseFactory", "base:factory"},
	{"/api/v1/baseWorkshop", "base:workshop"},
	{"/api/v1/baseProductionLine", "base:line"},
	{"/api/v1/baseProduct", "base:product"},
	{"/api/v1/baseProcessRoute", "base:route"},
	{"/api/v1/baseOperation", "base:operation"},
	{"/api/v1/baseRecipe", "base:recipe"},
	{"/api/v1/sysUserRole", "system:user"},
	{"/api/v1/sysUser", "system:user"},
	{"/api/v1/sysRoleMenu", "system:role"},
	{"/api/v1/sysRole", "system:role"},
	{"/api/v1/sysMenu", "system:menu"},
}

// RequirePermission checks the permission code mapped from method and route.
func RequirePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := permissionCode(c.Request.Method, c.FullPath())
		if code == "" {
			c.Next()
			return
		}
		raw, exists := c.Get("claims")
		claims, ok := raw.(*jwt.Claims)
		if !exists || !ok || claims == nil {
			c.JSON(http.StatusOK, gin.H{"code": 40001, "msg": i18n.T(c, "error.auth.unauthorized"), "data": struct{}{}})
			c.Abort()
			return
		}
		userID, err := strconv.ParseUint(claims.UID, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 40001, "msg": i18n.T(c, "error.auth.unauthorized"), "data": struct{}{}})
			c.Abort()
			return
		}
		allowed, err := UserHasPermission(userID, code)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 50000, "msg": i18n.T(c, "error.server.internal"), "data": struct{}{}})
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusOK, gin.H{"code": 40003, "msg": i18n.T(c, "error.auth.forbidden"), "data": struct{}{}})
			c.Abort()
			return
		}
		c.Next()
	}
}

func permissionCode(method, fullPath string) string {
	for _, item := range resources {
		if fullPath == item.prefix || strings.HasPrefix(fullPath, item.prefix+"/") {
			if strings.HasSuffix(fullPath, "/roles") || strings.HasSuffix(fullPath, "/menus") {
				if method == http.MethodGet {
					return item.perm + ":query"
				}
				return item.perm + ":edit"
			}
			if strings.HasSuffix(fullPath, "/list") || method == http.MethodGet {
				return item.perm + ":query"
			}
			switch method {
			case http.MethodPost:
				return item.perm + ":add"
			case http.MethodPut:
				return item.perm + ":edit"
			case http.MethodDelete:
				return item.perm + ":delete"
			}
		}
	}
	return ""
}

// UserHasPermission reports whether the user owns the permission code.
func UserHasPermission(userID uint64, code string) (bool, error) {
	db := database.GetDB()
	var roleCodes []string
	if err := db.Table("sys_user_role ur").
		Joins("JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 1", userID).
		Pluck("r.role_code", &roleCodes).Error; err != nil {
		return false, err
	}
	for _, roleCode := range roleCodes {
		if roleCode == "super_admin" {
			return true, nil
		}
	}
	var count int64
	err := db.Table("sys_user_role ur").
		Joins("JOIN sys_role_menu rm ON rm.role_id = ur.role_id").
		Joins("JOIN sys_menu m ON m.id = rm.menu_id").
		Where("ur.user_id = ? AND m.permission_code = ? AND m.status = 1", userID, code).
		Count(&count).Error
	return count > 0, err
}
