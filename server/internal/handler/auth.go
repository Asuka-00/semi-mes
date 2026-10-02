package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"semi-mes/server/internal/config"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/i18n"
	"semi-mes/server/internal/model"
)

type AuthHandler interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	UserInfo(c *gin.Context)
	UserRoutes(c *gin.Context)
	ConstantRoutes(c *gin.Context)
	RouteExist(c *gin.Context)
}

type authHandler struct{}

func NewAuthHandler() AuthHandler {
	return &authHandler{}
}

type loginRequest struct {
	UserName string `json:"userName"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type elegantRoute struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Component string         `json:"component"`
	Props     any            `json:"props,omitempty"`
	Meta      map[string]any `json:"meta"`
	Children  []elegantRoute `json:"children,omitempty"`
}

func (h *authHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		fail(c, 40009, "error.validation.invalid_params")
		return
	}
	username := req.UserName
	if username == "" {
		username = req.Username
	}
	var user model.SysUser
	err := database.GetDB().Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fail(c, 40001, "error.auth.invalid_credentials")
			return
		}
		fail(c, 50000, "error.server.internal")
		return
	}
	if user.Status != 1 {
		fail(c, 40001, "error.auth.user_disabled")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		fail(c, 40001, "error.auth.invalid_credentials")
		return
	}
	token, refresh, err := issueTokens(user.ID)
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	response.Success(c, gin.H{"token": token, "refreshToken": refresh})
}

func (h *authHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	claims, err := jwt.ValidateToken(req.RefreshToken, jwt.WithValidateTokenSignKey([]byte(signKey())))
	if err != nil {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	userID, err := strconv.ParseUint(claims.UID, 10, 64)
	if err != nil {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	token, refresh, err := issueTokens(userID)
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	response.Success(c, gin.H{"token": token, "refreshToken": refresh})
}

func (h *authHandler) UserInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	var user model.SysUser
	if err := database.GetDB().Where("id = ?", userID).First(&user).Error; err != nil {
		fail(c, 40004, "error.resource.not_found")
		return
	}
	var roles []string
	if err := database.GetDB().Table("sys_user_role ur").
		Joins("JOIN sys_role r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 1", userID).
		Pluck("r.role_code", &roles).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	var buttons []string
	if err := database.GetDB().Table("sys_menu m").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = m.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND m.status = 1 AND m.menu_type = 3 AND m.permission_code <> ''", userID).
		Distinct().
		Pluck("m.permission_code", &buttons).Error; err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if buttons == nil {
		buttons = []string{}
	}
	if roles == nil {
		roles = []string{}
	}
	response.Success(c, gin.H{
		"userId":   strconv.FormatUint(user.ID, 10),
		"userName": user.Username,
		"roles":    roles,
		"buttons":  buttons,
	})
}

func (h *authHandler) UserRoutes(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	var menus []model.SysMenu
	err := database.GetDB().Table("sys_menu m").
		Select("DISTINCT m.*").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = m.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND m.status = 1 AND m.menu_type IN (1, 2)", userID).
		Order("m.sort_order ASC, m.id ASC").
		Find(&menus).Error
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	routes := buildRoutes(menus, 0)
	if routes == nil {
		routes = []elegantRoute{}
	}
	response.Success(c, gin.H{"routes": routes, "home": "home"})
}

func (h *authHandler) ConstantRoutes(c *gin.Context) {
	response.Success(c, []elegantRoute{
		{Name: "403", Path: "/403", Component: "layout.blank$view.403", Meta: constantMeta("403", "route.403")},
		{Name: "404", Path: "/404", Component: "layout.blank$view.404", Meta: constantMeta("404", "route.404")},
		{Name: "500", Path: "/500", Component: "layout.blank$view.500", Meta: constantMeta("500", "route.500")},
		{
			Name: "iframe-page", Path: "/iframe-page/:url", Component: "layout.base$view.iframe-page", Props: true,
			Meta: map[string]any{"title": "iframe-page", "i18nKey": "route.iframe-page", "constant": true, "hideInMenu": true, "keepAlive": true},
		},
		{
			Name: "login", Path: "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			Component: "layout.blank$view.login", Props: true, Meta: constantMeta("login", "route.login"),
		},
	})
}

func (h *authHandler) RouteExist(c *gin.Context) {
	name := c.Query("routeName")
	userID, ok := currentUserID(c)
	if !ok {
		response.Success(c, false)
		return
	}
	var count int64
	_ = database.GetDB().Table("sys_menu m").
		Joins("JOIN sys_role_menu rm ON rm.menu_id = m.id").
		Joins("JOIN sys_user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND m.route_name = ? AND m.status = 1", userID, name).
		Count(&count).Error
	response.Success(c, count > 0)
}

func constantMeta(title, key string) map[string]any {
	return map[string]any{"title": title, "i18nKey": key, "constant": true, "hideInMenu": true}
}

func buildRoutes(menus []model.SysMenu, parentID int) []elegantRoute {
	var routes []elegantRoute
	for _, menu := range menus {
		if menu.ParentID != parentID || menu.RouteName == "" {
			continue
		}
		node := elegantRoute{
			Name:      menu.RouteName,
			Path:      menu.RoutePath,
			Component: menu.ComponentPath,
			Meta: map[string]any{
				"title":   menu.RouteName,
				"i18nKey": "route." + menu.RouteName,
				"icon":    menu.Icon,
				"order":   menu.SortOrder,
			},
		}
		if menu.RouteName == "lot_detail" || menu.RouteName == "equipment_detail" {
			node.Meta["hideInMenu"] = true
		}
		children := buildRoutes(menus, int(menu.ID))
		if len(children) > 0 {
			node.Children = children
		}
		if menu.MenuType == 1 && len(children) == 0 {
			continue
		}
		routes = append(routes, node)
	}
	return routes
}

func issueTokens(userID uint64) (string, string, error) {
	tokens, err := jwt.GenerateTwoTokens(strconv.FormatUint(userID, 10), jwt.WithGenerateTwoTokensSignKey([]byte(signKey())))
	if err != nil {
		return "", "", err
	}
	return tokens.AccessToken, tokens.RefreshToken, nil
}

func signKey() string {
	key := config.Get().Jwt.SignKey
	if key == "" {
		return "mes-jwt-secret-key-change-in-production"
	}
	return key
}

func currentUserID(c *gin.Context) (uint64, bool) {
	raw, ok := c.Get("claims")
	if !ok {
		return 0, false
	}
	claims, ok := raw.(*jwt.Claims)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseUint(claims.UID, 10, 64)
	return id, err == nil
}

func fail(c *gin.Context, code int, key string) {
	c.JSON(200, gin.H{"code": code, "msg": i18n.T(c, key), "data": struct{}{}})
}
