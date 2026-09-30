package handler

import (
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MenuHandler struct {
	db *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{db: db}
}

type MenuRequest struct {
	ParentID       int64  `json:"parent_id"`
	MenuType       int8   `json:"menu_type" binding:"required"`
	MenuName       string `json:"menu_name" binding:"required"`
	PermissionCode string `json:"permission_code"`
	RoutePath      string `json:"route_path"`
	ComponentPath  string `json:"component_path"`
	Icon           string `json:"icon"`
	SortOrder      int    `json:"sort_order"`
	Status         int8   `json:"status"`
}

func (h *MenuHandler) List(c *gin.Context) {
	var menus []model.Menu
	if err := h.db.Order("sort_order ASC").Find(&menus).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	menuTree := buildMenuTree(menus, 0)
	utils.Success(c, menuTree)
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	menu := model.Menu{
		ParentID:       req.ParentID,
		MenuType:       req.MenuType,
		MenuName:       req.MenuName,
		PermissionCode: req.PermissionCode,
		RoutePath:      req.RoutePath,
		ComponentPath:  req.ComponentPath,
		Icon:           req.Icon,
		SortOrder:      req.SortOrder,
		Status:         req.Status,
	}

	if err := h.db.Create(&menu).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, menu)
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var menu model.Menu
	if err := h.db.Where("id = ?", id).First(&menu).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"parent_id":       req.ParentID,
		"menu_type":       req.MenuType,
		"menu_name":       req.MenuName,
		"permission_code": req.PermissionCode,
		"route_path":      req.RoutePath,
		"component_path":  req.ComponentPath,
		"icon":            req.Icon,
		"sort_order":      req.SortOrder,
		"status":          req.Status,
	}

	if err := h.db.Model(&menu).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, nil)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var count int64
	h.db.Model(&model.Menu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 40009, "error.menu.has_children")
		return
	}

	if err := h.db.Delete(&model.Menu{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	h.db.Where("menu_id = ?", id).Delete(&model.RoleMenu{})

	utils.Success(c, nil)
}
