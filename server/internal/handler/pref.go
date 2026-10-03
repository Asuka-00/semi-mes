package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
)

type prefBody struct {
	Payload string `json:"payload"`
}

// GetPref returns the caller's saved list settings for one page.
func GetPref(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, ecode.Unauthorized)
		return
	}
	payload, err := dao.GetUserPref(database.GetDB(), userID, c.Param("page"))
	if err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Success(c, gin.H{"payload": payload})
}

// SavePref stores the caller's list settings for one page.
func SavePref(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Error(c, ecode.Unauthorized)
		return
	}
	body := prefBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.SaveUserPref(database.GetDB(), userID, c.Param("page"), body.Payload); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Success(c, gin.H{"page": c.Param("page")})
}
