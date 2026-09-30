package handler

import (
	"semi-mes/server/internal/config"
	"semi-mes/server/internal/middleware"
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo interface{} `json:"user_info"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, 40001, "error.auth.invalid_credentials")
			return
		}
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	if user.Status != 1 {
		utils.Error(c, 40001, "error.auth.user_disabled")
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		utils.Error(c, 40001, "error.auth.invalid_credentials")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, config.GlobalConfig.JWT.Secret, config.GlobalConfig.JWT.ExpireHours)
	if err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, LoginResponse{
		Token: token,
		UserInfo: gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"real_name": user.RealName,
			"email":     user.Email,
			"phone":     user.Phone,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	utils.Success(c, nil)
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user model.User
	if err := h.db.Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	permissions, err := middleware.GetUserPermissions(userID.(int64))
	if err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.Success(c, gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"real_name":   user.RealName,
		"email":       user.Email,
		"phone":       user.Phone,
		"roles":       user.Roles,
		"permissions": permissions,
	})
}

func (h *AuthHandler) GetMenus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	menus, err := middleware.GetUserMenus(userID.(int64))
	if err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	menuTree := buildMenuTree(menus, 0)
	utils.Success(c, menuTree)
}

func buildMenuTree(menus []model.Menu, parentID int64) []model.Menu {
	var tree []model.Menu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			menu.Children = buildMenuTree(menus, menu.ID)
			tree = append(tree, menu)
		}
	}
	return tree
}
