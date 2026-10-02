package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"golang.org/x/crypto/bcrypt"

	"semi-mes/server/internal/cache"
	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
	"semi-mes/server/internal/types"
)

var _ SysUserHandler = (*sysUserHandler)(nil)

// SysUserHandler defining the handler interface
type SysUserHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type sysUserHandler struct {
	iDao dao.SysUserDao
}

// NewSysUserHandler creating the handler interface
func NewSysUserHandler() SysUserHandler {
	return &sysUserHandler{
		iDao: dao.NewSysUserDao(
			database.GetDB(), // db driver is sqlite
			cache.NewSysUserCache(database.GetCacheType()),
		),
	}
}

// Create a new sysUser
// @Summary Create a new sysUser
// @Description Creates a new sysUser entity using the provided data in the request body.
// @Tags sysUser
// @Accept json
// @Produce json
// @Param data body types.CreateSysUserRequest true "sysUser information"
// @Success 200 {object} types.CreateSysUserReply{}
// @Router /api/v1/sysUser [post]
// @Security BearerAuth
func (h *sysUserHandler) Create(c *gin.Context) {
	form := &types.CreateSysUserRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	sysUser := &model.SysUser{}
	err = copier.Copy(sysUser, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateSysUser)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, ecode.InternalServerError)
		return
	}
	sysUser.Password = string(hashed)

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, sysUser)
	if err != nil {
		if errors.Is(err, dao.ErrDuplicate) {
			response.Error(c, ecode.Conflict)
			return
		}
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": sysUser.ID})
}

// DeleteByID delete a sysUser by id
// @Summary Delete a sysUser by id
// @Description Deletes a existing sysUser identified by the given id in the path.
// @Tags sysUser
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteSysUserByIDReply{}
// @Router /api/v1/sysUser/{id} [delete]
// @Security BearerAuth
func (h *sysUserHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getSysUserIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.DeleteByID(ctx, id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// UpdateByID update a sysUser by id
// @Summary Update a sysUser by id
// @Description Updates the specified sysUser by given id in the path, support partial update.
// @Tags sysUser
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateSysUserByIDRequest true "sysUser information"
// @Success 200 {object} types.UpdateSysUserByIDReply{}
// @Router /api/v1/sysUser/{id} [put]
// @Security BearerAuth
func (h *sysUserHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getSysUserIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateSysUserByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	sysUser := &model.SysUser{}
	err = copier.Copy(sysUser, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDSysUser)
		return
	}
	if form.Password != "" {
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			response.Error(c, ecode.InternalServerError)
			return
		}
		sysUser.Password = string(hashed)
	} else {
		sysUser.Password = ""
	}

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, sysUser)
	if err != nil {
		if errors.Is(err, dao.ErrDuplicate) {
			response.Error(c, ecode.Conflict)
			return
		}
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a sysUser by id
// @Summary Get a sysUser by id
// @Description Gets detailed information of a sysUser specified by the given id in the path.
// @Tags sysUser
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetSysUserByIDReply{}
// @Router /api/v1/sysUser/{id} [get]
// @Security BearerAuth
func (h *sysUserHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getSysUserIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysUser, err := h.iDao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByID not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Error(c, ecode.NotFound)
		} else {
			logger.Error("GetByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
			response.Output(c, ecode.InternalServerError.ToHTTPCode())
		}
		return
	}

	data := &types.SysUserObjDetail{}
	err = copier.Copy(data, sysUser)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDSysUser)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"sysUser": data})
}

// List get a paginated list of sysUsers by custom conditions
// @Summary Get a paginated list of sysUsers by custom conditions
// @Description Returns a paginated list of sysUser based on query filters, including page number and size.
// @Tags sysUser
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListSysUsersReply{}
// @Router /api/v1/sysUser/list [post]
// @Security BearerAuth
func (h *sysUserHandler) List(c *gin.Context) {
	form := &types.ListSysUsersRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysUsers, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertSysUsers(sysUsers)
	if err != nil {
		response.Error(c, ecode.ErrListSysUser)
		return
	}

	response.Success(c, gin.H{
		"sysUsers": data,
		"total":    total,
	})
}

func getSysUserIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertSysUser(sysUser *model.SysUser) (*types.SysUserObjDetail, error) {
	data := &types.SysUserObjDetail{}
	err := copier.Copy(data, sysUser)
	if err != nil {
		return nil, err
	}
	data.Password = ""

	return data, nil
}

func convertSysUsers(fromValues []*model.SysUser) ([]*types.SysUserObjDetail, error) {
	toValues := []*types.SysUserObjDetail{}
	for _, v := range fromValues {
		data, err := convertSysUser(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
