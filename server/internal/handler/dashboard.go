package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/rbac"
)

// ShopDashboard returns the home-page aggregates the caller is allowed to see.
func ShopDashboard(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	wip, err := allowed(userID, "wip:move:query", "lot:lot:query")
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	equipment, err := allowed(userID, "eqp:equipment:query")
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	view, err := dao.ShopDashboard(database.GetDB(), userID, wip, equipment)
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	response.Success(c, view)
}

func allowed(userID uint64, codes ...string) (bool, error) {
	for _, code := range codes {
		ok, err := rbac.UserHasPermission(userID, code)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
