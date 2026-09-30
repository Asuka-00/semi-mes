package middleware

import (
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitPermissionMiddleware(database *gorm.DB) {
	db = database
}

func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.Error(c, 40001, "error.auth.unauthorized")
			c.Abort()
			return
		}

		hasPermission, err := checkUserPermission(userID.(int64), permissionCode)
		if err != nil {
			utils.Error(c, 50000, "error.server.internal")
			c.Abort()
			return
		}

		if !hasPermission {
			utils.Error(c, 40003, "error.auth.insufficient_permission")
			c.Abort()
			return
		}

		c.Next()
	}
}

func checkUserPermission(userID int64, permissionCode string) (bool, error) {
	var count int64
	err := db.Table("sys_user_role ur").
		Joins("INNER JOIN sys_role_menu rm ON ur.role_id = rm.role_id").
		Joins("INNER JOIN sys_menu m ON rm.menu_id = m.id").
		Where("ur.user_id = ? AND m.permission_code = ? AND m.status = 1", userID, permissionCode).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetUserPermissions(userID int64) ([]string, error) {
	var permissions []string
	err := db.Table("sys_user_role ur").
		Select("DISTINCT m.permission_code").
		Joins("INNER JOIN sys_role_menu rm ON ur.role_id = rm.role_id").
		Joins("INNER JOIN sys_menu m ON rm.menu_id = m.id").
		Where("ur.user_id = ? AND m.status = 1 AND m.permission_code != ''", userID).
		Pluck("permission_code", &permissions).Error

	return permissions, err
}

func GetUserMenus(userID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := db.Table("sys_menu m").
		Select("DISTINCT m.*").
		Joins("INNER JOIN sys_role_menu rm ON m.id = rm.menu_id").
		Joins("INNER JOIN sys_user_role ur ON rm.role_id = ur.role_id").
		Where("ur.user_id = ? AND m.status = 1 AND m.menu_type IN (1, 2)", userID).
		Order("m.sort_order ASC").
		Find(&menus).Error

	return menus, err
}
