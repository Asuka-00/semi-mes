package handler

import (
	"semi-mes/server/internal/model"
	"semi-mes/server/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type UserRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password"`
	RealName string  `json:"real_name"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Status   int8    `json:"status"`
	RoleIDs  []int64 `json:"role_ids"`
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := h.db.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (page - 1) * pageSize
	if err := query.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	utils.SuccessPage(c, users, total, page, pageSize)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var exists model.User
	if err := h.db.Where("username = ?", req.Username).First(&exists).Error; err == nil {
		utils.Error(c, 40009, "error.user.username_exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	user := model.User{
		Username: req.Username,
		Password: hashedPassword,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   req.Status,
	}

	if err := h.db.Create(&user).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	if len(req.RoleIDs) > 0 {
		var userRoles []model.UserRole
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, model.UserRole{
				UserID: user.ID,
				RoleID: roleID,
			})
		}
		h.db.Create(&userRoles)
	}

	utils.Success(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 40009, "error.validation.invalid_params")
		return
	}

	var user model.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		utils.Error(c, 40004, "error.resource.not_found")
		return
	}

	updates := map[string]interface{}{
		"real_name": req.RealName,
		"email":     req.Email,
		"phone":     req.Phone,
		"status":    req.Status,
	}

	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Error(c, 50000, "error.server.internal")
			return
		}
		updates["password"] = hashedPassword
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	h.db.Where("user_id = ?", id).Delete(&model.UserRole{})
	if len(req.RoleIDs) > 0 {
		var userRoles []model.UserRole
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, model.UserRole{
				UserID: id,
				RoleID: roleID,
			})
		}
		h.db.Create(&userRoles)
	}

	utils.Success(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.db.Delete(&model.User{}, id).Error; err != nil {
		utils.Error(c, 50000, "error.server.internal")
		return
	}

	h.db.Where("user_id = ?", id).Delete(&model.UserRole{})

	utils.Success(c, nil)
}
