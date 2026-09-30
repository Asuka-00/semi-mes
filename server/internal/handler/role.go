package handler

import (
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

type RoleRequest struct {
	RoleCode    string  `json:"role_code" binding:"required"`
	RoleName    string  `json:"role_name" binding:"required"`
	Description string  `json:"description"`
	Status      int8    `json:"status"`
	MenuIDs     []int64 `json:"menu_ids"`
}

func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := h.db.Model(&model.Role{})
	if keyword != "" {
		query = query.Where("role_code LIKE ? OR role_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var roles []model.Role
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, roles, total, page, pageSize)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var exists model.Role
	if err := h.db.Where("role_code = ?", req.RoleCode).First(&exists).Error; err == nil {
		utils.Error(c, 40009, "error.role.code_exists")
		return
	}

	role := model.Role{
		RoleCode:    req.RoleCode,
		RoleName:    req.RoleName,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&role).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	if len(req.MenuIDs) > 0 {
		var roleMenus []model.RoleMenu
		for _, menuID := range req.MenuIDs {
			roleMenus = append(roleMenus, model.RoleMenu{
				RoleID: role.ID,
				MenuID: menuID,
			})
		}
		h.db.Create(&roleMenus)
	}

	utils.Success(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var role model.Role
	if err := h.db.Where("id = ?", id).First(&role).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"role_name":   req.RoleName,
		"description": req.Description,
		"status":      req.Status,
	}

	if err := h.db.Model(&role).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	h.db.Where("role_id = ?", id).Delete(&model.RoleMenu{})
	if len(req.MenuIDs) > 0 {
		var roleMenus []model.RoleMenu
		for _, menuID := range req.MenuIDs {
			roleMenus = append(roleMenus, model.RoleMenu{
				RoleID: id,
				MenuID: menuID,
			})
		}
		h.db.Create(&roleMenus)
	}

	utils.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.Role{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	h.db.Where("role_id = ?", id).Delete(&model.RoleMenu{})
	h.db.Where("role_id = ?", id).Delete(&model.UserRole{})

	utils.Success(c, nil)
}

func (h *RoleHandler) GetMenus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var menuIDs []int64
	h.db.Model(&model.RoleMenu{}).Where("role_id = ?", id).Pluck("menu_id", &menuIDs)

	utils.Success(c, menuIDs)
}
