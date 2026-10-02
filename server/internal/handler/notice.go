package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/config"
	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
)

// Features reports runtime feature flags. SPC is off unless configs/mes.yml enables it.
func Features(c *gin.Context) {
	response.Success(c, gin.H{"spc": config.SPCEnabled()})
}

// ListNotices returns the current user's notices.
func ListNotices(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	rows, unread, err := dao.ListNotices(database.GetDB(), userID, limit)
	if err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	if rows == nil {
		rows = []dao.NoticeView{}
	}
	response.Success(c, gin.H{"notices": rows, "unread": unread})
}

// MarkNoticeRead marks one notice read.
func MarkNoticeRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err = dao.MarkNoticeRead(database.GetDB(), userID, id); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c)
}

// MarkAllNoticesRead marks every notice read for the current user.
func MarkAllNoticesRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	if err := dao.MarkAllNoticesRead(database.GetDB(), userID); err != nil {
		fail(c, 50000, "error.server.internal")
		return
	}
	response.Success(c)
}
