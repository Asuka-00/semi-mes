package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/cache"
	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
	"semi-mes/server/internal/types"
)

var _ SysUserRoleHandler = (*sysUserRoleHandler)(nil)

// SysUserRoleHandler defining the handler interface
type SysUserRoleHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
}

type sysUserRoleHandler struct {
	iDao dao.SysUserRoleDao
}

// NewSysUserRoleHandler creating the handler interface
func NewSysUserRoleHandler() SysUserRoleHandler {
	return &sysUserRoleHandler{
		iDao: dao.NewSysUserRoleDao(
			database.GetDB(), // db driver is sqlite
			cache.NewSysUserRoleCache(database.GetCacheType()),
		),
	}
}

// Create a new sysUserRole
// @Summary Create a new sysUserRole
// @Description Creates a new sysUserRole entity using the provided data in the request body.
// @Tags sysUserRole
// @Accept json
// @Produce json
// @Param data body types.CreateSysUserRoleRequest true "sysUserRole information"
// @Success 200 {object} types.CreateSysUserRoleReply{}
// @Router /api/v1/sysUserRole [post]
// @Security BearerAuth
func (h *sysUserRoleHandler) Create(c *gin.Context) {
	form := &types.CreateSysUserRoleRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	sysUserRole := &model.SysUserRole{}
	err = copier.Copy(sysUserRole, form)
	if err != nil {
		response.Error(c, ecode.ErrCreateSysUserRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, sysUserRole)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c, gin.H{"id": sysUserRole.ID})
}

// DeleteByID delete a sysUserRole by id
// @Summary Delete a sysUserRole by id
// @Description Deletes a existing sysUserRole identified by the given id in the path.
// @Tags sysUserRole
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteSysUserRoleByIDReply{}
// @Router /api/v1/sysUserRole/{id} [delete]
// @Security BearerAuth
func (h *sysUserRoleHandler) DeleteByID(c *gin.Context) {
	_, id, isAbort := getSysUserRoleIDFromPath(c)
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

// UpdateByID update a sysUserRole by id
// @Summary Update a sysUserRole by id
// @Description Updates the specified sysUserRole by given id in the path, support partial update.
// @Tags sysUserRole
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateSysUserRoleByIDRequest true "sysUserRole information"
// @Success 200 {object} types.UpdateSysUserRoleByIDReply{}
// @Router /api/v1/sysUserRole/{id} [put]
// @Security BearerAuth
func (h *sysUserRoleHandler) UpdateByID(c *gin.Context) {
	_, id, isAbort := getSysUserRoleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	form := &types.UpdateSysUserRoleByIDRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}
	form.ID = id

	sysUserRole := &model.SysUserRole{}
	err = copier.Copy(sysUserRole, form)
	if err != nil {
		response.Error(c, ecode.ErrUpdateByIDSysUserRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	ctx := middleware.WrapCtx(c)
	err = h.iDao.UpdateByID(ctx, sysUserRole)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a sysUserRole by id
// @Summary Get a sysUserRole by id
// @Description Gets detailed information of a sysUserRole specified by the given id in the path.
// @Tags sysUserRole
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetSysUserRoleByIDReply{}
// @Router /api/v1/sysUserRole/{id} [get]
// @Security BearerAuth
func (h *sysUserRoleHandler) GetByID(c *gin.Context) {
	_, id, isAbort := getSysUserRoleIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysUserRole, err := h.iDao.GetByID(ctx, id)
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

	data := &types.SysUserRoleObjDetail{}
	err = copier.Copy(data, sysUserRole)
	if err != nil {
		response.Error(c, ecode.ErrGetByIDSysUserRole)
		return
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	response.Success(c, gin.H{"sysUserRole": data})
}

// List get a paginated list of sysUserRoles by custom conditions
// @Summary Get a paginated list of sysUserRoles by custom conditions
// @Description Returns a paginated list of sysUserRole based on query filters, including page number and size.
// @Tags sysUserRole
// @Accept json
// @Produce json
// @Param data body types.Params true "query parameters"
// @Success 200 {object} types.ListSysUserRolesReply{}
// @Router /api/v1/sysUserRole/list [post]
// @Security BearerAuth
func (h *sysUserRoleHandler) List(c *gin.Context) {
	form := &types.ListSysUserRolesRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	sysUserRoles, total, err := h.iDao.GetByColumns(ctx, &form.Params)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	data, err := convertSysUserRoles(sysUserRoles)
	if err != nil {
		response.Error(c, ecode.ErrListSysUserRole)
		return
	}

	response.Success(c, gin.H{
		"sysUserRoles": data,
		"total":        total,
	})
}

func getSysUserRoleIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
}

func convertSysUserRole(sysUserRole *model.SysUserRole) (*types.SysUserRoleObjDetail, error) {
	data := &types.SysUserRoleObjDetail{}
	err := copier.Copy(data, sysUserRole)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return data, nil
}

func convertSysUserRoles(fromValues []*model.SysUserRole) ([]*types.SysUserRoleObjDetail, error) {
	toValues := []*types.SysUserRoleObjDetail{}
	for _, v := range fromValues {
		data, err := convertSysUserRole(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}
