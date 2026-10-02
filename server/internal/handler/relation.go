package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

type RelationHandler interface {
	GetUserRoles(c *gin.Context)
	SetUserRoles(c *gin.Context)
	GetRoleMenus(c *gin.Context)
	SetRoleMenus(c *gin.Context)
}

type relationHandler struct{}

func NewRelationHandler() RelationHandler {
	return &relationHandler{}
}

type idListRequest struct {
	RoleIDs []uint64 `json:"roleIds"`
	MenuIDs []uint64 `json:"menuIds"`
}

func (h *relationHandler) GetUserRoles(c *gin.Context) {
	userID, ok := pathID(c)
	if !ok {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	var ids []uint64
	if err := database.GetDB().Model(&model.SysUserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if ids == nil {
		ids = []uint64{}
	}
	response.Success(c, gin.H{"roleIds": ids})
}

func (h *relationHandler) SetUserRoles(c *gin.Context) {
	userID, ok := pathID(c)
	if !ok {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	var req idListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	db := database.GetDB()
	if err := db.Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if len(req.RoleIDs) > 0 {
		rows := make([]model.SysUserRole, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			rows = append(rows, model.SysUserRole{UserID: int(userID), RoleID: int(roleID)})
		}
		if err := db.Create(&rows).Error; err != nil {
			fail(c, 50000, "error.server.internal")
			return
		}
	}
	response.Success(c)
}

func (h *relationHandler) GetRoleMenus(c *gin.Context) {
	roleID, ok := pathID(c)
	if !ok {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	var ids []uint64
	if err := database.GetDB().Model(&model.SysRoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if ids == nil {
		ids = []uint64{}
	}
	response.Success(c, gin.H{"menuIds": ids})
}

func (h *relationHandler) SetRoleMenus(c *gin.Context) {
	roleID, ok := pathID(c)
	if !ok {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	var req idListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	db := database.GetDB()
	if err := db.Where("role_id = ?", roleID).Delete(&model.SysRoleMenu{}).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if len(req.MenuIDs) > 0 {
		rows := make([]model.SysRoleMenu, 0, len(req.MenuIDs))
		for _, menuID := range req.MenuIDs {
			rows = append(rows, model.SysRoleMenu{RoleID: int(roleID), MenuID: int(menuID)})
		}
		if err := db.Create(&rows).Error; err != nil {
			fail(c, 50000, "error.server.internal")
			return
		}
	}
	response.Success(c)
}

func pathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
